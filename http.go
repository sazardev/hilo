package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClientConfig holds configurable options for the HTTP client.
type HTTPClientConfig struct {
	Timeout         time.Duration
	FollowRedirects bool
	VerifyTLS       bool
	ProxyURL        string
}

// DefaultHTTPClientConfig returns sensible defaults.
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:         30 * time.Second,
		FollowRedirects: true,
		VerifyTLS:       true,
	}
}

// ExecuteRequest sends an HTTP request and returns the captured response.
// The env map is used for {{VAR}} substitution in URL, headers, body, and auth.
func ExecuteRequest(req Request, env map[string]string, cfg HTTPClientConfig) Response {
	start := time.Now()

	// Build the HTTP request
	httpReq, err := buildHTTPRequest(req, env)
	if err != nil {
		return errorResponse(req.ID, err, start)
	}

	// Build the client
	client := buildHTTPClient(cfg)

	// Execute
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return errorResponse(req.ID, err, start)
	}
	defer httpResp.Body.Close()

	// Read body (cap at 10MB to avoid OOM)
	bodyBytes, err := io.ReadAll(io.LimitReader(httpResp.Body, 10<<20))
	if err != nil {
		return errorResponse(req.ID, fmt.Errorf("read body: %w", err), start)
	}

	// Capture response headers
	respHeaders := make(map[string]string)
	for k, v := range httpResp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	duration := time.Since(start)

	return Response{
		RequestID:  req.ID,
		StatusCode: httpResp.StatusCode,
		StatusText: httpResp.Status,
		Headers:    respHeaders,
		Body:       string(bodyBytes),
		BodySize:   int64(len(bodyBytes)),
		Duration:   duration,
		Timestamp:  time.Now(),
	}
}

// buildHTTPRequest constructs an *http.Request from our Request model.
func buildHTTPRequest(req Request, env map[string]string) (*http.Request, error) {
	// Resolve variables in URL
	resolvedURL := ResolveVars(req.URL, env)

	// Append query params
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(resolvedURL)
		if err != nil {
			return nil, fmt.Errorf("parse URL: %w", err)
		}
		q := parsedURL.Query()
		for k, v := range req.QueryParams {
			q.Set(ResolveVars(k, env), ResolveVars(v, env))
		}
		parsedURL.RawQuery = q.Encode()
		resolvedURL = parsedURL.String()
	}

	// Build body
	var bodyReader io.Reader
	if req.Body != "" && req.BodyType != BodyNone {
		resolvedBody := ResolveVars(req.Body, env)
		bodyReader = strings.NewReader(resolvedBody)
	}

	httpReq, err := http.NewRequest(req.Method, resolvedURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	for k, v := range req.Headers {
		httpReq.Header.Set(ResolveVars(k, env), ResolveVars(v, env))
	}

	// Set Content-Type based on BodyType
	if req.BodyType != BodyNone && req.Body != "" {
		if httpReq.Header.Get("Content-Type") == "" {
			httpReq.Header.Set("Content-Type", bodyContentType(req.BodyType))
		}
	}

	// Apply auth
	if req.Auth != nil && req.Auth.Type != AuthNone {
		applyAuth(httpReq, req.Auth, env)
	}

	return httpReq, nil
}

// buildHTTPClient constructs an *http.Client from config.
func buildHTTPClient(cfg HTTPClientConfig) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifyTLS,
		},
	}

	// Configure proxy if provided
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	if !cfg.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

// applyAuth sets the appropriate authorization header or query param.
func applyAuth(req *http.Request, auth *Auth, env map[string]string) {
	switch auth.Type {
	case AuthBearer:
		token := ResolveVars(auth.Value, env)
		req.Header.Set("Authorization", "Bearer "+token)

	case AuthBasic:
		user := ResolveVars(auth.Username, env)
		pass := ResolveVars(auth.Password, env)
		req.SetBasicAuth(user, pass)

	case AuthAPIKey:
		key := ResolveVars(auth.Key, env)
		value := ResolveVars(auth.Value, env)
		// If key looks like a header name (contains no URL special chars), use header
		// Otherwise treat as query param
		if strings.Contains(key, "X-") || strings.Contains(key, "Authorization") || strings.Contains(key, "Api") {
			req.Header.Set(key, value)
		} else {
			q := req.URL.Query()
			q.Set(key, value)
			req.URL.RawQuery = q.Encode()
		}

	case AuthDigest:
		// Digest auth requires challenge-response; use basic for now
		// Full digest implementation would require a round-trip to get the nonce
		user := ResolveVars(auth.Username, env)
		pass := ResolveVars(auth.Password, env)
		req.SetBasicAuth(user, pass)
	}
}

// bodyContentType returns the appropriate Content-Type header for a body type.
func bodyContentType(bodyType string) string {
	switch bodyType {
	case BodyJSON:
		return "application/json"
	case BodyXML:
		return "application/xml"
	case BodyForm:
		return "application/x-www-form-urlencoded"
	case BodyText:
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// errorResponse creates a Response with an error message.
func errorResponse(requestID string, err error, start time.Time) Response {
	return Response{
		RequestID:  requestID,
		StatusCode: 0,
		StatusText: "Error",
		Headers:    make(map[string]string),
		Body:       "",
		BodySize:   0,
		Duration:   time.Since(start),
		Timestamp:  time.Now(),
		Error:      err.Error(),
	}
}

// GenerateCurl produces an equivalent cURL command string for a request.
func GenerateCurl(req Request, env map[string]string) string {
	var parts []string
	parts = append(parts, "curl")

	// Method
	if req.Method != MethodGet {
		parts = append(parts, "-X", req.Method)
	}

	// URL
	resolvedURL := ResolveVars(req.URL, env)
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(resolvedURL)
		if err == nil {
			q := parsedURL.Query()
			for k, v := range req.QueryParams {
				q.Set(ResolveVars(k, env), ResolveVars(v, env))
			}
			parsedURL.RawQuery = q.Encode()
			resolvedURL = parsedURL.String()
		}
	}
	parts = append(parts, fmt.Sprintf("%q", resolvedURL))

	// Headers
	for k, v := range req.Headers {
		resolvedK := ResolveVars(k, env)
		resolvedV := ResolveVars(v, env)
		parts = append(parts, "-H", fmt.Sprintf("%q", resolvedK+": "+resolvedV))
	}

	// Auth
	if req.Auth != nil && req.Auth.Type != AuthNone {
		switch req.Auth.Type {
		case AuthBearer:
			token := ResolveVars(req.Auth.Value, env)
			parts = append(parts, "-H", fmt.Sprintf("%q", "Authorization: Bearer "+token))
		case AuthBasic:
			user := ResolveVars(req.Auth.Username, env)
			pass := ResolveVars(req.Auth.Password, env)
			parts = append(parts, "-u", fmt.Sprintf("%q", user+":"+pass))
		case AuthAPIKey:
			key := ResolveVars(req.Auth.Key, env)
			value := ResolveVars(req.Auth.Value, env)
			parts = append(parts, "-H", fmt.Sprintf("%q", key+": "+value))
		}
	}

	// Body
	if req.Body != "" && req.BodyType != BodyNone {
		resolvedBody := ResolveVars(req.Body, env)
		parts = append(parts, "-d", fmt.Sprintf("%q", resolvedBody))
	}

	return strings.Join(parts, " ")
}

// FormatBodySize returns a human-readable size string.
func FormatBodySize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// StatusColor returns a color name based on the HTTP status code.
func StatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "green"
	case code >= 300 && code < 400:
		return "cyan"
	case code >= 400 && code < 500:
		return "yellow"
	case code >= 500:
		return "red"
	default:
		return "white"
	}
}

// PrettyJSON attempts to pretty-print JSON content. Returns original if invalid.
func PrettyJSON(s string) string {
	var buf bytes.Buffer
	if json.Indent(&buf, []byte(s), "", "  ") == nil {
		return buf.String()
	}
	return s
}

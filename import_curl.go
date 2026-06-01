package main

import (
	"fmt"
	"strings"
)

// ParseCurl parses a curl command line into a Request. It understands the
// common flags (-X, -H, -d/--data*, -u, -b, -A, --url) and infers the method
// and body type, lifting Authorization: Bearer headers into structured auth.
func ParseCurl(input string) (Request, error) {
	tokens := tokenizeShell(input)
	if len(tokens) > 0 && tokens[0] == "curl" {
		tokens = tokens[1:]
	}

	req := NewRequest("", "", "")
	headers := map[string]string{}
	var method, url string
	var dataParts []string

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		next := func() string {
			if i+1 < len(tokens) {
				i++
				return tokens[i]
			}
			return ""
		}

		switch t {
		case "-X", "--request":
			method = strings.ToUpper(next())
		case "-H", "--header":
			if k, v, ok := splitHeader(next()); ok {
				headers[k] = v
			}
		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii", "--data-urlencode":
			dataParts = append(dataParts, next())
		case "-u", "--user":
			user, pass, _ := strings.Cut(next(), ":")
			req.Auth = &Auth{Type: AuthBasic, Username: user, Password: pass}
		case "-b", "--cookie":
			headers["Cookie"] = next()
		case "-A", "--user-agent":
			headers["User-Agent"] = next()
		case "-e", "--referer":
			headers["Referer"] = next()
		case "--url":
			url = next()
		case "--compressed", "-L", "--location", "-k", "--insecure", "-s", "--silent",
			"-i", "--include", "-v", "--verbose", "-g", "--globoff", "-f", "--fail":
			// valueless flags — ignore
		default:
			if strings.HasPrefix(t, "-") {
				continue // unknown flag, skip
			}
			if url == "" {
				url = t
			}
		}
	}

	if url == "" {
		return Request{}, fmt.Errorf("no URL found in curl command")
	}

	body := strings.Join(dataParts, "&")
	if method == "" {
		if body != "" {
			method = MethodPost
		} else {
			method = MethodGet
		}
	}

	// Lift a bearer token out of the Authorization header into structured auth.
	if a := headers["Authorization"]; a != "" && strings.HasPrefix(strings.ToLower(a), "bearer ") {
		req.Auth = &Auth{Type: AuthBearer, Value: strings.TrimSpace(a[len("bearer "):])}
		delete(headers, "Authorization")
	}

	req.Method = method
	req.URL = url
	req.Headers = headers
	if body != "" {
		req.Body = body
		if strings.Contains(strings.ToLower(headers["Content-Type"]), "json") || looksJSON(body) {
			req.BodyType = BodyJSON
		} else {
			req.BodyType = BodyText
		}
	}
	req.Name = fmt.Sprintf("%s %s", req.Method, shortURL(url))
	return req, nil
}

// tokenizeShell splits a shell-style command respecting single/double quotes,
// backslash escapes and line continuations.
func tokenizeShell(s string) []string {
	s = strings.ReplaceAll(s, "\\\r\n", " ")
	s = strings.ReplaceAll(s, "\\\n", " ")

	var tokens []string
	var buf strings.Builder
	inSingle, inDouble := false, false

	flush := func() {
		if buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				buf.WriteByte(c)
			}
		case inDouble:
			if c == '"' {
				inDouble = false
			} else if c == '\\' && i+1 < len(s) {
				i++
				buf.WriteByte(s[i])
			} else {
				buf.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == '\\' && i+1 < len(s):
			i++
			buf.WriteByte(s[i])
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			flush()
		default:
			buf.WriteByte(c)
		}
	}
	flush()
	return tokens
}

func splitHeader(h string) (key, value string, ok bool) {
	k, v, found := strings.Cut(h, ":")
	if !found {
		return "", "", false
	}
	return strings.TrimSpace(k), strings.TrimSpace(v), true
}

func looksJSON(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}

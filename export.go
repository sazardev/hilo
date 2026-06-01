package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getExportsDir() string { return filepath.Join(getBaseDir(), "exports") }

// SaveResponseToFile writes a response body to a timestamped file under the
// exports directory and returns the absolute path.
func SaveResponseToFile(resp *Response) (string, error) {
	if resp == nil {
		return "", os.ErrInvalid
	}
	dir := getExportsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	ext := "txt"
	if looksJSON(resp.Body) {
		ext = "json"
	}
	ts := resp.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	path := filepath.Join(dir, "response-"+ts.Format("2006-01-02T15-04-05")+"."+ext)
	if err := os.WriteFile(path, []byte(resp.Body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ExportRequestFile writes a request as a portable .hilo.json file.
func ExportRequestFile(req Request) (string, error) {
	dir := getExportsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", err
	}
	base := sanitizeFilename(req.Method + "-" + shortURL(req.URL))
	if base == "" {
		base = "request"
	}
	path := filepath.Join(dir, base+".hilo.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == '.' || r == '/' || r == ':' || r == ' ':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

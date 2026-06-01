package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestParseCurl(t *testing.T) {
	in := `curl -X POST 'https://api.example.com/users?page=2' ` +
		`-H 'Content-Type: application/json' -H 'Authorization: Bearer abc123' ` +
		`-d '{"name":"Jo"}'`
	req, err := ParseCurl(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("method = %q, want POST", req.Method)
	}
	if req.URL != "https://api.example.com/users?page=2" {
		t.Errorf("url = %q", req.URL)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("content-type header missing: %v", req.Headers)
	}
	if req.Auth == nil || req.Auth.Type != AuthBearer || req.Auth.Value != "abc123" {
		t.Errorf("bearer not lifted: %+v", req.Auth)
	}
	if req.Body != `{"name":"Jo"}` || req.BodyType != BodyJSON {
		t.Errorf("body = %q type=%q", req.Body, req.BodyType)
	}
}

func TestParseCurlInfersAndBasicAuth(t *testing.T) {
	// No -X but has data → POST; -u → basic auth; bare URL.
	req, err := ParseCurl(`curl https://x.test/login -u alice:secret -d field=1`)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != "POST" {
		t.Errorf("expected inferred POST, got %s", req.Method)
	}
	if req.Auth == nil || req.Auth.Type != AuthBasic || req.Auth.Username != "alice" || req.Auth.Password != "secret" {
		t.Errorf("basic auth not parsed: %+v", req.Auth)
	}
}

func TestGenerateSnippetsAllLanguages(t *testing.T) {
	req := NewRequest("t", "POST", "https://api.example.com/v1/items")
	req.Headers = map[string]string{"Accept": "application/json"}
	req.Body = `{"x":1}`
	req.BodyType = BodyJSON

	for i, lang := range snippetLangs {
		code := GenerateSnippet(req, map[string]string{}, i)
		if !strings.Contains(code, "api.example.com") {
			t.Errorf("%s snippet missing URL:\n%s", lang, code)
		}
	}
	if !strings.Contains(GenerateSnippet(req, nil, 1), "import requests") {
		t.Error("python snippet missing requests import")
	}
	if !strings.Contains(GenerateSnippet(req, nil, 3), "net/http") {
		t.Error("go snippet missing net/http import")
	}
}

func TestImportPostmanCollection(t *testing.T) {
	data := []byte(`{
		"info": {"name": "Demo API"},
		"item": [
			{"name": "list", "request": {"method": "GET", "url": "https://demo.test/list",
				"header": [{"key": "Accept", "value": "application/json"}]}},
			{"name": "folder", "item": [
				{"name": "create", "request": {"method": "POST",
					"url": {"raw": "https://demo.test/create"},
					"body": {"mode": "raw", "raw": "{\"a\":1}"}}}
			]}
		]
	}`)
	col, reqs, err := ImportPostmanCollection(data)
	if err != nil {
		t.Fatal(err)
	}
	if col.Name != "Demo API" {
		t.Errorf("collection name = %q", col.Name)
	}
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requests (incl nested), got %d", len(reqs))
	}
	if reqs[1].Method != "POST" || reqs[1].URL != "https://demo.test/create" {
		t.Errorf("nested request parsed wrong: %+v", reqs[1])
	}
}

func TestOAuth2ClientCredentialsFlow(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("grant_type") != "client_credentials" || r.FormValue("client_id") != "id" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"TOK123","token_type":"Bearer"}`)
	}))
	defer tokenSrv.Close()

	var gotAuth string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, "ok")
	}))
	defer apiSrv.Close()

	req := NewRequest("t", "GET", apiSrv.URL)
	req.Auth = &Auth{Type: AuthOAuth2, TokenURL: tokenSrv.URL, ClientID: "id", ClientSecret: "sec"}

	resp := ExecuteRequest(req, map[string]string{}, DefaultHTTPClientConfig())
	if resp.Error != "" {
		t.Fatalf("request error: %s", resp.Error)
	}
	if gotAuth != "Bearer TOK123" {
		t.Fatalf("Authorization header = %q, want Bearer TOK123", gotAuth)
	}
}

func TestExportResponseAndRequest(t *testing.T) {
	resp := &Response{StatusCode: 200, Body: `{"ok":true}`}
	path, err := SaveResponseToFile(resp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("response file not written: %v", err)
	}
	if !strings.HasSuffix(path, ".json") {
		t.Errorf("expected .json extension, got %s", path)
	}

	req := NewRequest("t", "GET", "https://api.example.com/users")
	rpath, err := ExportRequestFile(req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(rpath, ".hilo.json") {
		t.Errorf("expected .hilo.json, got %s", rpath)
	}
}

// Pressing ctrl+i opens the import overlay; ctrl+s parses the pasted curl into
// the editor.
func TestCurlImportOverlayFlow(t *testing.T) {
	m := newModel()
	m = press(m, "ctrl+i")
	if m.overlay != overlayCurl {
		t.Fatalf("ctrl+i did not open curl overlay (overlay=%d)", m.overlay)
	}
	m.pasteArea.SetValue(`curl https://imported.test/api -H "X-Test: yes"`)
	next, _ := m.Update(key("ctrl+s"))
	m = next.(model)
	if m.overlay != overlayNone {
		t.Fatal("overlay did not close after import")
	}
	if m.urlInput.Value() != "https://imported.test/api" {
		t.Fatalf("url not populated from curl: %q", m.urlInput.Value())
	}
}

// OAuth2 has four navigable fields reachable by tab.
func TestOAuth2FieldNavigation(t *testing.T) {
	m := newModel()
	m.subTab = subTabAuth
	m.authType = authOAuth2
	m.focusArea = focusEditor
	m.authFieldIdx = 0
	m.focusCurrent()
	if got := len(m.authFields()); got != 4 {
		t.Fatalf("OAuth2 should expose 4 fields, got %d", got)
	}
	moved := m.focusNextAuthField()
	if !moved || m.authFieldIdx != 1 {
		t.Fatalf("tab did not advance auth field (moved=%v idx=%d)", moved, m.authFieldIdx)
	}
}

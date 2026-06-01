package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// snippetLangs is the ordered list of languages the snippet viewer cycles.
var snippetLangs = []string{"cURL", "Python", "JavaScript", "Go"}

// GenerateSnippet renders the request as runnable code in the given language
// (by index into snippetLangs), with environment variables resolved.
func GenerateSnippet(req Request, env map[string]string, langIdx int) string {
	switch langIdx {
	case 1:
		return pythonSnippet(req, env)
	case 2:
		return jsSnippet(req, env)
	case 3:
		return goSnippet(req, env)
	default:
		return GenerateCurl(req, env)
	}
}

// resolvedHeaders returns the request headers (plus any auth header) with
// variables resolved, sorted for stable output.
func resolvedHeaders(req Request, env map[string]string) [][2]string {
	merged := map[string]string{}
	for k, v := range req.Headers {
		merged[ResolveVars(k, env)] = ResolveVars(v, env)
	}
	if req.Auth != nil {
		switch req.Auth.Type {
		case AuthBearer:
			merged["Authorization"] = "Bearer " + ResolveVars(req.Auth.Value, env)
		case AuthAPIKey:
			merged[ResolveVars(req.Auth.Key, env)] = ResolveVars(req.Auth.Value, env)
		}
	}
	if req.BodyType != BodyNone && req.Body != "" {
		if _, ok := merged["Content-Type"]; !ok {
			merged["Content-Type"] = bodyContentType(req.BodyType)
		}
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([][2]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, [2]string{k, merged[k]})
	}
	return out
}

func resolvedURL(req Request, env map[string]string) string {
	return ResolveVars(req.URL, env)
}

func pythonSnippet(req Request, env map[string]string) string {
	var b strings.Builder
	b.WriteString("import requests\n\n")
	b.WriteString(fmt.Sprintf("url = %s\n", strconv.Quote(resolvedURL(req, env))))

	headers := resolvedHeaders(req, env)
	if len(headers) > 0 {
		b.WriteString("headers = {\n")
		for _, h := range headers {
			b.WriteString(fmt.Sprintf("    %s: %s,\n", strconv.Quote(h[0]), strconv.Quote(h[1])))
		}
		b.WriteString("}\n")
	}

	params := sortedPairs(req.QueryParams, env)
	if len(params) > 0 {
		b.WriteString("params = {\n")
		for _, p := range params {
			b.WriteString(fmt.Sprintf("    %s: %s,\n", strconv.Quote(p[0]), strconv.Quote(p[1])))
		}
		b.WriteString("}\n")
	}

	body := ResolveVars(req.Body, env)
	if body != "" && req.BodyType != BodyNone {
		b.WriteString(fmt.Sprintf("data = %s\n", strconv.Quote(body)))
	}

	b.WriteString("\nresponse = requests.request(\n")
	b.WriteString(fmt.Sprintf("    %s,\n    url,\n", strconv.Quote(req.Method)))
	if len(headers) > 0 {
		b.WriteString("    headers=headers,\n")
	}
	if len(params) > 0 {
		b.WriteString("    params=params,\n")
	}
	if body != "" && req.BodyType != BodyNone {
		b.WriteString("    data=data,\n")
	}
	if req.Auth != nil && req.Auth.Type == AuthBasic {
		b.WriteString(fmt.Sprintf("    auth=(%s, %s),\n",
			strconv.Quote(ResolveVars(req.Auth.Username, env)),
			strconv.Quote(ResolveVars(req.Auth.Password, env))))
	}
	b.WriteString(")\n\nprint(response.status_code)\nprint(response.text)\n")
	return b.String()
}

func jsSnippet(req Request, env map[string]string) string {
	var b strings.Builder
	headers := resolvedHeaders(req, env)
	body := ResolveVars(req.Body, env)

	b.WriteString("const options = {\n")
	b.WriteString(fmt.Sprintf("  method: %s,\n", strconv.Quote(req.Method)))
	if len(headers) > 0 {
		b.WriteString("  headers: {\n")
		for _, h := range headers {
			b.WriteString(fmt.Sprintf("    %s: %s,\n", strconv.Quote(h[0]), strconv.Quote(h[1])))
		}
		b.WriteString("  },\n")
	}
	if body != "" && req.BodyType != BodyNone {
		b.WriteString(fmt.Sprintf("  body: %s,\n", strconv.Quote(body)))
	}
	b.WriteString("};\n\n")
	b.WriteString(fmt.Sprintf("fetch(%s, options)\n", strconv.Quote(fullURL(req, env))))
	b.WriteString("  .then((res) => res.json())\n  .then((data) => console.log(data))\n  .catch((err) => console.error(err));\n")
	return b.String()
}

func goSnippet(req Request, env map[string]string) string {
	var b strings.Builder
	body := ResolveVars(req.Body, env)

	b.WriteString("package main\n\n")
	b.WriteString("import (\n\t\"fmt\"\n\t\"io\"\n\t\"net/http\"\n")
	if body != "" && req.BodyType != BodyNone {
		b.WriteString("\t\"strings\"\n")
	}
	b.WriteString(")\n\nfunc main() {\n")

	if body != "" && req.BodyType != BodyNone {
		b.WriteString(fmt.Sprintf("\tbody := strings.NewReader(%s)\n", strconv.Quote(body)))
		b.WriteString(fmt.Sprintf("\treq, _ := http.NewRequest(%s, %s, body)\n", strconv.Quote(req.Method), strconv.Quote(fullURL(req, env))))
	} else {
		b.WriteString(fmt.Sprintf("\treq, _ := http.NewRequest(%s, %s, nil)\n", strconv.Quote(req.Method), strconv.Quote(fullURL(req, env))))
	}

	for _, h := range resolvedHeaders(req, env) {
		b.WriteString(fmt.Sprintf("\treq.Header.Set(%s, %s)\n", strconv.Quote(h[0]), strconv.Quote(h[1])))
	}
	if req.Auth != nil && req.Auth.Type == AuthBasic {
		b.WriteString(fmt.Sprintf("\treq.SetBasicAuth(%s, %s)\n",
			strconv.Quote(ResolveVars(req.Auth.Username, env)),
			strconv.Quote(ResolveVars(req.Auth.Password, env))))
	}

	b.WriteString("\n\tresp, err := http.DefaultClient.Do(req)\n")
	b.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n\tdefer resp.Body.Close()\n\n")
	b.WriteString("\tout, _ := io.ReadAll(resp.Body)\n\tfmt.Println(resp.Status)\n\tfmt.Println(string(out))\n}\n")
	return b.String()
}

// fullURL appends resolved query params to the resolved URL for languages that
// don't take a separate params argument.
func fullURL(req Request, env map[string]string) string {
	base := resolvedURL(req, env)
	pairs := sortedPairs(req.QueryParams, env)
	if len(pairs) == 0 {
		return base
	}
	var qs []string
	for _, p := range pairs {
		qs = append(qs, p[0]+"="+p[1])
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + strings.Join(qs, "&")
}

func sortedPairs(m map[string]string, env map[string]string) [][2]string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([][2]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, [2]string{ResolveVars(k, env), ResolveVars(m[k], env)})
	}
	return out
}

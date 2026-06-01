package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Postman Collection v2.1 (subset we care about).
type pmCollection struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item []pmItem `json:"item"`
}

type pmItem struct {
	Name    string     `json:"name"`
	Request *pmRequest `json:"request,omitempty"`
	Item    []pmItem   `json:"item,omitempty"`
}

type pmRequest struct {
	Method string          `json:"method"`
	Header []pmKV          `json:"header"`
	Body   *pmBody         `json:"body,omitempty"`
	URL    json.RawMessage `json:"url"`
	Auth   *pmAuth         `json:"auth,omitempty"`
}

type pmKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type pmBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

type pmAuth struct {
	Type   string `json:"type"`
	Bearer []pmKV `json:"bearer"`
	Basic  []pmKV `json:"basic"`
}

// ImportPostmanCollection parses a Postman v2.1 collection into a hilo
// collection plus its requests (folders are flattened, names prefixed).
func ImportPostmanCollection(data []byte) (Collection, []Request, error) {
	var pc pmCollection
	if err := json.Unmarshal(data, &pc); err != nil {
		return Collection{}, nil, fmt.Errorf("invalid Postman JSON: %w", err)
	}
	name := strings.TrimSpace(pc.Info.Name)
	if name == "" {
		name = "imported"
	}

	col := NewCollection(name, "imported from Postman")
	var reqs []Request
	flattenPostman(pc.Item, "", &reqs)
	for i := range reqs {
		reqs[i].Collection = name
		col.Requests = append(col.Requests, reqs[i].ID)
	}
	return col, reqs, nil
}

func flattenPostman(items []pmItem, prefix string, out *[]Request) {
	for _, it := range items {
		if len(it.Item) > 0 {
			p := it.Name
			if prefix != "" {
				p = prefix + " / " + it.Name
			}
			flattenPostman(it.Item, p, out)
			continue
		}
		if it.Request == nil {
			continue
		}
		req := postmanToRequest(it)
		*out = append(*out, req)
	}
}

func postmanToRequest(it pmItem) Request {
	pr := it.Request
	req := NewRequest(it.Name, strings.ToUpper(pr.Method), postmanURL(pr.URL))
	if req.Method == "" {
		req.Method = MethodGet
	}

	for _, h := range pr.Header {
		if h.Key != "" {
			req.Headers[h.Key] = h.Value
		}
	}

	if pr.Body != nil && pr.Body.Raw != "" {
		req.Body = pr.Body.Raw
		if looksJSON(pr.Body.Raw) {
			req.BodyType = BodyJSON
		} else {
			req.BodyType = BodyText
		}
	}

	if pr.Auth != nil {
		switch pr.Auth.Type {
		case "bearer":
			req.Auth = &Auth{Type: AuthBearer, Value: pmFind(pr.Auth.Bearer, "token")}
		case "basic":
			req.Auth = &Auth{Type: AuthBasic,
				Username: pmFind(pr.Auth.Basic, "username"),
				Password: pmFind(pr.Auth.Basic, "password")}
		}
	}

	if req.Name == "" {
		req.Name = fmt.Sprintf("%s %s", req.Method, shortURL(req.URL))
	}
	return req
}

// postmanURL handles the two URL encodings Postman uses: a bare string, or an
// object with a "raw" field.
func postmanURL(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var obj struct {
		Raw string `json:"raw"`
	}
	if json.Unmarshal(raw, &obj) == nil {
		return obj.Raw
	}
	return ""
}

func pmFind(kvs []pmKV, key string) string {
	for _, kv := range kvs {
		if kv.Key == key {
			return kv.Value
		}
	}
	return ""
}

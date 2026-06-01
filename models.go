package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// HTTP methods
const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)

// Body types
const (
	BodyNone = "none"
	BodyJSON = "json"
	BodyForm = "form"
	BodyText = "text"
	BodyXML  = "xml"
	BodyFile = "file"
)

// Auth types
const (
	AuthNone   = "none"
	AuthBearer = "bearer"
	AuthBasic  = "basic"
	AuthDigest = "digest"
	AuthAPIKey = "apikey"
	AuthOAuth2 = "oauth2"
)

// Request represents an HTTP request saved in a collection.
type Request struct {
	ID          string            `json:"id"`
	Collection  string            `json:"collection,omitempty"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	Body        string            `json:"body,omitempty"`
	BodyType    string            `json:"body_type"`
	Auth        *Auth             `json:"auth,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Auth holds authentication details for a request.
type Auth struct {
	Type     string `json:"type"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Response represents the result of an executed HTTP request.
type Response struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	BodySize   int64             `json:"body_size"`
	Duration   time.Duration     `json:"duration_ms"`
	Timestamp  time.Time         `json:"timestamp"`
	Error      string            `json:"error,omitempty"`
}

// Environment holds a named set of key-value variables for request substitution.
type Environment struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

// Collection groups related requests with optional description and environment binding.
type Collection struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Requests    []string `json:"requests"`
}

// HistoryEntry is a single historical record pairing a request with its response.
type HistoryEntry struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// NewRequest creates a new Request with a generated ID and current timestamps.
func NewRequest(name, method, url string) Request {
	now := time.Now()
	return Request{
		ID:          generateID(),
		Name:        name,
		Method:      method,
		URL:         url,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		BodyType:    BodyNone,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewEnvironment creates a new Environment with initialized values map.
func NewEnvironment(name string) Environment {
	return Environment{
		Name:   name,
		Values: make(map[string]string),
	}
}

// NewCollection creates a new Collection with an initialized requests list.
func NewCollection(name, description string) Collection {
	return Collection{
		Name:        name,
		Description: description,
		Requests:    []string{},
	}
}

// generateID produces a random 16-hex-char identifier.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

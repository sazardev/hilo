package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Base directories under ~/.config/hilo/
func getBaseDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	return filepath.Join(dir, "hilo")
}

func getCollectionsDir() string { return filepath.Join(getBaseDir(), "collections") }
func getHistoryDir() string     { return filepath.Join(getBaseDir(), "history") }
func getEnvsDir() string        { return filepath.Join(getBaseDir(), "environments") }

// --- Requests (inside a collection) ---

func getRequestPath(collection, requestID string) string {
	return filepath.Join(getCollectionsDir(), collection, "requests", requestID+".json")
}

func SaveRequest(req Request) error {
	req.UpdatedAt = time.Now()
	path := getRequestPath(req.Collection, req.ID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadRequest(collection, requestID string) (Request, error) {
	path := getRequestPath(collection, requestID)
	data, err := os.ReadFile(path)
	if err != nil {
		return Request{}, err
	}
	var req Request
	err = json.Unmarshal(data, &req)
	return req, err
}

func DeleteRequest(collection, requestID string) error {
	return os.Remove(getRequestPath(collection, requestID))
}

func ListRequests(collection string) ([]Request, error) {
	dir := filepath.Join(getCollectionsDir(), collection, "requests")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Request{}, nil
		}
		return nil, err
	}

	var requests []Request
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			continue
		}
		requests = append(requests, req)
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].UpdatedAt.After(requests[j].UpdatedAt)
	})

	return requests, nil
}

// --- Collections ---

func getCollectionPath(name string) string {
	return filepath.Join(getCollectionsDir(), name, "collection.json")
}

func SaveCollection(col Collection) error {
	path := getCollectionPath(col.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(col, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadCollection(name string) (Collection, error) {
	path := getCollectionPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return Collection{}, err
	}
	var col Collection
	err = json.Unmarshal(data, &col)
	return col, err
}

func DeleteCollection(name string) error {
	dir := filepath.Join(getCollectionsDir(), name)
	return os.RemoveAll(dir)
}

func ListCollections() ([]Collection, error) {
	base := getCollectionsDir()
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return []Collection{}, nil
		}
		return nil, err
	}

	var cols []Collection
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		col, err := LoadCollection(e.Name())
		if err != nil {
			continue
		}
		cols = append(cols, col)
	}

	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Name < cols[j].Name
	})

	return cols, nil
}

// --- Environments ---

func getEnvPath(name string) string {
	return filepath.Join(getEnvsDir(), name+".json")
}

func SaveEnvironment(env Environment) error {
	path := getEnvPath(env.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadEnvironment(name string) (Environment, error) {
	path := getEnvPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return Environment{}, err
	}
	var env Environment
	err = json.Unmarshal(data, &env)
	return env, err
}

func DeleteEnvironment(name string) error {
	return os.Remove(getEnvPath(name))
}

func ListEnvironments() ([]Environment, error) {
	dir := getEnvsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Environment{}, nil
		}
		return nil, err
	}

	var envs []Environment
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var env Environment
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		envs = append(envs, env)
	}

	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Name < envs[j].Name
	})

	return envs, nil
}

// --- History ---

func getHistoryPath(entry HistoryEntry) string {
	ts := entry.Response.Timestamp.Format("2006-01-02T15-04-05")
	return filepath.Join(getHistoryDir(), ts+".json")
}

func SaveHistoryEntry(entry HistoryEntry) error {
	path := getHistoryPath(entry)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ListHistory() ([]HistoryEntry, error) {
	dir := getHistoryDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, err
	}

	var history []HistoryEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var h HistoryEntry
		if err := json.Unmarshal(data, &h); err != nil {
			continue
		}
		history = append(history, h)
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].Response.Timestamp.After(history[j].Response.Timestamp)
	})

	return history, nil
}

func ClearHistory() error {
	return os.RemoveAll(getHistoryDir())
}

// PruneHistory keeps only the most recent `limit` history files, deleting the
// oldest beyond it. A non-positive limit disables pruning.
func PruneHistory(limit int) {
	if limit <= 0 {
		return
	}
	dir := getHistoryDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, e.Name())
		}
	}
	if len(files) <= limit {
		return
	}

	// File names are timestamp-prefixed, so lexical sort is chronological.
	sort.Strings(files)
	for _, name := range files[:len(files)-limit] {
		_ = os.Remove(filepath.Join(dir, name))
	}
}

func DeleteHistoryEntry(entry HistoryEntry) error {
	path := getHistoryPath(entry)
	return os.Remove(path)
}

// --- Variable substitution ---

// ResolveVars replaces {{VAR}} placeholders in s using the provided values map.
func ResolveVars(s string, values map[string]string) string {
	for k, v := range values {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

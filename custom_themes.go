package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type customTheme struct {
	Name    string `json:"name"`
	Bg      string `json:"bg"`
	Surface string `json:"surface"`
	Border  string `json:"border"`
	Text    string `json:"text"`
	Muted   string `json:"muted"`
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
	Warn    string `json:"warn"`
}

func getThemesDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	return filepath.Join(dir, "hilo", "themes")
}

func loadCustomThemes() []theme {
	dir := getThemesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var custom []theme
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var ct customTheme
		if err := json.Unmarshal(data, &ct); err != nil {
			continue
		}
		t := customToTheme(ct)
		if t.name != "" {
			custom = append(custom, t)
		}
	}

	sort.Slice(custom, func(i, j int) bool {
		return custom[i].name < custom[j].name
	})

	return custom
}

func customToTheme(ct customTheme) theme {
	if ct.Name == "" {
		return theme{}
	}
	return theme{
		name:    ct.Name,
		bg:      lipgloss.Color(ct.Bg),
		surface: lipgloss.Color(ct.Surface),
		border:  lipgloss.Color(ct.Border),
		text:    lipgloss.Color(ct.Text),
		muted:   lipgloss.Color(ct.Muted),
		primary: lipgloss.Color(ct.Primary),
		accent:  lipgloss.Color(ct.Accent),
		warn:    lipgloss.Color(ct.Warn),
	}
}

func themeToCustom(t theme) customTheme {
	return customTheme{
		Name:    t.name,
		Bg:      string(t.bg),
		Surface: string(t.surface),
		Border:  string(t.border),
		Text:    string(t.text),
		Muted:   string(t.muted),
		Primary: string(t.primary),
		Accent:  string(t.accent),
		Warn:    string(t.warn),
	}
}

func saveCustomTheme(ct customTheme) error {
	if ct.Name == "" {
		return os.ErrInvalid
	}
	dir := getThemesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ct, "", "  ")
	if err != nil {
		return err
	}
	safeName := strings.ReplaceAll(ct.Name, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	return os.WriteFile(filepath.Join(dir, safeName+".json"), data, 0o644)
}

func deleteCustomTheme(name string) error {
	safeName := strings.ReplaceAll(name, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	return os.Remove(filepath.Join(getThemesDir(), safeName+".json"))
}

func importThemeJSON(data []byte) (customTheme, error) {
	var ct customTheme
	if err := json.Unmarshal(data, &ct); err != nil {
		return ct, err
	}
	if ct.Name == "" {
		ct.Name = "imported"
	}
	return ct, nil
}

func importThemesJSON(data []byte) ([]customTheme, error) {
	var themes []customTheme
	if err := json.Unmarshal(data, &themes); err != nil {
		var single customTheme
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return nil, err
		}
		if single.Name == "" {
			single.Name = "imported"
		}
		return []customTheme{single}, nil
	}
	for i := range themes {
		if themes[i].Name == "" {
			themes[i].Name = "imported"
		}
	}
	return themes, nil
}

func exportThemeJSON(t theme) ([]byte, error) {
	ct := themeToCustom(t)
	return json.MarshalIndent(ct, "", "  ")
}

func exportThemesJSON(themes []theme) ([]byte, error) {
	var cts []customTheme
	for _, t := range themes {
		cts = append(cts, themeToCustom(t))
	}
	return json.MarshalIndent(cts, "", "  ")
}

func themeExists(name string) bool {
	safeName := strings.ReplaceAll(name, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	path := filepath.Join(getThemesDir(), safeName+".json")
	_, err := os.Stat(path)
	return err == nil
}

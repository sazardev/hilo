package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fgColor extracts the foreground color string of a style for comparison.
func fgColor(st lipgloss.Style) string {
	if c, ok := st.GetForeground().(lipgloss.Color); ok {
		return string(c)
	}
	return ""
}

func key(s string) tea.KeyMsg {
	switch s {
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func press(m model, keys ...string) model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(model)
	}
	return m
}

// The headline bug: 'q' must be typed into the URL field, not quit the app.
func TestTypingQDoesNotQuit(t *testing.T) {
	m := newModel()
	m.focusArea = focusURL
	m.focusCurrent()
	m = press(m, "q", "u", "e", "r", "y")
	if got := m.urlInput.Value(); got != "query" {
		t.Fatalf("URL field = %q, want %q (q was likely treated as quit)", got, "query")
	}
}

// 'q' should still quit when not editing text.
func TestQQuitsWhenNotTyping(t *testing.T) {
	m := newModel()
	m.focusArea = focusResponse // not a text input
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Fatal("expected quit command when pressing q outside a text field")
	}
}

// Selecting a color scheme must override the theme's primary, proving styles
// are rebuilt from the live indices rather than the stale saved config.
func TestColorSchemeApplies(t *testing.T) {
	m := newModel()
	m.setColor(0) // blue
	blue := fgColor(m.styles.panelLabel)
	m.setColor(1) // coral
	coral := fgColor(m.styles.panelLabel)

	if blue == coral {
		t.Fatalf("changing color scheme did not change styles (%q)", blue)
	}
	if coral != string(colorSchemes[1].primary) {
		t.Errorf("panelLabel = %q, want color scheme primary %q", coral, colorSchemes[1].primary)
	}
}

// Ctrl+B cycles the body type.
func TestBodyTypeCycle(t *testing.T) {
	m := newModel()
	start := m.bodyType
	m = press(m, "ctrl+b")
	if m.bodyType == start {
		t.Fatal("ctrl+b did not change body type")
	}
}

// Activating an environment persists and feeds variable resolution.
func TestEnvActivationResolves(t *testing.T) {
	m := newModel()
	m.activeEnv = "" // start from a known state regardless of persisted config
	m.envs = []Environment{{Name: "dev", Values: map[string]string{"BASE_URL": "https://dev.example.com"}}}
	m.activeTab = tabEnvironments
	m.envIdx = 0
	m = press(m, "a")
	if m.activeEnv != "dev" {
		t.Fatalf("activeEnv = %q, want dev", m.activeEnv)
	}
	env := m.resolveEnv()
	if env["BASE_URL"] != "https://dev.example.com" {
		t.Fatalf("resolveEnv did not return active env values: %v", env)
	}
}

package main

import (
	"strings"
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

// Tab walks the request in visual order: URL → Actions → Sections → Editor.
func TestRequestFocusOrder(t *testing.T) {
	m := newModel()
	want := []focusArea{focusActions, focusSubTabs, focusEditor}
	for i, w := range want {
		m = press(m, "tab")
		if m.focusArea != w {
			t.Fatalf("after %d tabs focus=%d, want %d", i+1, m.focusArea, w)
		}
	}
}

// Inside the editor, Tab steps key→value→next field, then leaves to Response.
func TestEditorTabTraversal(t *testing.T) {
	m := newModel()
	m = press(m, "tab", "tab", "tab") // url → actions → subtabs → editor
	if m.focusArea != focusEditor || m.editorCol != 0 {
		t.Fatalf("not at editor key cell: area=%d col=%d", m.focusArea, m.editorCol)
	}
	m = press(m, "tab")
	if m.focusArea != focusEditor || m.editorCol != 1 {
		t.Fatalf("tab should move to value cell: area=%d col=%d", m.focusArea, m.editorCol)
	}
	m = press(m, "tab")
	if m.focusArea != focusResponse {
		t.Fatalf("tab past last field should reach response, got %d", m.focusArea)
	}
}

// Arrow keys switch the editor section when the selector row is focused.
func TestSubTabSwitchWithArrows(t *testing.T) {
	m := newModel()
	m = press(m, "tab", "tab") // → subtabs
	if m.focusArea != focusSubTabs {
		t.Fatalf("expected focusSubTabs, got %d", m.focusArea)
	}
	start := m.subTab
	m = press(m, "right")
	if m.subTab == start {
		t.Fatal("right arrow did not switch editor section")
	}
}

// Response scrolling is clamped and g/G jump to the bounds.
func TestResponseScrollKeys(t *testing.T) {
	m := newModel()
	m.response = sampleResponse()
	m.focusArea = focusResponse
	m.width, m.height = 100, 30

	for i := 0; i < 1000; i++ {
		m.scrollResponse(1)
	}
	if m.responseScroll > m.responseLineCount()-1 {
		t.Fatalf("scroll %d exceeded line count %d", m.responseScroll, m.responseLineCount())
	}
	m = press(m, "g")
	if m.responseScroll != 0 {
		t.Fatalf("g should jump to top, got %d", m.responseScroll)
	}
	m = press(m, "G")
	if m.responseScroll == 0 {
		t.Fatal("G should jump to the bottom")
	}
}

// End-to-end Git UI: create a collection, generate commits, then exercise the
// log, diff and branch views against a real go-git repository.
func TestGitUIFlow(t *testing.T) {
	const name = "gitflow"
	SaveCollection(NewCollection(name, ""))
	if _, err := InitCollectionRepo(name); err != nil {
		t.Fatalf("init repo: %v", err)
	}
	repo, err := OpenCollectionRepo(name)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}

	req := NewRequest("list users", "GET", "https://api.example.com/users")
	req.Collection = name
	if err := repo.SaveRequest(req, CommitAdd); err != nil {
		t.Fatalf("commit add: %v", err)
	}
	req.URL = "https://api.example.com/users?page=2"
	if err := repo.SaveRequest(req, CommitUpdate); err != nil {
		t.Fatalf("commit update: %v", err)
	}

	m := newModel()
	m.collections, _ = ListCollections()
	for i, c := range m.collections {
		if c.Name == name {
			m.collIdx = i
		}
	}
	m.activeTab = tabCollections
	m.collMode = collDetail
	m.loadCollectionReqs()
	m.collReqIdx = 0

	m.openGitLog()
	if m.collMode != collGitLog {
		t.Fatalf("openGitLog did not switch mode: %d", m.collMode)
	}
	if len(m.gitLog) < 3 { // init + add + update
		t.Fatalf("expected >=3 commits, got %d", len(m.gitLog))
	}
	if m.gitReqPath == "" {
		t.Fatal("git diff target not set from selected request")
	}

	// Diff the oldest commit against HEAD — should be a non-empty addition.
	m.gitLogIdx = len(m.gitLog) - 1
	m.showGitDiff()
	if m.collMode != collGitDiff || strings.TrimSpace(m.gitDiff) == "" {
		t.Fatalf("diff not produced (mode=%d, len=%d)", m.collMode, len(m.gitDiff))
	}

	// Revert the request file to the first commit that introduced it.
	m.gitLogIdx = len(m.gitLog) - 2
	m.revertToCommit()
	if !strings.Contains(m.message, "reverted") {
		t.Fatalf("revert message unexpected: %q", m.message)
	}

	// Branch view lists at least the default branch.
	m.openGitBranches()
	if m.collMode != collGitBranch || len(m.gitBranches) == 0 {
		t.Fatalf("branches not listed (mode=%d, n=%d)", m.collMode, len(m.gitBranches))
	}
}

// The mouse wheel scrolls the response body.
func TestMouseWheelScrollsResponse(t *testing.T) {
	m := newModel()
	m.response = sampleResponse()
	m.activeTab = tabRequest
	next, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	m = next.(model)
	if m.responseScroll == 0 {
		t.Fatal("wheel down did not scroll the response")
	}
}

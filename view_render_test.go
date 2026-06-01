package main

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderAt drives View() at a given size and returns the output, failing on
// any panic so layout-math bugs surface as test failures.
func renderAt(t *testing.T, m model, w, h int) string {
	t.Helper()
	m.width, m.height = w, h
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View() panicked at %dx%d: %v", w, h, r)
		}
	}()
	return m.View()
}

func sampleResponse() *Response {
	body := strings.Repeat("{\n  \"id\": 42,\n  \"name\": \"John Doe\"\n}\n", 30)
	return &Response{
		StatusCode: 200,
		StatusText: "200 OK",
		Headers:    map[string]string{"Content-Type": "application/json", "Server": "nginx"},
		Body:       body,
		BodySize:   int64(len(body)),
		Duration:   342 * time.Millisecond,
	}
}

func TestViewRendersAllSizesAndTabs(t *testing.T) {
	sizes := [][2]int{{130, 40}, {90, 30}, {60, 24}, {45, 14}}
	tabs := []tab{tabRequest, tabCollections, tabHistory, tabEnvironments, tabConfig, tabHelp}

	for _, sz := range sizes {
		for _, tb := range tabs {
			m := newModel()
			m.activeTab = tb
			m.response = sampleResponse()
			out := renderAt(t, m, sz[0], sz[1])

			if out == "" {
				t.Fatalf("empty render at %dx%d tab=%d", sz[0], sz[1], tb)
			}
			// The rendered frame must never exceed the terminal height.
			if got := lipgloss.Height(out); got > sz[1] {
				t.Errorf("render too tall at %dx%d tab=%d: %d > %d", sz[0], sz[1], tb, got, sz[1])
			}
			// And no single line may exceed the terminal width.
			for i, line := range strings.Split(out, "\n") {
				if wln := lipgloss.Width(line); wln > sz[0] {
					t.Errorf("line %d too wide at %dx%d tab=%d: %d > %d", i, sz[0], sz[1], tb, wln, sz[0])
					break
				}
			}
		}
	}
}

func TestResponseFocusAndSearch(t *testing.T) {
	m := newModel()
	m.response = sampleResponse()
	m.focusArea = focusResponse
	m.searchInput.SetValue("John")
	m.computeSearchMatches()
	if len(m.searchMatches) == 0 {
		t.Fatal("expected search matches for 'John'")
	}
	_ = renderAt(t, m, 130, 40)
}

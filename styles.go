package main

import "github.com/charmbracelet/lipgloss"

type styles struct {
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	tabBar      lipgloss.Style
	helpHint    lipgloss.Style
	tabBarOuter lipgloss.Style

	page lipgloss.Style

	title  lipgloss.Style
	body   lipgloss.Style
	muted  lipgloss.Style
	accent lipgloss.Style

	hello     lipgloss.Style
	subtitle  lipgloss.Style
	dot       lipgloss.Style
	separator lipgloss.Style

	sectionHead lipgloss.Style
	bullet      lipgloss.Style

	keyLabel lipgloss.Style
	keyDesc  lipgloss.Style

	sidePanel     lipgloss.Style
	sideTab       lipgloss.Style
	sideTabActive lipgloss.Style

	configCursor lipgloss.Style
	configActive lipgloss.Style
	configNormal lipgloss.Style

	themeCursor lipgloss.Style
	themeActive lipgloss.Style
	themeDot    lipgloss.Style

	warn lipgloss.Style
}

func newStyles(t theme) *styles {
	s := &styles{}

	s.tabActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 2)

	s.tabInactive = lipgloss.NewStyle().
		Foreground(t.muted).
		Padding(0, 2)

	s.tabBar = lipgloss.NewStyle().Padding(1, 0)
	s.tabBarOuter = lipgloss.NewStyle().Padding(0, 1)

	s.helpHint = lipgloss.NewStyle().
		Foreground(t.muted).
		Italic(true)

	s.page = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.border).
		Padding(2, 4)

	s.title = lipgloss.NewStyle().Foreground(t.text).Bold(true)
	s.body = lipgloss.NewStyle().Foreground(t.text)
	s.muted = lipgloss.NewStyle().Foreground(t.muted)
	s.accent = lipgloss.NewStyle().Foreground(t.accent)

	s.hello = lipgloss.NewStyle().Foreground(t.accent).Bold(true)
	s.subtitle = lipgloss.NewStyle().Foreground(t.muted)
	s.dot = lipgloss.NewStyle().Foreground(t.primary)
	s.separator = lipgloss.NewStyle().Foreground(t.border)

	s.sectionHead = lipgloss.NewStyle().
		Foreground(t.primary).
		Bold(true).
		MarginTop(1)

	s.bullet = lipgloss.NewStyle().Foreground(t.muted).PaddingLeft(2)

	s.keyLabel = lipgloss.NewStyle().
		Foreground(t.accent).
		Bold(true).
		Width(16)

	s.keyDesc = lipgloss.NewStyle().Foreground(t.text)

	s.sidePanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.border).
		Padding(1, 2).
		MarginRight(1)

	s.sideTab = lipgloss.NewStyle().Foreground(t.muted).Padding(0, 1)

	s.sideTabActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 1)

	s.configCursor = lipgloss.NewStyle().Foreground(t.accent).Bold(true)

	s.configActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 2)

	s.configNormal = lipgloss.NewStyle().Foreground(t.muted).Padding(0, 2)

	s.themeCursor = lipgloss.NewStyle().Foreground(t.accent).Bold(true)

	s.themeActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 2)

	s.themeDot = lipgloss.NewStyle().Foreground(t.primary)

	s.warn = lipgloss.NewStyle().
		Foreground(t.warn).
		Bold(true)

	return s
}

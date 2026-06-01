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

	methodGet     lipgloss.Style
	methodPost    lipgloss.Style
	methodPut     lipgloss.Style
	methodPatch   lipgloss.Style
	methodDelete  lipgloss.Style
	methodHead    lipgloss.Style
	methodOptions lipgloss.Style

	urlInput      lipgloss.Style
	urlInputFocus lipgloss.Style

	subTabActive   lipgloss.Style
	subTabInactive lipgloss.Style

	bodyInput      lipgloss.Style
	bodyInputFocus lipgloss.Style

	btnSend lipgloss.Style
	btn     lipgloss.Style

	statusOK   lipgloss.Style
	statusErr  lipgloss.Style
	statusWarn lipgloss.Style

	editorHeader    lipgloss.Style
	editorRowActive lipgloss.Style
	btnSendActive   lipgloss.Style
	btnActive       lipgloss.Style

	panel      lipgloss.Style
	panelFocus lipgloss.Style
	panelLabel lipgloss.Style

	footerBar  lipgloss.Style
	footerKey  lipgloss.Style
	footerText lipgloss.Style

	envBadge    lipgloss.Style
	envBadgeOff lipgloss.Style

	searchBar lipgloss.Style
	matchHi   lipgloss.Style
}

func newStyles(t theme) *styles {
	s := &styles{}

	s.tabActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 1)

	s.tabInactive = lipgloss.NewStyle().
		Foreground(t.muted).
		Padding(0, 1)

	s.tabBar = lipgloss.NewStyle().Padding(1, 0)
	s.tabBarOuter = lipgloss.NewStyle().Padding(0, 1)

	s.helpHint = lipgloss.NewStyle().
		Foreground(t.muted).
		Italic(true)

	s.page = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.border).
		Padding(2, 2)

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

	s.methodGet = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ADE80")).
		Bold(true).
		Background(lipgloss.Color("#1A3A2A")).
		Padding(0, 1)

	s.methodPost = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#60A5FA")).
		Bold(true).
		Background(lipgloss.Color("#1A2A3A")).
		Padding(0, 1)

	s.methodPut = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FBBF24")).
		Bold(true).
		Background(lipgloss.Color("#3A3A1A")).
		Padding(0, 1)

	s.methodPatch = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#C084FC")).
		Bold(true).
		Background(lipgloss.Color("#2A1A3A")).
		Padding(0, 1)

	s.methodDelete = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F87171")).
		Bold(true).
		Background(lipgloss.Color("#3A1A1A")).
		Padding(0, 1)

	s.methodHead = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Bold(true).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 1)

	s.methodOptions = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Bold(true).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 1)

	s.urlInput = lipgloss.NewStyle().
		Foreground(t.muted).
		Padding(0, 1)

	s.urlInputFocus = lipgloss.NewStyle().
		Foreground(t.text).
		Padding(0, 1)

	s.subTabActive = lipgloss.NewStyle().
		Foreground(t.primary).
		Bold(true)

	s.subTabInactive = lipgloss.NewStyle().
		Foreground(t.muted)

	s.bodyInput = lipgloss.NewStyle().
		Foreground(t.text).
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.border).
		Padding(0, 1)

	s.bodyInputFocus = lipgloss.NewStyle().
		Foreground(t.text).
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.primary).
		Padding(0, 1)

	s.btnSend = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 2)

	s.btn = lipgloss.NewStyle().
		Foreground(t.muted).
		Padding(0, 1)

	s.statusOK = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ADE80")).
		Bold(true)

	s.statusErr = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F87171")).
		Bold(true)

	s.statusWarn = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FBBF24")).
		Bold(true)

	s.editorHeader = lipgloss.NewStyle().
		Foreground(t.muted).
		Bold(true)

	s.editorRowActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text)

	s.btnSendActive = lipgloss.NewStyle().
		Background(t.primary).
		Foreground(t.text).
		Bold(true).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.accent)

	s.btnActive = lipgloss.NewStyle().
		Foreground(t.accent).
		Bold(true).
		Padding(0, 1)

	s.panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.border).
		Padding(0, 1)

	s.panelFocus = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.primary).
		Padding(0, 1)

	s.panelLabel = lipgloss.NewStyle().
		Foreground(t.primary).
		Bold(true)

	s.footerBar = lipgloss.NewStyle().
		Foreground(t.muted).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.border).
		Padding(0, 1)

	s.footerKey = lipgloss.NewStyle().
		Foreground(t.accent).
		Bold(true)

	s.footerText = lipgloss.NewStyle().Foreground(t.muted)

	s.envBadge = lipgloss.NewStyle().
		Background(t.accent).
		Foreground(t.bg).
		Bold(true).
		Padding(0, 1)

	s.envBadgeOff = lipgloss.NewStyle().
		Foreground(t.muted).
		Padding(0, 1)

	s.searchBar = lipgloss.NewStyle().
		Foreground(t.text).
		Background(t.surface).
		Padding(0, 1)

	s.matchHi = lipgloss.NewStyle().
		Background(t.warn).
		Foreground(t.bg).
		Bold(true)

	return s
}

func methodStyle(s *styles, method string) lipgloss.Style {
	switch method {
	case "POST":
		return s.methodPost
	case "PUT":
		return s.methodPut
	case "PATCH":
		return s.methodPatch
	case "DELETE":
		return s.methodDelete
	case "HEAD":
		return s.methodHead
	case "OPTIONS":
		return s.methodOptions
	default:
		return s.methodGet
	}
}

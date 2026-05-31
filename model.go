package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tabHome = iota
	tabAbout
	tabConfig
	tabHelp
	tabChangelog
)

var tabNames = []string{"home", "about", "config"}

type model struct {
	activeTab     int
	width         int
	height        int
	isSmall       bool
	showSide      bool
	styles        *styles
	spinner       spinner.Model
	themeIdx      int
	colorIdx      int
	modeIdx       int // 0 = dark, 1 = light
	configSection int // 0 = theme, 1 = color, 2 = mode
	configCursor  int
	colorScroll   int
}

func newModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	cfg := loadConfig()
	m := model{
		activeTab: tabHome,
		themeIdx:  cfg.Theme,
		colorIdx:  cfg.Color,
		modeIdx:   cfg.Mode,
		spinner:   sp,
	}
	m.rebuildStyles()
	return m
}

func (m *model) rebuildStyles() {
	var t theme

	if m.colorIdx > 0 {
		cs := colorSchemes[m.colorIdx]
		t = theme{
			name:    cs.name,
			bg:      lipgloss.Color("#0F0F23"),
			surface: lipgloss.Color("#1E1E38"),
			border:  lipgloss.Color("#3A3A5C"),
			text:    lipgloss.Color("#F0F0F5"),
			muted:   lipgloss.Color("#8888A0"),
			primary: cs.primary,
			accent:  cs.accent,
			warn:    lipgloss.Color("#FFD700"),
		}
	} else {
		t = themes[m.themeIdx]
	}

	if m.modeIdx == 1 {
		t = toLight(t)
	}

	m.styles = newStyles(t)
	m.spinner.Style = lipgloss.NewStyle().Foreground(t.accent)
}

func toLight(t theme) theme {
	return theme{
		name:    t.name,
		bg:      lipgloss.Color("#F5F5FA"),
		surface: lipgloss.Color("#E8E8F0"),
		border:  lipgloss.Color("#C8C8D8"),
		text:    lipgloss.Color("#1A1A2E"),
		muted:   lipgloss.Color("#6B6B8D"),
		primary: t.primary,
		accent:  t.accent,
		warn:    lipgloss.Color("#D4A017"),
	}
}

func (m *model) setTheme(idx int) {
	m.themeIdx = idx
	m.colorIdx = 0
	m.rebuildStyles()
}

func (m *model) setColor(idx int) {
	m.colorIdx = idx
	m.rebuildStyles()
}

func (m *model) setMode(idx int) {
	m.modeIdx = idx
	m.rebuildStyles()
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.isSmall = msg.Width < 70
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.activeTab == tabHelp || m.activeTab == tabChangelog {
				m.activeTab = tabHome
				return m, nil
			}
			return m, tea.Quit

		case "?":
			if m.activeTab == tabHelp || m.activeTab == tabChangelog {
				m.activeTab = tabHome
			} else {
				m.activeTab = tabHelp
			}
			return m, nil

		case "c":
			if m.activeTab == tabHelp {
				m.activeTab = tabChangelog
				return m, nil
			}

		case "enter":
			if m.activeTab == tabConfig {
				switch m.configSection {
				case 0:
					m.setTheme(m.configCursor)
				case 1:
					m.setColor(m.configCursor)
				case 2:
					m.setMode(m.configCursor)
				case 3:
					if m.configCursor == 0 {
						deleteConfig()
						m.themeIdx = 0
						m.colorIdx = 0
						m.modeIdx = 0
						m.rebuildStyles()
					}
				}
				if m.configSection < 3 {
					saveConfig(appConfig{Theme: m.themeIdx, Color: m.colorIdx, Mode: m.modeIdx})
				}
				return m, nil
			}

		case "up", "k":
			if m.activeTab == tabConfig {
				m.configCursor--
				max := m.configMax()
				if m.configCursor < 0 {
					m.configCursor = max
				}
				if m.configSection == 1 {
					const visibleColors = 10
					if m.configCursor < m.colorScroll {
						m.colorScroll = m.configCursor
					}
					if m.colorScroll < 0 {
						m.colorScroll = 0
					}
				}
			}
			return m, nil

		case "down", "j":
			if m.activeTab == tabConfig {
				m.configCursor++
				max := m.configMax()
				if m.configCursor > max {
					m.configCursor = 0
				}
				if m.configSection == 1 {
					const visibleColors = 10
					if m.configCursor >= m.colorScroll+visibleColors {
						m.colorScroll = m.configCursor - visibleColors + 1
					}
					total := len(colorSchemes)
					if m.colorScroll > total-visibleColors {
						m.colorScroll = total - visibleColors
					}
					if m.colorScroll < 0 {
						m.colorScroll = 0
					}
				}
			}
			return m, nil

		case "tab":
			if m.activeTab == tabConfig {
				m.configSection = (m.configSection + 1) % 4
				m.configCursor = m.configCurrent()
				if m.configSection == 1 {
					m.colorScroll = max(0, m.configCursor-9)
				} else {
					m.colorScroll = 0
				}
				return m, nil
			}
			if !m.isSmall {
				m.showSide = !m.showSide
			}
			return m, nil

		case "left", "h":
			if m.activeTab != tabHelp && m.activeTab != tabChangelog {
				m.activeTab--
				if m.activeTab < 0 {
					m.activeTab = len(tabNames) - 1
				}
				m.configSection = 0
				m.configCursor = m.configCurrent()
			}
			return m, nil

		case "right", "l":
			if m.activeTab != tabHelp && m.activeTab != tabChangelog {
				m.activeTab++
				if m.activeTab >= len(tabNames) {
					m.activeTab = 0
				}
				m.configSection = 0
				m.configCursor = m.configCurrent()
			}
			return m, nil
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) configMax() int {
	switch m.configSection {
	case 0:
		return len(themes) - 1
	case 1:
		return len(colorSchemes) - 1
	case 2:
		return 1
	case 3:
		return 0
	}
	return 0
}

func (m model) configCurrent() int {
	switch m.configSection {
	case 0:
		if m.colorIdx == 0 {
			return m.themeIdx
		}
		return 0
	case 1:
		return m.colorIdx
	case 2:
		return m.modeIdx
	case 3:
		return 0
	}
	return 0
}

func (m model) View() string {
	if m.width < 36 || m.height < 8 {
		return m.styles.warn.Render("\n  terminal too small!\n  minimum 36x8 required.\n")
	}

	if m.showSide {
		return m.viewSidebar()
	}

	return m.viewNormal()
}

func (m model) viewNormal() string {
	s := m.styles

	tabsRendered := m.renderTabsInline()

	hintText := "? help"
	if m.width < 50 {
		hintText = "?"
	}
	hint := s.helpHint.Render(hintText)

	tabsW := lipgloss.Width(tabsRendered)
	hintW := lipgloss.Width(hint)
	gap := max(m.width-tabsW-hintW-4, 1)
	spacer := lipgloss.NewStyle().Width(gap).Render("")
	topBar := s.tabBarOuter.Render(
		lipgloss.JoinHorizontal(lipgloss.Center, tabsRendered, spacer, hint),
	)

	barH := lipgloss.Height(topBar)
	content := m.renderContent()

	availH := max(m.height-barH-2, 3)

	page := s.page.
		Width(m.width-4).
		Height(availH).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topBar, page)
}

func (m model) viewSidebar() string {
	s := m.styles
	side := m.renderSidePanel()

	sideW := lipgloss.Width(side)
	contentW := max(m.width-sideW-4, 20)

	content := m.renderContent()

	page := s.page.
		Width(contentW).
		Height(m.height-4).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	return lipgloss.JoinHorizontal(lipgloss.Top, side, page)
}

func (m model) renderTabsInline() string {
	s := m.styles
	var tabs []string

	sep := s.muted.Render(" . ")

	for i, name := range tabNames {
		if i == m.activeTab {
			tabs = append(tabs, s.tabActive.Render(name))
		} else {
			tabs = append(tabs, s.tabInactive.Render(name))
		}
		if i < len(tabNames)-1 {
			tabs = append(tabs, sep)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m model) renderSidePanel() string {
	s := m.styles
	var items []string

	for i, name := range tabNames {
		if i == m.activeTab {
			items = append(items, s.sideTabActive.Render(name))
		} else {
			items = append(items, s.sideTab.Render(name))
		}
	}

	helpStyle := s.sideTab
	if m.activeTab == tabHelp || m.activeTab == tabChangelog {
		helpStyle = s.sideTabActive
	}
	items = append(items, helpStyle.Render("? help"))

	panel := lipgloss.JoinVertical(lipgloss.Center, items...)
	return s.sidePanel.Render(panel)
}

func (m model) renderContent() string {
	switch m.activeTab {
	case tabHome:
		return m.viewHome()
	case tabAbout:
		return m.viewAbout()
	case tabConfig:
		return m.viewConfig()
	case tabHelp:
		return m.viewHelp()
	case tabChangelog:
		return m.viewChangelog()
	}
	return ""
}

func (m model) viewHome() string {
	s := m.styles

	dots := s.dot.Render("◆  ◆  ◆")

	hello := s.hello.
		MarginTop(1).
		MarginBottom(1).
		Render("hello world")

	sub := s.subtitle.Render("welcome to the modern tui")
	sep := s.separator.Render(strings.Repeat("─", min(28, m.width-14)))
	built := s.muted.Render("built with bubble tea")
	spin := s.accent.Render(m.spinner.View() + " ready")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		dots, hello, sep, sub, "", built, spin,
	)
}

func (m model) viewAbout() string {
	s := m.styles

	title := s.title.Render("about")
	sep := s.separator.Render(strings.Repeat("─", min(28, m.width-14)))
	version := s.body.Render("version     1.0.0")
	built := s.body.Render("framework   bubble tea + lipgloss")

	head1 := s.sectionHead.Render("description")
	desc := s.body.Render("a modern terminal user interface\ndemonstrating responsive design\nand clean aesthetics.")

	head2 := s.sectionHead.Render("features")
	f1 := s.bullet.Render("  responsive layout")
	f2 := s.bullet.Render("  color themes")
	f3 := s.bullet.Render("  minimalist design")
	f4 := s.bullet.Render("  keyboard navigation")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title, sep, "", version, built, "",
		head1, desc, "", head2, f1, f2, f3, f4,
	)
}

func (m model) viewConfig() string {
	s := m.styles

	title := s.title.Render("config")
	sep := s.separator.Render(strings.Repeat("─", min(28, m.width-14)))

	// Section tabs
	sections := []string{"theme", "color", "mode", "user"}
	var secParts []string
	for i, name := range sections {
		if i == m.configSection {
			secParts = append(secParts, s.accent.Bold(true).Render(name))
		} else {
			secParts = append(secParts, s.muted.Render(name))
		}
		if i < len(sections)-1 {
			secParts = append(secParts, s.muted.Render("  |  "))
		}
	}
	secBar := lipgloss.JoinHorizontal(lipgloss.Top, secParts...)

	var items []string

	switch m.configSection {
	case 0:
		for i, t := range themes {
			cursor := "  "
			if i == m.configCursor {
				cursor = s.themeCursor.Render("▸ ")
			}

			preview := lipgloss.NewStyle().
				Background(t.primary).
				Foreground(t.text).
				Render("  ")

			label := s.body.Render(t.name)

			active := ""
			if i == m.themeIdx && m.colorIdx == 0 {
				active = s.muted.Render("  (active)")
			}

			row := lipgloss.JoinHorizontal(lipgloss.Center, cursor, preview, " ", label, active)
			items = append(items, row)
		}

	case 1:
		const visibleColors = 10
		total := len(colorSchemes)
		start := m.colorScroll
		end := start + visibleColors
		if end > total {
			end = total
		}
		for i := start; i < end; i++ {
			cs := colorSchemes[i]
			cursor := "  "
			if i == m.configCursor {
				cursor = s.configCursor.Render("▸ ")
			}

			dot := lipgloss.NewStyle().
				Background(cs.primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render("  ")

			label := s.body.Render(cs.name)

			active := ""
			if i == m.colorIdx {
				active = s.muted.Render("  (active)")
			}

			row := lipgloss.JoinHorizontal(lipgloss.Center, cursor, dot, " ", label, active)
			items = append(items, row)
		}
		if total > visibleColors {
			scrollInfo := s.muted.Render(fmt.Sprintf("  %d/%d", m.configCursor+1, total))
			items = append(items, scrollInfo)
		}

	case 2:
		modes := []string{"dark", "light"}
		for i, name := range modes {
			cursor := "  "
			if i == m.configCursor {
				cursor = s.themeCursor.Render("▸ ")
			}

			label := s.body.Render(name)

			active := ""
			if i == m.modeIdx {
				active = s.muted.Render("  (active)")
			}

			row := lipgloss.JoinHorizontal(lipgloss.Center, cursor, label, active)
			items = append(items, row)
		}

	case 3:
		head1 := s.sectionHead.Render("config path")
		path := s.muted.Render("  " + getConfigPath())
		items = append(items, head1, path, "")

		head2 := s.sectionHead.Render("actions")
		cursor := "  "
		if m.configCursor == 0 {
			cursor = s.themeCursor.Render("▸ ")
		}
		resetRow := lipgloss.JoinHorizontal(lipgloss.Center, cursor, s.body.Render("reset all settings"))
		items = append(items, head2, resetRow)
	}

	head3 := s.sectionHead.Render("controls")
	hint1 := s.muted.Render("  ↑/k ↓/j  navigate")
	hint2 := s.muted.Render("  enter     apply")
	hint3 := s.muted.Render("  tab       switch section")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title, sep, "",
		secBar, "",
		lipgloss.JoinVertical(lipgloss.Left, items...),
		"", head3, hint1, hint2, hint3,
	)
}

func (m model) viewHelp() string {
	s := m.styles

	title := s.title.Render("keyboard shortcuts")
	sep := s.separator.Render(strings.Repeat("─", min(28, m.width-14)))

	head1 := s.sectionHead.Render("navigation")

	row := func(key, desc string) string {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			s.keyLabel.Render(key),
			s.keyDesc.Render(desc),
		)
	}

	nav1 := row("  ← / h", "previous tab")
	nav2 := row("  → / l", "next tab")

	head2 := s.sectionHead.Render("general")

	gen1 := row("  ?", "toggle this help page")
	gen2 := row("  c", "view changelog")
	gen3 := row("  q / esc", "quit application")
	gen4 := row("  tab", "toggle sidebar / switch section")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title, sep, "",
		head1, nav1, nav2, "",
		head2, gen1, gen2, gen3, gen4,
	)
}

func (m model) viewChangelog() string {
	s := m.styles

	title := s.title.Render("changelog")
	sep := s.separator.Render(strings.Repeat("─", min(28, m.width-14)))

	row := func(ver, date, desc string) string {
		v := s.accent.Render(ver)
		d := s.muted.Render(date)
		return lipgloss.JoinHorizontal(lipgloss.Top, v, "  ", d, "  ", s.body.Render(desc))
	}

	v1 := row("v1.3.0", "2026-05-31", "mode: dark / light")
	v2 := row("", "", "  light mode palette")
	v3 := row("v1.2.0", "2026-05-31", "themes: nord, gruvbox, terminal")
	v4 := row("", "", "  full palette switching")
	v5 := row("v1.1.0", "2026-05-31", "config: color accents")
	v6 := row("v1.0.0", "2026-05-31", "initial release")
	v7 := row("", "", "  home, about, help, changelog")
	v8 := row("", "", "  responsive sidebar layout")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title, sep, "",
		v1, v2, "",
		v3, v4, "",
		v5, "",
		v6, v7, v8,
	)
}

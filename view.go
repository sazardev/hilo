package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ---------------------------------------------------------------------------
// Top-level layout
// ---------------------------------------------------------------------------

func (m model) View() string {
	if m.width < 40 || m.height < 12 {
		return m.styles.warn.Render("\n  terminal too small\n  need at least 40 x 12\n")
	}
	if m.importMode {
		return m.viewImport()
	}
	return m.viewMain()
}

func (m model) viewMain() string {
	top := m.renderTopBar()
	footer := m.renderFooter()

	bodyH := m.height - lipgloss.Height(top) - lipgloss.Height(footer)
	if bodyH < 3 {
		bodyH = 3
	}
	body := clipHeight(m.renderBody(m.width, bodyH), bodyH)

	return lipgloss.JoinVertical(lipgloss.Left, top, body, footer)
}

// renderTopBar draws the tab strip on the left and the active-environment
// badge plus help hint on the right, degrading gracefully as width shrinks.
func (m model) renderTopBar() string {
	s := m.styles

	var env string
	if m.activeEnv != "" {
		env = s.envBadge.Render("● " + m.activeEnv)
	} else {
		env = s.envBadgeOff.Render("○ no env")
	}
	right := lipgloss.JoinHorizontal(lipgloss.Center, env, " ", s.helpHint.Render("? help"))

	avail := m.width - 2 // outer padding
	tabs := m.renderTabs(avail - lipgloss.Width(right) - 1)

	// Drop the right-hand cluster if there still isn't room.
	if lipgloss.Width(tabs)+lipgloss.Width(right)+1 > avail {
		right = s.helpHint.Render("?")
	}
	if lipgloss.Width(tabs)+lipgloss.Width(right)+1 > avail {
		right = ""
	}

	gap := max(avail-lipgloss.Width(tabs)-lipgloss.Width(right), 1)
	bar := clipLine(tabs+strings.Repeat(" ", gap)+right, avail)
	return lipgloss.NewStyle().Padding(0, 1).Render(bar)
}

// renderTabs picks the most descriptive label set that fits the budget:
// numbered → names → abbreviations.
func (m model) renderTabs(budget int) string {
	numbered := m.tabStrip(func(i int, name string) string {
		if i >= 5 {
			return name
		}
		return fmt.Sprintf("%d %s", i+1, name)
	})
	if lipgloss.Width(numbered) <= budget {
		return numbered
	}

	names := m.tabStrip(func(i int, name string) string { return name })
	if lipgloss.Width(names) <= budget {
		return names
	}

	short := []string{"Req", "Col", "Hist", "Env", "Cfg", "?"}
	return m.tabStrip(func(i int, _ string) string { return short[i] })
}

func (m model) tabStrip(label func(i int, name string) string) string {
	s := m.styles
	var parts []string
	for i, name := range tabNames {
		if tab(i) == m.activeTab {
			parts = append(parts, s.tabActive.Render(label(i, name)))
		} else {
			parts = append(parts, s.tabInactive.Render(label(i, name)))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// renderFooter draws the contextual key hints and the latest status message.
func (m model) renderFooter() string {
	s := m.styles
	left := m.contextHints()
	right := ""
	if m.message != "" {
		right = s.footerKey.Render(m.message)
	}

	avail := m.width - 2 // content width inside the bar's horizontal padding
	gap := avail - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		right = ""
		gap = max(avail-lipgloss.Width(left), 1)
	}
	bar := clipLine(left+strings.Repeat(" ", gap)+right, avail)
	return s.footerBar.Width(m.width).Render(bar)
}

func (m model) hint(key, desc string) string {
	s := m.styles
	return s.footerKey.Render(key) + " " + s.footerText.Render(desc)
}

func (m model) contextHints() string {
	var parts []string
	switch m.activeTab {
	case tabRequest:
		if m.focusArea == focusResponse {
			parts = []string{
				m.hint("←→", "view"), m.hint("↑↓", "scroll"),
				m.hint("/", "search"), m.hint("n/N", "match"), m.hint("ctrl+k", "clear"),
			}
		} else {
			parts = []string{
				m.hint("ctrl+s", "send"), m.hint("tab", "focus"),
				m.hint("ctrl+e", "section"), m.hint("ctrl+y", "curl"), m.hint("←→", "tabs"),
			}
		}
	case tabCollections:
		parts = []string{m.hint("n", "new"), m.hint("↵", "open/load"), m.hint("d", "del"), m.hint("esc", "back")}
	case tabHistory:
		parts = []string{m.hint("↵", "reuse"), m.hint("d", "del"), m.hint("esc", "back")}
	case tabEnvironments:
		if m.envEdit != nil {
			parts = []string{m.hint("tab", "field"), m.hint("↵", "add var"), m.hint("ctrl+s", "save"), m.hint("esc", "cancel")}
		} else {
			parts = []string{m.hint("n", "new"), m.hint("a", "activate"), m.hint("↵", "edit"), m.hint("d", "del")}
		}
	case tabConfig:
		parts = []string{m.hint("tab", "section"), m.hint("↑↓", "nav"), m.hint("↵", "apply")}
	case tabHelp:
		parts = []string{m.hint("esc", "back"), m.hint("q", "quit")}
	}
	return strings.Join(parts, m.styles.footerText.Render("   "))
}

// renderBody dispatches to the active tab, wrapping non-request tabs in a
// single titled panel.
func (m model) renderBody(w, h int) string {
	if m.activeTab == tabRequest {
		return m.viewRequest(w, h)
	}

	var title, content string
	innerW, innerH := w-4, h-3
	switch m.activeTab {
	case tabCollections:
		title, content = "Collections", m.viewCollections(innerW, innerH)
	case tabHistory:
		title, content = "History", m.viewHistory(innerW, innerH)
	case tabEnvironments:
		title, content = "Environments", m.viewEnvironments(innerW, innerH)
	case tabConfig:
		title, content = "Config", m.viewConfig(innerW, innerH)
	case tabHelp:
		title, content = "Keyboard Shortcuts", m.viewHelp(innerW, innerH)
	}
	return m.panelBox(title, content, w, h, false)
}

// panelBox renders content inside a rounded, optionally-focused border with a
// bold title line. Width/height are the *outer* dimensions.
func (m model) panelBox(title, content string, w, h int, focused bool) string {
	s := m.styles
	st := s.panel
	if focused {
		st = s.panelFocus
	}

	// lipgloss Width includes horizontal padding (2) but not the border (2),
	// so Width(w-2) yields a w-wide box whose text area is w-4 — matching the
	// width every content renderer is built against.
	innerW := max(w-2, 1)
	innerH := max(h-2, 1)

	var b strings.Builder
	if title != "" {
		b.WriteString(s.panelLabel.Render(title))
		b.WriteByte('\n')
	}
	b.WriteString(content)

	body := clipHeight(b.String(), innerH)
	return st.Width(innerW).Height(innerH).Render(body)
}

// ---------------------------------------------------------------------------
// Request tab — responsive
// ---------------------------------------------------------------------------

func (m model) viewRequest(w, h int) string {
	url := m.renderURLBar(w)
	urlH := lipgloss.Height(url)
	rest := h - urlH

	// Too short for two panels: show only the one the user is working in.
	if rest < 8 {
		panel := m.renderEditorPanel(w, max(rest, 3))
		if m.focusArea == focusResponse {
			panel = m.renderResponsePanel(w, max(rest, 3))
		}
		return clipHeight(lipgloss.JoinVertical(lipgloss.Left, url, panel), h)
	}

	if m.layoutMode() == layoutWide {
		leftW := w * 48 / 100
		rightW := w - leftW
		left := m.renderEditorPanel(leftW, rest)
		right := m.renderResponsePanel(rightW, rest)
		cols := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
		return clipHeight(lipgloss.JoinVertical(lipgloss.Left, url, cols), h)
	}

	editorH := rest / 2
	respH := rest - editorH
	editor := m.renderEditorPanel(w, editorH)
	resp := m.renderResponsePanel(w, respH)
	return clipHeight(lipgloss.JoinVertical(lipgloss.Left, url, editor, resp), h)
}

func (m model) renderURLBar(w int) string {
	s := m.styles
	method := methods[m.methodIdx]
	badge := methodStyle(s, method).Render(fmt.Sprintf(" %s ▾ ", method))

	buttons := m.renderActionButtons(w >= 80)

	innerW := w - 4
	urlW := innerW - lipgloss.Width(badge) - lipgloss.Width(buttons) - 3
	if urlW < 8 {
		urlW = 8
	}

	urlStyle := s.urlInput
	if m.focusArea == focusURL {
		urlStyle = s.urlInputFocus
	}
	m.urlInput.Width = urlW - 2
	urlField := urlStyle.Width(urlW).Render(m.urlInput.View())

	content := lipgloss.JoinHorizontal(lipgloss.Center, badge, " ", urlField, " ", buttons)
	focused := m.focusArea == focusURL || m.focusArea == focusActions
	return m.panelBox("", content, w, 3, focused)
}

func (m model) renderActionButtons(full bool) string {
	s := m.styles
	labels := []string{"▶ Send", "Save", "cURL"}
	if !full {
		labels = []string{"▶ Send"}
	}

	var out []string
	for i, lb := range labels {
		st := s.btn
		switch {
		case m.focusArea == focusActions && i == m.actionIdx && i == 0:
			st = s.btnSendActive
		case m.focusArea == focusActions && i == m.actionIdx:
			st = s.btnActive
		case i == 0:
			st = s.btnSend
		}
		out = append(out, st.Render(" "+lb+" "))
	}
	return strings.Join(out, " ")
}

func (m model) renderEditorPanel(w, h int) string {
	s := m.styles
	innerW := w - 4
	focused := m.focusArea == focusSubTabs || m.focusArea == focusEditor

	subtabs := m.renderSubTabs()
	divider := s.separator.Render(strings.Repeat("─", max(innerW, 1)))

	contentH := max(h-2-1-1-1, 1) // border, title, subtabs, divider

	var body string
	switch m.subTab {
	case subTabParams:
		body = m.renderTableEditor(m.params, innerW, contentH)
	case subTabHeaders:
		body = m.renderTableEditor(m.headers, innerW, contentH)
	case subTabAuth:
		body = m.renderAuthEditor(innerW, contentH)
	case subTabBody:
		body = m.renderBodyEditor(innerW, contentH)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, subtabs, divider, body)
	return m.panelBox("Request", content, w, h, focused)
}

func (m model) renderSubTabs() string {
	s := m.styles
	var items []string
	for i, name := range subTabNames {
		label := " " + name + " "
		if subTab(i) == m.subTab {
			items = append(items, s.subTabActive.Render(label))
		} else {
			items = append(items, s.subTabInactive.Render(label))
		}
	}
	sep := s.muted.Render("│")
	row := strings.Join(items, sep)

	marker := "  "
	if m.focusArea == focusSubTabs {
		marker = s.accent.Render("▸ ")
	}
	return marker + row
}

func (m model) renderTableEditor(kvs []keyValue, w, h int) string {
	s := m.styles
	colKeyW := (w - 5) / 2
	colValW := w - colKeyW - 5
	if colKeyW < 3 {
		colKeyW = 3
	}
	if colValW < 3 {
		colValW = 3
	}

	header := " " + s.editorHeader.Render(padRight("Key", colKeyW)) + "  " + s.editorHeader.Render(padRight("Value", colValW))
	rows := []string{header}

	// Window the rows around the selected one to fit the available height.
	maxRows := max(h-1, 1)
	start := 0
	if m.editorRow >= maxRows {
		start = m.editorRow - maxRows + 1
	}
	end := min(start+maxRows, len(kvs))

	for i := start; i < end; i++ {
		kv := kvs[i]
		marker := "  "
		active := i == m.editorRow && m.focusArea == focusEditor
		if active {
			marker = s.accent.Render("▸ ")
		}

		keyStr := kv.Key.Value()
		if keyStr == "" {
			keyStr = s.muted.Render("—")
		}
		valStr := kv.Value.Value()
		if valStr == "" {
			valStr = s.muted.Render("—")
		}
		row := marker + clipLine(padRight(keyStr, colKeyW), colKeyW) + "  " + clipLine(valStr, colValW)
		if active {
			row = s.editorRowActive.Render(row)
		}
		rows = append(rows, row)
	}

	if len(kvs) == 0 {
		rows = append(rows, s.muted.Render("  (empty) — enter to add a row"))
	}
	return strings.Join(rows, "\n")
}

func (m model) renderAuthEditor(w, h int) string {
	s := m.styles
	title := s.accent.Render(authTypeNames[m.authType]) + s.muted.Render("   ↑/↓ change type")

	var fields []string
	switch m.authType {
	case authBearer:
		fields = append(fields, "  Token  "+m.authValue.View())
	case authBasic, authDigest:
		fields = append(fields, "  User   "+m.authUser.View())
		fields = append(fields, "  Pass   "+m.authPass.View())
	case authAPIKey:
		fields = append(fields, "  Key    "+m.authKey.View())
		fields = append(fields, "  Value  "+m.authValue.View())
	case authOAuth2:
		fields = append(fields, "  "+s.muted.Render("OAuth2 client-credentials — set token endpoint in Body"))
	default:
		fields = append(fields, "  "+s.muted.Render("no authentication"))
	}
	return title + "\n\n" + strings.Join(fields, "\n")
}

func (m model) renderBodyEditor(w, h int) string {
	s := m.styles
	title := s.accent.Render(bodyTypeNames[m.bodyType]) + s.muted.Render("   ctrl+b change type")

	if m.bodyType == bodyNone {
		return title + "\n\n  " + s.muted.Render("no body for this request")
	}

	m.bodyInput.SetWidth(max(w-2, 8))
	m.bodyInput.SetHeight(max(h-2, 2))
	return title + "\n" + m.bodyInput.View()
}

func (m model) renderResponsePanel(w, h int) string {
	focused := m.focusArea == focusResponse
	innerW := w - 4
	innerH := max(h-2-1, 1) // border + title
	content := m.responseView(innerW, innerH)
	return m.panelBox("Response", content, w, h, focused)
}

func (m model) responseView(w, h int) string {
	s := m.styles

	if m.response == nil {
		if m.sending {
			return "\n  " + s.accent.Render(m.spinner.View()) + s.muted.Render(" sending request...")
		}
		return "\n  " + s.muted.Render("no response yet — press ") + s.footerKey.Render("ctrl+s") + s.muted.Render(" to send")
	}

	r := m.response
	if r.Error != "" {
		return "\n  " + s.warn.Render("✗ "+r.Error)
	}

	// Status header.
	statusStyle := s.statusOK
	switch {
	case r.StatusCode >= 400:
		statusStyle = s.statusErr
	case r.StatusCode >= 300:
		statusStyle = s.statusWarn
	}
	statusText := strings.TrimSpace(strings.TrimPrefix(r.StatusText, fmt.Sprintf("%d", r.StatusCode)))
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		statusStyle.Render(fmt.Sprintf(" %d ", r.StatusCode)),
		" ", s.body.Render(statusText),
		s.muted.Render(fmt.Sprintf("  ·  %dms  ·  %s", r.Duration.Milliseconds(), FormatBodySize(r.BodySize))),
	)
	modeBar := m.renderRespModes()

	// Compute available body height.
	bodyH := h - 2 // header + mode bar
	searchLine := ""
	if m.searchMode {
		searchLine = s.searchBar.Render("/" + m.searchInput.View())
		bodyH--
	}
	bodyH-- // scroll indicator line
	if bodyH < 1 {
		bodyH = 1
	}

	lines := m.responseLines()
	total := len(lines)

	scroll := m.responseScroll
	if scroll > total-bodyH {
		scroll = total - bodyH
	}
	if scroll < 0 {
		scroll = 0
	}
	end := min(scroll+bodyH, total)

	query := strings.TrimSpace(m.searchInput.Value())
	var rendered []string
	for i := scroll; i < end; i++ {
		ln := clipLine(lines[i], w)
		if query != "" {
			ln = highlightMatch(ln, query, s)
		}
		rendered = append(rendered, ln)
	}
	bodyStr := strings.Join(rendered, "\n")

	scrollInfo := ""
	if total > bodyH {
		pct := 0
		if total-bodyH > 0 {
			pct = scroll * 100 / (total - bodyH)
		}
		scrollInfo = s.muted.Render(fmt.Sprintf("%d–%d / %d   %d%%", scroll+1, end, total, pct))
	}

	parts := []string{header, modeBar}
	if searchLine != "" {
		parts = append(parts, searchLine)
	}
	parts = append(parts, bodyStr)
	if scrollInfo != "" {
		parts = append(parts, scrollInfo)
	}
	return strings.Join(parts, "\n")
}

func (m model) renderRespModes() string {
	s := m.styles
	var items []string
	for i, name := range respModeNames {
		label := " " + name + " "
		if respMode(i) == m.responseMode {
			items = append(items, s.subTabActive.Render(label))
		} else {
			items = append(items, s.subTabInactive.Render(label))
		}
	}
	return strings.Join(items, s.muted.Render("│"))
}

// ---------------------------------------------------------------------------
// Collections / History / Environments / Config / Help
// ---------------------------------------------------------------------------

func (m model) viewCollections(w, h int) string {
	s := m.styles

	if m.creatingMode {
		return s.muted.Render("new collection name:") + "\n\n  " + s.accent.Render(m.creatingName.View()+"█") +
			"\n\n" + s.muted.Render("enter to create · esc to cancel")
	}
	if len(m.collections) == 0 {
		return s.muted.Render("no collections yet\n\npress ") + s.footerKey.Render("n") + s.muted.Render(" to create one")
	}

	var list []string
	for i, col := range m.collections {
		marker := "  "
		if i == m.collIdx {
			marker = s.accent.Render("▸ ")
		}
		line := marker + s.body.Render(col.Name) + s.muted.Render(fmt.Sprintf("  (%d)", len(col.Requests)))
		list = append(list, line)
	}

	out := strings.Join(list, "\n")

	if m.collMode == collDetail && m.collIdx < len(m.collections) {
		head := "\n" + s.sectionHead.Render("requests in "+m.collections[m.collIdx].Name)
		if len(m.collReqs) == 0 {
			out += head + "\n" + s.muted.Render("  (no requests)")
		} else {
			var reqLines []string
			for i, req := range m.collReqs {
				marker := "  "
				if i == m.collReqIdx {
					marker = s.accent.Render("▸ ")
				}
				reqLines = append(reqLines, marker+methodStyle(s, req.Method).Render(" "+req.Method+" ")+" "+s.muted.Render(shortURL(req.URL)))
			}
			out += head + "\n" + strings.Join(reqLines, "\n")
		}
	}
	return clipHeight(out, h)
}

func (m model) viewHistory(w, h int) string {
	s := m.styles
	if len(m.history) == 0 {
		return s.muted.Render("no history yet\n\nsent requests appear here automatically")
	}

	maxRows := max(h, 1)
	start := 0
	if m.histIdx >= maxRows {
		start = m.histIdx - maxRows + 1
	}
	end := min(start+maxRows, len(m.history))

	var list []string
	for i := start; i < end; i++ {
		entry := m.history[i]
		marker := "  "
		if i == m.histIdx {
			marker = s.accent.Render("▸ ")
		}
		status := s.statusOK
		switch {
		case entry.Response.StatusCode == 0 || entry.Response.StatusCode >= 400:
			status = s.statusErr
		case entry.Response.StatusCode >= 300:
			status = s.statusWarn
		}
		code := entry.Response.StatusCode
		codeStr := fmt.Sprintf("%d", code)
		if code == 0 {
			codeStr = "ERR"
		}
		line := marker + methodStyle(s, entry.Request.Method).Render(fmt.Sprintf(" %-4s ", entry.Request.Method)) +
			" " + status.Render(codeStr) + " " + s.muted.Render(shortURL(entry.Request.URL))
		list = append(list, line)
	}
	return strings.Join(list, "\n")
}

func (m model) viewEnvironments(w, h int) string {
	s := m.styles

	if m.creatingMode {
		return s.muted.Render("new environment name:") + "\n\n  " + s.accent.Render(m.creatingName.View()+"█") +
			"\n\n" + s.muted.Render("enter to create · esc to cancel")
	}

	if m.envEdit != nil {
		return m.viewEnvEditor(w, h)
	}

	if len(m.envs) == 0 {
		return s.muted.Render("no environments yet\n\npress ") + s.footerKey.Render("n") + s.muted.Render(" to create one")
	}

	var list []string
	for i, env := range m.envs {
		marker := "  "
		if i == m.envIdx {
			marker = s.accent.Render("▸ ")
		}
		active := ""
		if env.Name == m.activeEnv {
			active = " " + s.envBadge.Render("active")
		}
		list = append(list, marker+s.body.Render(env.Name)+s.muted.Render(fmt.Sprintf("  (%d vars)", len(env.Values)))+active)
	}
	return clipHeight(strings.Join(list, "\n"), h)
}

func (m model) viewEnvEditor(w, h int) string {
	s := m.styles
	e := m.envEdit
	head := s.sectionHead.Render("editing: "+e.Name.Value()) + "\n" +
		s.separator.Render(strings.Repeat("─", max(w, 1))) + "\n"

	colKeyW := (w - 7) / 2
	colValW := w - colKeyW - 7
	if colKeyW < 3 {
		colKeyW = 3
	}
	if colValW < 3 {
		colValW = 3
	}

	var rows []string
	rows = append(rows, " "+s.editorHeader.Render(padRight("Variable", colKeyW))+"  "+s.editorHeader.Render(padRight("Value", colValW)))

	for i, kv := range e.Vars {
		marker := "  "
		if i == e.Idx {
			marker = s.accent.Render("▸ ")
		}
		key := kv.Key.Value()
		if key == "" {
			key = s.muted.Render("—")
		}
		val := kv.Value.Value()
		if val == "" {
			val = s.muted.Render("—")
		}
		row := marker + clipLine(padRight(key, colKeyW), colKeyW) + "  " + clipLine(val, colValW)
		if i == e.Idx {
			row = s.editorRowActive.Render(row)
		}
		rows = append(rows, row)
	}
	return clipHeight(head+strings.Join(rows, "\n"), h)
}

func (m model) viewConfig(w, h int) string {
	s := m.styles

	sections := []string{"theme", "color", "mode", "user"}
	var secParts []string
	for i, name := range sections {
		if i == m.configSection {
			secParts = append(secParts, s.accent.Bold(true).Render(name))
		} else {
			secParts = append(secParts, s.muted.Render(name))
		}
	}
	secBar := strings.Join(secParts, s.muted.Render("  ·  "))

	var items []string
	visible := max(h-4, 4)

	switch m.configSection {
	case 0:
		all := m.allThemes()
		start := 0
		if m.configCursor >= visible {
			start = m.configCursor - visible + 1
		}
		end := min(start+visible, len(all))
		for i := start; i < end; i++ {
			var t theme
			if i < len(themes) {
				t = themes[i]
			} else {
				t = m.customThemes[i-len(themes)]
			}
			marker := "  "
			if i == m.configCursor {
				marker = s.themeCursor.Render("▸ ")
			}
			swatch := lipgloss.NewStyle().Background(t.primary).Render("  ") +
				lipgloss.NewStyle().Background(t.accent).Render("  ")
			active := ""
			if i == m.themeIdx {
				active = s.muted.Render("  (active)")
			}
			star := ""
			if i >= len(themes) {
				star = s.accent.Render(" ★")
			}
			items = append(items, marker+swatch+" "+s.body.Render(t.name)+active+star)
		}
		items = append(items, "", s.muted.Render("i import · e export · d delete"))

	case 1:
		start := m.colorScroll
		end := min(start+visible, len(colorSchemes))
		for i := start; i < end; i++ {
			cs := colorSchemes[i]
			marker := "  "
			if i == m.configCursor {
				marker = s.configCursor.Render("▸ ")
			}
			dot := lipgloss.NewStyle().Background(cs.primary).Render("  ") +
				lipgloss.NewStyle().Background(cs.accent).Render("  ")
			active := ""
			if i == m.colorIdx {
				active = s.muted.Render("  (active)")
			}
			items = append(items, marker+dot+" "+s.body.Render(cs.name)+active)
		}

	case 2:
		for i, name := range []string{"dark", "light"} {
			marker := "  "
			if i == m.configCursor {
				marker = s.themeCursor.Render("▸ ")
			}
			active := ""
			if i == m.modeIdx {
				active = s.muted.Render("  (active)")
			}
			items = append(items, marker+s.body.Render(name)+active)
		}

	case 3:
		items = append(items,
			s.sectionHead.Render("config path"), s.muted.Render("  "+getConfigPath()), "",
			s.sectionHead.Render("themes path"), s.muted.Render("  "+getThemesDir()), "",
			s.sectionHead.Render("history limit"), s.muted.Render(fmt.Sprintf("  %d entries", m.config.HistoryLimit)), "",
			s.sectionHead.Render("actions"),
		)
		marker := "  "
		if m.configCursor == 0 {
			marker = s.themeCursor.Render("▸ ")
		}
		items = append(items, marker+s.body.Render("reset all settings"))
	}

	out := secBar + "\n\n" + strings.Join(items, "\n")
	return clipHeight(out, h)
}

func (m model) viewHelp(w, h int) string {
	s := m.styles
	row := func(key, desc string) string {
		return s.keyLabel.Render(key) + s.keyDesc.Render(desc)
	}

	cols := [][]string{
		{
			s.sectionHead.Render("navigation"),
			row("  1-6", "switch tab"),
			row("  ←/→ h/l", "prev / next tab"),
			row("  tab", "next focus area"),
			row("  shift+tab", "prev focus area"),
			row("  ?", "toggle help"),
			row("  q / esc", "quit"),
		},
		{
			s.sectionHead.Render("request"),
			row("  ctrl+s", "send request"),
			row("  ctrl+e", "cycle sections"),
			row("  ctrl+b", "cycle body type"),
			row("  ctrl+n", "new request"),
			row("  ctrl+d", "duplicate request"),
			row("  ctrl+y", "copy as cURL"),
			row("  ctrl+k", "clear response"),
		},
		{
			s.sectionHead.Render("response"),
			row("  ↑/↓", "scroll body"),
			row("  ←/→", "switch view mode"),
			row("  /", "search in body"),
			row("  n / N", "next / prev match"),
			"",
			s.sectionHead.Render("collections / envs"),
			row("  n / d", "new / delete"),
			row("  a", "activate environment"),
			row("  enter", "open / load / edit"),
		},
	}

	colW := max((w-2)/len(cols), 18)
	var rendered []string
	for _, c := range cols {
		rendered = append(rendered, lipgloss.NewStyle().Width(colW).Render(strings.Join(c, "\n")))
	}
	return clipHeight(lipgloss.JoinHorizontal(lipgloss.Top, rendered...), h)
}

func (m model) viewImport() string {
	s := m.styles
	title := s.title.Render("import theme")
	sep := s.separator.Render(strings.Repeat("─", min(40, m.width-4)))
	hint := s.muted.Render("paste theme JSON and press enter (esc to cancel)")
	input := s.accent.Render(m.importBuf + "█")
	return lipgloss.NewStyle().Padding(1, 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, sep, "", hint, "", input),
	)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// padRight pads or truncates s to exactly w display columns (plain text only).
func padRight(s string, w int) string {
	width := lipgloss.Width(s)
	if width >= w {
		if w > 1 {
			return ansi.Truncate(s, w-1, "…")
		}
		return ansi.Truncate(s, w, "")
	}
	return s + strings.Repeat(" ", w-width)
}

// clipLine truncates a (possibly styled) line to w display columns, ANSI-aware.
func clipLine(s string, w int) string {
	if w < 1 {
		return ""
	}
	return ansi.Truncate(s, w, "")
}

// clipHeight keeps at most n lines, preventing vertical overflow of a panel.
func clipHeight(s string, n int) string {
	if n < 1 {
		n = 1
	}
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// highlightMatch wraps each case-insensitive occurrence of query in line with
// the match-highlight style. Operates on plain text.
func highlightMatch(line, query string, s *styles) string {
	if query == "" {
		return line
	}
	lower := strings.ToLower(line)
	ql := strings.ToLower(query)

	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lower[i:], ql)
		if j < 0 {
			b.WriteString(line[i:])
			break
		}
		start := i + j
		stop := start + len(query)
		b.WriteString(line[i:start])
		b.WriteString(s.matchHi.Render(line[start:stop]))
		i = stop
	}
	return b.String()
}

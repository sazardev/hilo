package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case responseMsg:
		m.response = &msg.response
		m.sending = false
		m.lastError = msg.response.Error
		m.responseScroll = 0
		m.lastSentReq = msg.request
		m.searchMode = false
		m.searchMatches = nil
		m.focusArea = focusResponse
		m.focusCurrent()

		// Persist every executed request to history and prune to the limit.
		entry := HistoryEntry{Request: msg.request, Response: msg.response}
		_ = SaveHistoryEntry(entry)
		PruneHistory(m.config.HistoryLimit)
		m.history, _ = ListHistory()
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.importMode {
		return m.handleImportKey(msg)
	}
	if m.searchMode {
		return m.handleSearchKey(msg)
	}

	k := msg.String()
	typing := m.isTyping()

	// ctrl+c always quits, no matter the context.
	if k == "ctrl+c" {
		return m, tea.Quit
	}

	// esc: contextual back / close.
	if k == "esc" {
		if m.activeTab == tabHelp {
			m.activeTab = tabRequest
			return m, nil
		}
		if m.activeTab == tabRequest && m.focusArea != focusURL {
			m.focusArea = focusURL
			m.focusCurrent()
			return m, nil
		}
		// Delegate to per-tab handlers so they can close their own panels.
		if m.activeTab != tabRequest {
			return m.routeToTab(msg)
		}
		return m, nil
	}

	// Tab / Shift+Tab navigate within a tab's own focus model.
	if k == "tab" || k == "shift+tab" {
		return m.routeToTab(msg)
	}

	// While typing in a text input, route everything to the active tab so the
	// input receives the keystroke. Only true control chords are intercepted.
	if typing {
		return m.routeToTab(msg)
	}

	// Global navigation — only when NOT typing in an input.
	switch k {
	case "q":
		return m, tea.Quit

	case "1", "2", "3", "4", "5":
		idx := int(k[0] - '1')
		if idx < len(tabNames) {
			m.switchTab(tab(idx))
		}
		return m, nil

	case "left", "h":
		if m.activeTab != tabHelp {
			t := m.activeTab - 1
			if t < 0 {
				t = tab(len(tabNames) - 1)
			}
			m.switchTab(t)
		}
		return m, nil

	case "right", "l":
		if m.activeTab != tabHelp {
			t := m.activeTab + 1
			if int(t) >= len(tabNames) {
				t = 0
			}
			m.switchTab(t)
		}
		return m, nil

	case "?":
		if m.activeTab == tabHelp {
			m.activeTab = tabRequest
		} else {
			m.activeTab = tabHelp
		}
		return m, nil
	}

	return m.routeToTab(msg)
}

// routeToTab dispatches a key to the handler for the active tab.
func (m model) routeToTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.activeTab {
	case tabRequest:
		return m.handleRequestKey(msg)
	case tabCollections:
		return m.handleCollectionsKey(msg)
	case tabHistory:
		return m.handleHistoryKey(msg)
	case tabEnvironments:
		return m.handleEnvironmentsKey(msg)
	case tabConfig:
		return m.handleConfigKey(msg)
	}
	return m, nil
}

// switchTab activates a tab, resets request focus and refreshes its data.
func (m *model) switchTab(t tab) {
	m.activeTab = t
	m.focusArea = focusURL
	m.loadTabData()
	m.focusCurrent()
}

func (m model) handleRequestKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	switch k {
	case "ctrl+s":
		if !m.sending && m.urlInput.Value() != "" {
			m.sending = true
			return m, m.sendRequest()
		}
		return m, nil

	case "ctrl+e":
		m.subTab = (m.subTab + 1) % subTab(len(subTabNames))
		m.focusArea = focusSubTabs
		m.editorRow = 0
		m.editorCol = 0
		m.focusCurrent()
		return m, nil

	case "ctrl+n":
		m.urlInput.SetValue("")
		m.bodyInput.SetValue("")
		m.authKey.SetValue("")
		m.authValue.SetValue("")
		m.authUser.SetValue("")
		m.authPass.SetValue("")
		m.params = []keyValue{newKeyValue("", "")}
		m.headers = []keyValue{newKeyValue("", "")}
		m.methodIdx = 0
		m.authType = authNone
		m.bodyType = bodyNone
		m.response = nil
		m.editorRow = 0
		m.editorCol = 0
		m.message = "new request"
		m.messageTime = timeNow()
		m.focusCurrent()
		return m, nil

	case "ctrl+d":
		return m.duplicateRequest()

	case "ctrl+b":
		// Cycle the body type (works from anywhere in the Request tab).
		m.bodyType = (m.bodyType + 1) % bodyTypeIdx(len(bodyTypeNames))
		if m.subTab != subTabBody {
			m.subTab = subTabBody
			m.focusArea = focusSubTabs
		}
		m.message = "body: " + bodyTypeNames[m.bodyType]
		m.messageTime = timeNow()
		return m, nil

	case "ctrl+y":
		return m.copyCurl()

	case "ctrl+k":
		m.response = nil
		m.searchMode = false
		m.searchMatches = nil
		return m, nil

	case "/":
		if m.focusArea == focusResponse && m.response != nil {
			m.searchMode = true
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			return m, nil
		}

	case "tab":
		return m.requestFocusNext(), nil

	case "shift+tab":
		return m.requestFocusPrev(), nil

	case "up", "k":
		switch m.focusArea {
		case focusURL:
			m.methodIdx = (m.methodIdx - 1 + len(methods)) % len(methods)
		case focusActions:
			m.focusArea = focusURL
			m.focusCurrent()
		case focusSubTabs:
			m.focusArea = focusActions
		case focusEditor:
			return m.handleEditorVertical(msg)
		case focusResponse:
			m.scrollResponse(-1)
		}
		return m, nil

	case "down", "j":
		switch m.focusArea {
		case focusURL:
			m.methodIdx = (m.methodIdx + 1) % len(methods)
		case focusActions:
			m.focusArea = focusSubTabs
		case focusSubTabs:
			m.focusArea = focusEditor
			m.editorRow, m.editorCol = 0, 0
			m.focusCurrent()
		case focusEditor:
			return m.handleEditorVertical(msg)
		case focusResponse:
			m.scrollResponse(1)
		}
		return m, nil

	case "pgup":
		if m.focusArea == focusResponse {
			m.scrollResponse(-m.responsePageSize())
		}
		return m, nil

	case "pgdown":
		if m.focusArea == focusResponse {
			m.scrollResponse(m.responsePageSize())
		}
		return m, nil

	case "home", "g":
		if m.focusArea == focusResponse {
			m.responseScroll = 0
			return m, nil
		}

	case "end", "G":
		if m.focusArea == focusResponse {
			m.responseScroll = max(0, m.responseLineCount()-1)
			return m, nil
		}

	case "left":
		switch {
		case m.focusArea == focusActions:
			if m.actionIdx > 0 {
				m.actionIdx--
			}
			return m, nil
		case m.focusArea == focusSubTabs:
			m.cycleSubTab(-1)
			return m, nil
		case m.focusArea == focusResponse:
			m.responseMode = (m.responseMode - 1 + respMode(len(respModeNames))) % respMode(len(respModeNames))
			m.responseScroll = 0
			m.searchMatches = nil
			return m, nil
		case m.focusArea == focusEditor && (m.subTab == subTabParams || m.subTab == subTabHeaders):
			return m.handleTableEditorKey(msg)
		}
		return m, nil

	case "right":
		switch {
		case m.focusArea == focusActions:
			if m.actionIdx < 2 {
				m.actionIdx++
			}
			return m, nil
		case m.focusArea == focusSubTabs:
			m.cycleSubTab(1)
			return m, nil
		case m.focusArea == focusResponse:
			m.responseMode = (m.responseMode + 1) % respMode(len(respModeNames))
			m.responseScroll = 0
			m.searchMatches = nil
			return m, nil
		case m.focusArea == focusEditor && (m.subTab == subTabParams || m.subTab == subTabHeaders):
			return m.handleTableEditorKey(msg)
		}
		return m, nil

	case "n", "N":
		if m.focusArea == focusResponse && len(m.searchMatches) > 0 {
			return m.jumpSearch(k == "n")
		}

	case "y":
		if m.focusArea == focusResponse && m.response != nil {
			return m.copyResponseBody()
		}

	case "enter":
		if m.focusArea == focusSubTabs {
			m.focusArea = focusEditor
			m.focusCurrent()
			return m, nil
		}
		if m.focusArea == focusActions {
			return m.runAction()
		}
		if m.focusArea == focusEditor {
			if m.subTab == subTabBody {
				var cmd tea.Cmd
				m.bodyInput, cmd = m.bodyInput.Update(msg)
				return m, cmd
			}
			return m.handleEditorEnter()
		}
		return m, nil

	case "backspace":
		if m.focusArea == focusEditor {
			if m.subTab == subTabBody {
				var cmd tea.Cmd
				m.bodyInput, cmd = m.bodyInput.Update(msg)
				return m, cmd
			}
			if m.subTab == subTabParams || m.subTab == subTabHeaders {
				return m.handleEditorBackspace()
			}
		}
	}

	if m.focusArea == focusURL {
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}

	if m.focusArea == focusEditor {
		return m.handleEditorKey(msg)
	}

	return m, nil
}

func (m model) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.subTab {
	case subTabParams, subTabHeaders:
		return m.handleTableEditorKey(msg)
	case subTabAuth:
		return m.handleAuthKey(msg)
	case subTabBody:
		var cmd tea.Cmd
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

// requestFocusNext advances focus through the Request tab in reading order:
// URL → actions → section selector → each editor field → response → URL.
func (m model) requestFocusNext() model {
	switch m.focusArea {
	case focusURL:
		m.focusArea = focusActions
	case focusActions:
		m.focusArea = focusSubTabs
	case focusSubTabs:
		m.focusArea = focusEditor
		m.editorRow, m.editorCol = 0, 0
	case focusEditor:
		if m.editorTabForward() {
			return m // moved to the next field within the editor
		}
		m.focusArea = focusResponse
	case focusResponse:
		m.focusArea = focusURL
	}
	m.focusCurrent()
	return m
}

func (m model) requestFocusPrev() model {
	switch m.focusArea {
	case focusURL:
		m.focusArea = focusResponse
	case focusActions:
		m.focusArea = focusURL
	case focusSubTabs:
		m.focusArea = focusActions
	case focusEditor:
		if m.editorTabBackward() {
			return m
		}
		m.focusArea = focusSubTabs
	case focusResponse:
		m.focusArea = focusEditor
	}
	m.focusCurrent()
	return m
}

// editorTabForward moves to the next field inside the editor, returning false
// when already at the last field so the caller advances out of the editor.
func (m *model) editorTabForward() bool {
	switch m.subTab {
	case subTabParams, subTabHeaders:
		kvs := m.params
		if m.subTab == subTabHeaders {
			kvs = m.headers
		}
		if m.editorCol == 0 {
			m.editorCol = 1
			m.focusCurrent()
			return true
		}
		if m.editorRow < len(kvs)-1 {
			m.editorRow++
			m.editorCol = 0
			m.focusCurrent()
			return true
		}
		return false
	case subTabAuth:
		return m.focusNextAuthField()
	default:
		return false
	}
}

func (m *model) editorTabBackward() bool {
	switch m.subTab {
	case subTabParams, subTabHeaders:
		if m.editorCol == 1 {
			m.editorCol = 0
			m.focusCurrent()
			return true
		}
		if m.editorRow > 0 {
			m.editorRow--
			m.editorCol = 1
			m.focusCurrent()
			return true
		}
		return false
	case subTabAuth:
		return m.focusPrevAuthField()
	default:
		return false
	}
}

// handleEditorVertical routes up/down within the editor based on the sub-tab:
// the body textarea moves its cursor, the auth panel cycles auth type, and the
// param/header tables move the selected row.
func (m model) handleEditorVertical(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.subTab {
	case subTabBody:
		var cmd tea.Cmd
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		return m, cmd
	case subTabAuth:
		return m.handleAuthKey(msg)
	default:
		return m.handleTableEditorKey(msg)
	}
}

// runAction performs the focused action button (Send / Save / Copy cURL).
func (m model) runAction() (tea.Model, tea.Cmd) {
	switch m.actionIdx {
	case 0:
		if !m.sending && m.urlInput.Value() != "" {
			m.sending = true
			return m, m.sendRequest()
		}
		return m, nil
	case 1:
		return m.saveToCollection(false)
	case 2:
		return m.copyCurl()
	}
	return m, nil
}

// saveToCollection persists the current editor request into the selected
// collection and auto-commits the change to its git repo. When duplicate is
// true the saved request name gets a "(copy)" suffix.
func (m model) saveToCollection(duplicate bool) (tea.Model, tea.Cmd) {
	if m.urlInput.Value() == "" {
		m.message = "nothing to save"
		m.messageTime = timeNow()
		return m, nil
	}
	if len(m.collections) == 0 {
		m.message = "create a collection first (tab Collections, n)"
		m.messageTime = timeNow()
		return m, nil
	}
	if m.collIdx < 0 || m.collIdx >= len(m.collections) {
		m.collIdx = 0
	}

	req := m.currentRequest()
	req.ID = generateID()
	if duplicate {
		req.Name += " (copy)"
	}
	col := m.collections[m.collIdx]
	req.Collection = col.Name

	commitType := CommitAdd
	if repo, err := OpenCollectionRepo(col.Name); err == nil {
		// SaveRequest on the repo writes the file and commits atomically.
		_ = repo.SaveRequest(req, commitType)
	} else {
		_ = SaveRequest(req)
	}

	col.Requests = append(col.Requests, req.ID)
	_ = SaveCollection(col)
	m.collections[m.collIdx] = col

	verb := "saved to"
	if duplicate {
		verb = "duplicated into"
	}
	m.message = fmt.Sprintf("%s '%s'", verb, col.Name)
	m.messageTime = timeNow()
	return m, nil
}

// duplicateRequest clears the response and stores a copy of the current
// request in the active collection so it can be tweaked and re-sent.
func (m model) duplicateRequest() (tea.Model, tea.Cmd) {
	m.response = nil
	if len(m.collections) == 0 {
		m.message = "duplicated in editor (no collection to save into)"
		m.messageTime = timeNow()
		return m, nil
	}
	return m.saveToCollection(true)
}

// copyCurl renders the current request as a cURL command and copies it to the
// system clipboard, with environment variables resolved.
func (m model) copyCurl() (tea.Model, tea.Cmd) {
	curl := GenerateCurl(m.currentRequest(), m.resolveEnv())
	if err := clipboard.WriteAll(curl); err != nil {
		m.message = "clipboard unavailable — cURL not copied"
	} else {
		m.message = "cURL copied to clipboard"
	}
	m.messageTime = timeNow()
	return m, nil
}

// handleSearchKey drives the response search input overlay.
func (m model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchInput.Blur()
		return m, nil
	case "enter":
		m.searchMode = false
		m.searchInput.Blur()
		m.computeSearchMatches()
		if len(m.searchMatches) > 0 {
			m.searchIdx = 0
			m.responseScroll = m.searchMatches[0]
			m.message = fmt.Sprintf("%d match(es) — n/N to navigate", len(m.searchMatches))
		} else {
			m.message = "no matches"
		}
		m.messageTime = timeNow()
		return m, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
}

// jumpSearch advances the response scroll to the next/previous search match.
func (m model) jumpSearch(forward bool) (tea.Model, tea.Cmd) {
	n := len(m.searchMatches)
	if n == 0 {
		return m, nil
	}
	if forward {
		m.searchIdx = (m.searchIdx + 1) % n
	} else {
		m.searchIdx = (m.searchIdx - 1 + n) % n
	}
	m.responseScroll = m.searchMatches[m.searchIdx]
	m.message = fmt.Sprintf("match %d/%d", m.searchIdx+1, n)
	m.messageTime = timeNow()
	return m, nil
}

// computeSearchMatches records the line indices of the current response body
// (in the active view mode) that contain the search query, case-insensitive.
func (m *model) computeSearchMatches() {
	m.searchMatches = nil
	q := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if q == "" || m.response == nil {
		return
	}
	for i, ln := range m.responseLines() {
		if strings.Contains(strings.ToLower(ln), q) {
			m.searchMatches = append(m.searchMatches, i)
		}
	}
}

func (m model) handleTableEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	kvs := &m.params
	if m.subTab == subTabHeaders {
		kvs = &m.headers
	}

	switch msg.String() {
	case "up", "k":
		if m.editorRow > 0 {
			m.editorRow--
		}
		m.focusCurrent()
		return m, nil

	case "down", "j":
		if m.editorRow < len(*kvs)-1 {
			m.editorRow++
		}
		m.focusCurrent()
		return m, nil

	case "left":
		if m.editorCol > 0 {
			m.editorCol--
			m.focusCurrent()
		}
		return m, nil

	case "right":
		if m.editorCol < 1 {
			m.editorCol++
			m.focusCurrent()
		}
		return m, nil

	default:
		if m.editorRow < len(*kvs) {
			kv := &(*kvs)[m.editorRow]
			var cmd tea.Cmd
			if m.editorCol == 0 {
				kv.Key, cmd = kv.Key.Update(msg)
			} else {
				kv.Value, cmd = kv.Value.Update(msg)
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m model) handleAuthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.authType--
		if m.authType < authNone {
			m.authType = authOAuth2
		}
		m.focusCurrent()
		return m, nil

	case "down", "j":
		m.authType++
		if m.authType > authOAuth2 {
			m.authType = authNone
		}
		m.focusCurrent()
		return m, nil
	}

	// tab / shift+tab between auth fields are handled upstream in
	// handleRequestKey via focusNextAuthField / focusPrevAuthField.
	var cmd tea.Cmd
	switch m.authType {
	case authBearer:
		m.authValue, cmd = m.authValue.Update(msg)
	case authBasic, authDigest:
		if m.authPass.Focused() {
			m.authPass, cmd = m.authPass.Update(msg)
		} else {
			m.authUser, cmd = m.authUser.Update(msg)
		}
	case authAPIKey:
		if m.authValue.Focused() {
			m.authValue, cmd = m.authValue.Update(msg)
		} else {
			m.authKey, cmd = m.authKey.Update(msg)
		}
	}
	return m, cmd
}

func (m model) handleEditorEnter() (tea.Model, tea.Cmd) {
	kvs := &m.params
	if m.subTab == subTabHeaders {
		kvs = &m.headers
	}

	if m.subTab == subTabParams || m.subTab == subTabHeaders {
		newKV := newKeyValue("", "")
		*kvs = append(*kvs, newKV)
		if m.editorRow < len(*kvs)-1 {
			m.editorRow++
		}
		m.editorCol = 0
		m.focusCurrent()
	}
	return m, nil
}

func (m model) handleEditorBackspace() (tea.Model, tea.Cmd) {
	kvs := &m.params
	if m.subTab == subTabHeaders {
		kvs = &m.headers
	}

	if len(*kvs) <= 1 {
		return m, nil
	}

	kv := (*kvs)[m.editorRow]
	if kv.Key.Value() == "" && kv.Value.Value() == "" {
		*kvs = append((*kvs)[:m.editorRow], (*kvs)[m.editorRow+1:]...)
		if m.editorRow >= len(*kvs) {
			m.editorRow = max(0, len(*kvs)-1)
		}
		m.focusCurrent()
	}
	return m, nil
}

func (m model) handleCollectionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.creatingMode {
		switch msg.String() {
		case "esc":
			m.creatingMode = false
			m.creatingName.SetValue("")
			return m, nil
		case "enter":
			name := m.creatingName.Value()
			if name != "" {
				col := NewCollection(name, "")
				SaveCollection(col)
				InitCollectionRepo(name)
				m.collections, _ = ListCollections()
				m.message = fmt.Sprintf("created collection '%s'", name)
				m.messageTime = timeNow()
			}
			m.creatingMode = false
			m.creatingName.SetValue("")
			return m, nil
		default:
			var cmd tea.Cmd
			m.creatingName, cmd = m.creatingName.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "n":
		m.creatingMode = true
		m.creatingName.SetValue("")
		m.creatingName.Focus()
		return m, nil
	case "d", "delete":
		if len(m.collections) > 0 && m.collIdx < len(m.collections) {
			name := m.collections[m.collIdx].Name
			DeleteCollection(name)
			m.collections, _ = ListCollections()
			if m.collIdx >= len(m.collections) {
				m.collIdx = max(0, len(m.collections)-1)
			}
			m.collReqs = nil
			m.message = fmt.Sprintf("deleted collection '%s'", name)
			m.messageTime = timeNow()
		}
		return m, nil
	case "up", "k":
		if m.collIdx > 0 {
			m.collIdx--
			m.loadCollectionReqs()
		}
	case "down", "j":
		if m.collIdx < len(m.collections)-1 {
			m.collIdx++
			m.loadCollectionReqs()
		}
	case "enter":
		if len(m.collections) > 0 {
			if m.collMode == collDetail {
				if len(m.collReqs) > 0 && m.collReqIdx < len(m.collReqs) {
					m.loadRequestIntoEditor(m.collReqs[m.collReqIdx])
					m.message = "loaded request into editor"
					m.messageTime = timeNow()
				}
			} else {
				m.collMode = collDetail
				m.collReqIdx = 0
				m.loadCollectionReqs()
			}
		}
	case "esc":
		if m.collMode == collDetail {
			m.collMode = collList
		} else {
			m.activeTab = tabRequest
			m.focusArea = focusURL
			m.focusCurrent()
		}
	}
	return m, nil
}

func (m model) handleHistoryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "d", "delete":
		if len(m.history) > 0 && m.histIdx < len(m.history) {
			entry := m.history[m.histIdx]
			DeleteHistoryEntry(entry)
			m.history, _ = ListHistory()
			if m.histIdx >= len(m.history) {
				m.histIdx = max(0, len(m.history)-1)
			}
			m.message = "deleted history entry"
			m.messageTime = timeNow()
		}
		return m, nil
	case "up", "k":
		if m.histIdx > 0 {
			m.histIdx--
		}
	case "down", "j":
		if m.histIdx < len(m.history)-1 {
			m.histIdx++
		}
	case "enter":
		if len(m.history) > 0 && m.histIdx < len(m.history) {
			entry := m.history[m.histIdx]
			m.loadRequestIntoEditor(entry.Request)
			resp := entry.Response
			m.response = &resp
			m.message = "reused request from history"
			m.messageTime = timeNow()
		}
	case "esc":
		m.activeTab = tabRequest
		m.focusArea = focusURL
		m.focusCurrent()
	}
	return m, nil
}

func (m model) handleEnvironmentsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.creatingMode {
		switch msg.String() {
		case "esc":
			m.creatingMode = false
			m.creatingName.SetValue("")
			return m, nil
		case "enter":
			name := m.creatingName.Value()
			if name != "" {
				env := NewEnvironment(name)
				SaveEnvironment(env)
				m.envs, _ = ListEnvironments()
				m.message = fmt.Sprintf("created environment '%s'", name)
				m.messageTime = timeNow()
			}
			m.creatingMode = false
			m.creatingName.SetValue("")
			return m, nil
		default:
			var cmd tea.Cmd
			m.creatingName, cmd = m.creatingName.Update(msg)
			return m, cmd
		}
	}

	if m.envEdit != nil {
		return m.handleEnvEditKey(msg)
	}

	switch msg.String() {
	case "n":
		m.creatingMode = true
		m.creatingName.SetValue("")
		m.creatingName.Focus()
		return m, nil
	case "a", " ":
		// Activate (or deactivate) the selected environment.
		if len(m.envs) > 0 && m.envIdx < len(m.envs) {
			name := m.envs[m.envIdx].Name
			if m.activeEnv == name {
				m.activeEnv = ""
				m.message = "environment deactivated"
			} else {
				m.activeEnv = name
				m.message = fmt.Sprintf("activated environment '%s'", name)
			}
			m.persistConfig()
			m.messageTime = timeNow()
		}
		return m, nil
	case "d", "delete":
		if len(m.envs) > 0 && m.envIdx < len(m.envs) {
			name := m.envs[m.envIdx].Name
			if m.activeEnv == name {
				m.activeEnv = ""
				m.persistConfig()
			}
			DeleteEnvironment(name)
			m.envs, _ = ListEnvironments()
			if m.envIdx >= len(m.envs) {
				m.envIdx = max(0, len(m.envs)-1)
			}
			m.message = fmt.Sprintf("deleted environment '%s'", name)
			m.messageTime = timeNow()
		}
		return m, nil
	case "up", "k":
		if m.envIdx > 0 {
			m.envIdx--
		}
	case "down", "j":
		if m.envIdx < len(m.envs)-1 {
			m.envIdx++
		}
	case "enter":
		if len(m.envs) > 0 && m.envIdx < len(m.envs) {
			m.openEnvEditor()
		}
	case "esc":
		m.activeTab = tabRequest
		m.focusArea = focusURL
		m.focusCurrent()
	}
	return m, nil
}

// openEnvEditor seeds the inline variable editor from the selected environment.
func (m *model) openEnvEditor() {
	env := m.envs[m.envIdx]
	keys := make([]string, 0, len(env.Values))
	for k := range env.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vars := make([]keyValue, 0, len(keys))
	for _, k := range keys {
		vars = append(vars, newKeyValue(k, env.Values[k]))
	}
	if len(vars) == 0 {
		vars = append(vars, newKeyValue("", ""))
	}

	nameInput := textinput.New()
	nameInput.SetValue(env.Name)

	m.envEdit = &envEditState{Name: nameInput, Vars: vars, Idx: 0, Col: 0}
	m.focusEnvEditor()
}

// focusEnvEditor focuses the input under the editor cursor and blurs the rest.
func (m *model) focusEnvEditor() {
	e := m.envEdit
	if e == nil {
		return
	}
	for i := range e.Vars {
		e.Vars[i].Key.Blur()
		e.Vars[i].Value.Blur()
	}
	if e.Idx < len(e.Vars) {
		if e.Col == 0 {
			e.Vars[e.Idx].Key.Focus()
		} else {
			e.Vars[e.Idx].Value.Focus()
		}
	}
}

// saveEnvEditor writes the inline editor's variables back to disk.
func (m *model) saveEnvEditor() {
	e := m.envEdit
	if e == nil || m.envIdx >= len(m.envs) {
		return
	}
	vals := make(map[string]string)
	for _, kv := range e.Vars {
		k := strings.TrimSpace(kv.Key.Value())
		if k != "" {
			vals[k] = kv.Value.Value()
		}
	}
	m.envs[m.envIdx].Values = vals
	_ = SaveEnvironment(m.envs[m.envIdx])
	m.envEdit = nil
	m.message = "environment saved"
	m.messageTime = timeNow()
}

// handleEnvEditKey drives the inline key/value editor for an environment.
func (m model) handleEnvEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	e := m.envEdit

	switch msg.String() {
	case "esc":
		m.envEdit = nil
		return m, nil
	case "ctrl+s":
		m.saveEnvEditor()
		return m, nil
	case "up":
		if e.Idx > 0 {
			e.Idx--
			m.focusEnvEditor()
		}
		return m, nil
	case "down":
		if e.Idx < len(e.Vars)-1 {
			e.Idx++
			m.focusEnvEditor()
		}
		return m, nil
	case "tab", "right":
		if e.Col == 0 {
			e.Col = 1
		} else if e.Idx < len(e.Vars)-1 {
			e.Idx++
			e.Col = 0
		}
		m.focusEnvEditor()
		return m, nil
	case "shift+tab", "left":
		if e.Col == 1 {
			e.Col = 0
		} else if e.Idx > 0 {
			e.Idx--
			e.Col = 1
		}
		m.focusEnvEditor()
		return m, nil
	case "enter":
		// Append a new empty row after the current one.
		e.Vars = append(e.Vars, newKeyValue("", ""))
		e.Idx = len(e.Vars) - 1
		e.Col = 0
		m.focusEnvEditor()
		return m, nil
	case "backspace":
		cur := e.Vars[e.Idx]
		if cur.Key.Value() == "" && cur.Value.Value() == "" && len(e.Vars) > 1 {
			e.Vars = append(e.Vars[:e.Idx], e.Vars[e.Idx+1:]...)
			if e.Idx >= len(e.Vars) {
				e.Idx = len(e.Vars) - 1
			}
			m.focusEnvEditor()
			return m, nil
		}
	}

	// Forward the keystroke to the focused input.
	if e.Idx < len(e.Vars) {
		var cmd tea.Cmd
		if e.Col == 0 {
			e.Vars[e.Idx].Key, cmd = e.Vars[e.Idx].Key.Update(msg)
		} else {
			e.Vars[e.Idx].Value, cmd = e.Vars[e.Idx].Value.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

func (m model) handleConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i":
		if m.configSection == 0 {
			m.importMode = true
			m.importBuf = ""
			return m, nil
		}
	case "e":
		if m.configSection == 0 {
			all := m.allThemes()
			if m.themeIdx < len(all) {
				name := all[m.themeIdx]
				var t theme
				if m.themeIdx < len(themes) {
					t = themes[m.themeIdx]
				} else {
					t = m.customThemes[m.themeIdx-len(themes)]
				}
				data, err := exportThemeJSON(t)
				if err == nil {
					exportPath := filepath.Join(getThemesDir(), name+".json")
					os.MkdirAll(filepath.Dir(exportPath), 0o755)
					os.WriteFile(exportPath, data, 0o644)
					m.message = fmt.Sprintf("exported to %s", exportPath)
					m.messageTime = timeNow()
				}
			}
			return m, nil
		}
	case "d", "delete":
		if m.configSection == 0 && m.themeIdx >= len(themes) {
			customIdx := m.themeIdx - len(themes)
			if customIdx >= 0 && customIdx < len(m.customThemes) {
				name := m.customThemes[customIdx].name
				deleteCustomTheme(name)
				m.customThemes = loadCustomThemes()
				total := len(themes) + len(m.customThemes)
				if m.themeIdx >= total {
					m.themeIdx = total - 1
				}
				if m.themeIdx < 0 {
					m.themeIdx = 0
				}
				m.rebuildStyles()
				m.message = fmt.Sprintf("deleted '%s'", name)
				m.messageTime = timeNow()
			}
			return m, nil
		}
	case "enter":
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
				m.activeEnv = ""
				m.config = loadConfig()
				m.rebuildStyles()
			}
		}
		return m, nil
	case "up", "k":
		m.configCursor--
		if m.configCursor < 0 {
			m.configCursor = m.configMax()
		}
		if m.configSection == 1 {
			if m.configCursor < m.colorScroll {
				m.colorScroll = m.configCursor
			}
			if m.colorScroll < 0 {
				m.colorScroll = 0
			}
		}
		return m, nil
	case "down", "j":
		m.configCursor++
		if m.configCursor > m.configMax() {
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
		return m, nil
	case "tab":
		m.configSection = (m.configSection + 1) % 4
		m.configCursor = 0
		if m.configSection == 1 {
			m.colorScroll = 0
		}
		return m, nil
	}

	return m, nil
}

func (m model) handleImportKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.importMode = false
		m.importBuf = ""
		return m, nil
	case "enter":
		if m.importBuf != "" {
			themes, err := importThemesJSON([]byte(m.importBuf))
			if err != nil {
				m.message = "invalid JSON"
				m.messageTime = timeNow()
			} else {
				count := 0
				for _, ct := range themes {
					if err := saveCustomTheme(ct); err == nil {
						count++
					}
				}
				m.customThemes = loadCustomThemes()
				m.message = fmt.Sprintf("imported %d theme(s)", count)
				m.messageTime = timeNow()
			}
		}
		m.importMode = false
		m.importBuf = ""
		return m, nil
	case "backspace":
		if len(m.importBuf) > 0 {
			m.importBuf = m.importBuf[:len(m.importBuf)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.importBuf += msg.String()
		}
		return m, nil
	}
}

func (m model) sendRequest() tea.Cmd {
	req := m.currentRequest()
	if req.ID == "" {
		req.ID = generateID()
	}
	env := m.resolveEnv()
	cfg := DefaultHTTPClientConfig()
	return func() tea.Msg {
		resp := ExecuteRequest(req, env, cfg)
		return responseMsg{response: resp, request: req}
	}
}

type responseMsg struct {
	response Response
	request  Request
}

func (m *model) setTheme(idx int) {
	m.themeIdx = idx
	m.config.Theme = idx
	m.rebuildStyles()
	m.persistConfig()
}

func (m *model) setColor(idx int) {
	m.colorIdx = idx
	m.config.Color = idx
	m.rebuildStyles()
	m.persistConfig()
}

func (m *model) setMode(idx int) {
	m.modeIdx = idx
	m.config.Mode = idx
	m.rebuildStyles()
	m.persistConfig()
}

// persistConfig writes the current live settings to disk in one place so the
// individual setters can't drift from what's saved.
func (m *model) persistConfig() {
	m.config.Theme = m.themeIdx
	m.config.Color = m.colorIdx
	m.config.Mode = m.modeIdx
	m.config.ActiveEnv = m.activeEnv
	saveConfig(m.config)
}

func timeNow() time.Time {
	return time.Now()
}

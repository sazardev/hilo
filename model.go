package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type tab int

const (
	tabRequest tab = iota
	tabCollections
	tabHistory
	tabEnvironments
	tabConfig
	tabHelp
)

var tabNames = []string{"Request", "Collections", "History", "Environments", "Config", "Help"}

type subTab int

const (
	subTabParams subTab = iota
	subTabHeaders
	subTabAuth
	subTabBody
)

var subTabNames = []string{"Params", "Headers", "Auth", "Body"}

type focusArea int

// Focus regions, ordered to match the visual layout top-to-bottom so Tab
// walks through them the way a user reads the screen.
const (
	focusURL focusArea = iota
	focusActions
	focusSubTabs
	focusEditor
	focusResponse
)

type authTypeIdx int

const (
	authNone authTypeIdx = iota
	authBearer
	authBasic
	authDigest
	authAPIKey
	authOAuth2
)

var authTypeNames = []string{"None", "Bearer Token", "Basic Auth", "Digest Auth", "API Key", "OAuth 2.0"}

type bodyTypeIdx int

const (
	bodyNone bodyTypeIdx = iota
	bodyJSON
	bodyForm
	bodyRaw
	bodyBinary
)

var bodyTypeNames = []string{"None", "JSON", "Form Data", "Raw", "Binary"}

type respMode int

const (
	respPretty respMode = iota
	respRaw
	respHeaders
	respCookies
)

var respModeNames = []string{"Pretty", "Raw", "Headers", "Cookies"}

// overlayKind identifies a modal overlay drawn over the whole screen.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayCurl
	overlaySnippet
)

type collViewMode int

const (
	collList collViewMode = iota
	collDetail
	collGitLog
	collGitDiff
	collGitBranch
)

type keyValue struct {
	Key   textinput.Model
	Value textinput.Model
}

func newKeyValue(key, value string) keyValue {
	kt := textinput.New()
	kt.Placeholder = "key"
	kt.SetValue(key)
	kt.CharLimit = 256

	vt := textinput.New()
	vt.Placeholder = "value"
	vt.SetValue(value)
	vt.CharLimit = 4096

	return keyValue{Key: kt, Value: vt}
}

type envEditState struct {
	Name textinput.Model
	Vars []keyValue
	Idx  int
	Col  int
}

type model struct {
	width, height int
	styles        *styles
	spinner       spinner.Model

	activeTab tab
	subTab    subTab
	focusArea focusArea

	methodIdx int
	urlInput  textinput.Model
	params    []keyValue
	headers   []keyValue
	editorRow int
	editorCol int

	authType     authTypeIdx
	authFieldIdx int
	authKey      textinput.Model
	authValue    textinput.Model
	authUser     textinput.Model
	authPass     textinput.Model
	bodyType     bodyTypeIdx
	bodyInput    textarea.Model

	response       *Response
	responseMode   respMode
	responseScroll int
	sending        bool
	lastError      string
	lastSentReq    Request

	actionIdx int

	searchMode    bool
	searchInput   textinput.Model
	searchMatches []int
	searchIdx     int

	collections []Collection
	collIdx     int
	collReqs    []Request
	collReqIdx  int
	collMode    collViewMode

	// Git versioning UI state (within the Collections tab).
	gitLog        []GitLogEntry
	gitLogIdx     int
	gitDiff       string
	gitDiffScroll int
	gitReqPath    string // repo-relative path of the request being inspected
	gitReqName    string
	gitBranches   []string
	gitBranchIdx  int
	gitCurBranch  string

	history []HistoryEntry
	histIdx int

	envs      []Environment
	envIdx    int
	envEdit   *envEditState
	activeEnv string

	config        appConfig
	themeIdx      int
	colorIdx      int
	modeIdx       int
	configSection int
	configCursor  int
	colorScroll   int
	customThemes  []theme

	importMode bool
	importBuf  string

	// Overlays and import/export UI.
	overlay       overlayKind
	pasteArea     textarea.Model
	snippetLang   int
	pathInput     textinput.Model
	collImporting bool

	message     string
	messageTime time.Time

	creatingName textinput.Model
	creatingMode bool
}

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

func newModel() model {
	urlInput := textinput.New()
	urlInput.Placeholder = "https://api.example.com/users"
	urlInput.CharLimit = 4096
	urlInput.Width = 60
	urlInput.Prompt = ""

	authKey := textinput.New()
	authKey.Placeholder = "Token / Key"
	authKey.CharLimit = 512

	authValue := textinput.New()
	authValue.Placeholder = "Value"
	authValue.CharLimit = 512

	authUser := textinput.New()
	authUser.Placeholder = "Username"
	authUser.CharLimit = 256

	authPass := textinput.New()
	authPass.Placeholder = "Password"
	authPass.CharLimit = 256
	authPass.EchoMode = textinput.EchoPassword // mask secrets in the UI

	bodyInput := textarea.New()
	bodyInput.Placeholder = "Request body..."
	bodyInput.CharLimit = 1024 * 1024
	bodyInput.ShowLineNumbers = false

	s := spinner.New()
	s.Spinner = spinner.Dot

	creatingName := textinput.New()
	creatingName.Placeholder = "name..."
	creatingName.CharLimit = 64

	searchInput := textinput.New()
	searchInput.Placeholder = "search response..."
	searchInput.CharLimit = 256
	searchInput.Prompt = ""

	pasteArea := textarea.New()
	pasteArea.Placeholder = "paste a curl command here..."
	pasteArea.CharLimit = 1024 * 64
	pasteArea.ShowLineNumbers = false

	pathInput := textinput.New()
	pathInput.Placeholder = "/path/to/collection.json"
	pathInput.CharLimit = 1024

	cfg := loadConfig()
	ct := loadCustomThemes()

	// Validate the theme index now that custom themes are known.
	if cfg.Theme >= len(themes)+len(ct) {
		cfg.Theme = 0
	}

	m := model{
		spinner:      s,
		activeTab:    tabRequest,
		subTab:       subTabParams,
		focusArea:    focusURL,
		methodIdx:    0,
		urlInput:     urlInput,
		params:       []keyValue{newKeyValue("", "")},
		headers:      []keyValue{newKeyValue("", "")},
		editorRow:    0,
		editorCol:    0,
		authType:     authNone,
		authKey:      authKey,
		authValue:    authValue,
		authUser:     authUser,
		authPass:     authPass,
		bodyType:     bodyNone,
		bodyInput:    bodyInput,
		responseMode: respPretty,
		collMode:     collList,
		config:       cfg,
		themeIdx:     cfg.Theme,
		colorIdx:     cfg.Color,
		modeIdx:      cfg.Mode,
		activeEnv:    cfg.ActiveEnv,
		customThemes: ct,
		creatingName: creatingName,
		searchInput:  searchInput,
		pasteArea:    pasteArea,
		pathInput:    pathInput,
	}
	m.loadTabData()
	m.rebuildStyles()
	m.focusCurrent()
	return m
}

func (m model) Init() tea.Cmd {
	// No continuous ticking at startup — the UI only redraws in response to
	// input or in-flight requests, keeping it idle-quiet.
	return nil
}

// responseLineCount is the number of lines in the current response view mode.
func (m model) responseLineCount() int {
	return len(m.responseLines())
}

// responsePageSize estimates the visible response height for page scrolling.
func (m model) responsePageSize() int {
	if m.layoutMode() == layoutWide {
		return max(m.height-10, 5)
	}
	return max(m.height/2-6, 4)
}

func (m *model) blurAll() {
	m.urlInput.Blur()
	m.bodyInput.Blur()
	m.authKey.Blur()
	m.authValue.Blur()
	m.authUser.Blur()
	m.authPass.Blur()
	for i := range m.params {
		m.params[i].Key.Blur()
		m.params[i].Value.Blur()
	}
	for i := range m.headers {
		m.headers[i].Key.Blur()
		m.headers[i].Value.Blur()
	}
}

func (m *model) focusCurrent() {
	m.blurAll()
	switch m.focusArea {
	case focusURL:
		m.urlInput.Focus()
	case focusEditor:
		switch m.subTab {
		case subTabParams:
			if m.editorRow < len(m.params) {
				if m.editorCol == 0 {
					m.params[m.editorRow].Key.Focus()
				} else {
					m.params[m.editorRow].Value.Focus()
				}
			}
		case subTabHeaders:
			if m.editorRow < len(m.headers) {
				if m.editorCol == 0 {
					m.headers[m.editorRow].Key.Focus()
				} else {
					m.headers[m.editorRow].Value.Focus()
				}
			}
		case subTabAuth:
			m.focusAuthField()
		case subTabBody:
			m.bodyInput.Focus()
		}
	}
}

// authFields returns the focusable inputs for the current auth type, in tab
// order. The four backing inputs are reused across types (for OAuth2 they map
// to token URL, client id, client secret and scope).
func (m *model) authFields() []*textinput.Model {
	switch m.authType {
	case authBearer:
		return []*textinput.Model{&m.authValue}
	case authBasic, authDigest:
		return []*textinput.Model{&m.authUser, &m.authPass}
	case authAPIKey:
		return []*textinput.Model{&m.authKey, &m.authValue}
	case authOAuth2:
		return []*textinput.Model{&m.authKey, &m.authUser, &m.authPass, &m.authValue}
	}
	return nil
}

// focusAuthField focuses the input under authFieldIdx and blurs the others.
func (m *model) focusAuthField() {
	m.authKey.Blur()
	m.authValue.Blur()
	m.authUser.Blur()
	m.authPass.Blur()

	fields := m.authFields()
	if m.authFieldIdx >= len(fields) {
		m.authFieldIdx = 0
	}
	if m.authFieldIdx < len(fields) {
		fields[m.authFieldIdx].Focus()
	}
}

// focusNextAuthField advances to the next auth input, returning false when at
// the last field so Tab can move on to the next focus region.
func (m *model) focusNextAuthField() bool {
	if m.authFieldIdx < len(m.authFields())-1 {
		m.authFieldIdx++
		m.focusAuthField()
		return true
	}
	return false
}

func (m *model) focusPrevAuthField() bool {
	if m.authFieldIdx > 0 {
		m.authFieldIdx--
		m.focusAuthField()
		return true
	}
	return false
}

func (m model) currentRequest() Request {
	reqHeaders := make(map[string]string)
	for _, kv := range m.headers {
		k := strings.TrimSpace(kv.Key.Value())
		v := strings.TrimSpace(kv.Value.Value())
		if k != "" {
			reqHeaders[k] = v
		}
	}

	params := make(map[string]string)
	for _, kv := range m.params {
		k := strings.TrimSpace(kv.Key.Value())
		v := strings.TrimSpace(kv.Value.Value())
		if k != "" {
			params[k] = v
		}
	}

	var auth *Auth
	switch m.authType {
	case authBearer:
		auth = &Auth{Type: AuthBearer, Key: m.authKey.Value(), Value: m.authValue.Value()}
	case authBasic:
		auth = &Auth{Type: AuthBasic, Username: m.authUser.Value(), Password: m.authPass.Value()}
	case authAPIKey:
		auth = &Auth{Type: AuthAPIKey, Key: m.authKey.Value(), Value: m.authValue.Value()}
	case authDigest:
		auth = &Auth{Type: AuthDigest, Username: m.authUser.Value(), Password: m.authPass.Value()}
	case authOAuth2:
		auth = &Auth{
			Type:         AuthOAuth2,
			TokenURL:     m.authKey.Value(),
			ClientID:     m.authUser.Value(),
			ClientSecret: m.authPass.Value(),
			Scope:        m.authValue.Value(),
		}
	}

	var bt string
	switch m.bodyType {
	case bodyJSON:
		bt = BodyJSON
	case bodyForm:
		bt = BodyForm
	case bodyRaw:
		bt = BodyText
	case bodyBinary:
		bt = BodyFile
	default:
		bt = BodyNone
	}

	return Request{
		Name:        fmt.Sprintf("%s %s", methods[m.methodIdx], shortURL(m.urlInput.Value())),
		Method:      methods[m.methodIdx],
		URL:         m.urlInput.Value(),
		Headers:     reqHeaders,
		QueryParams: params,
		Body:        m.bodyInput.Value(),
		BodyType:    bt,
		Auth:        auth,
	}
}

// loadRequestIntoEditor populates every editor field from a stored request and
// switches focus to the Request tab. Shared by the Collections and History tabs.
func (m *model) loadRequestIntoEditor(req Request) {
	m.urlInput.SetValue(req.URL)

	m.methodIdx = 0
	for i, me := range methods {
		if me == req.Method {
			m.methodIdx = i
			break
		}
	}

	m.headers = nil
	for k, v := range req.Headers {
		m.headers = append(m.headers, newKeyValue(k, v))
	}
	if len(m.headers) == 0 {
		m.headers = []keyValue{newKeyValue("", "")}
	}

	m.params = nil
	for k, v := range req.QueryParams {
		m.params = append(m.params, newKeyValue(k, v))
	}
	if len(m.params) == 0 {
		m.params = []keyValue{newKeyValue("", "")}
	}

	m.bodyInput.SetValue(req.Body)
	m.bodyType = bodyTypeFromString(req.BodyType)

	m.authType = authNone
	m.authFieldIdx = 0
	if req.Auth != nil {
		switch req.Auth.Type {
		case AuthBearer:
			m.authType = authBearer
			m.authKey.SetValue(req.Auth.Key)
			m.authValue.SetValue(req.Auth.Value)
		case AuthBasic:
			m.authType = authBasic
			m.authUser.SetValue(req.Auth.Username)
			m.authPass.SetValue(req.Auth.Password)
		case AuthDigest:
			m.authType = authDigest
			m.authUser.SetValue(req.Auth.Username)
			m.authPass.SetValue(req.Auth.Password)
		case AuthAPIKey:
			m.authType = authAPIKey
			m.authKey.SetValue(req.Auth.Key)
			m.authValue.SetValue(req.Auth.Value)
		case AuthOAuth2:
			m.authType = authOAuth2
			m.authKey.SetValue(req.Auth.TokenURL)
			m.authUser.SetValue(req.Auth.ClientID)
			m.authPass.SetValue(req.Auth.ClientSecret)
			m.authValue.SetValue(req.Auth.Scope)
		}
	}

	m.activeTab = tabRequest
	m.focusArea = focusURL
	m.editorRow = 0
	m.editorCol = 0
	m.focusCurrent()
}

func bodyTypeFromString(bt string) bodyTypeIdx {
	switch bt {
	case BodyJSON:
		return bodyJSON
	case BodyForm:
		return bodyForm
	case BodyText, BodyXML:
		return bodyRaw
	case BodyFile:
		return bodyBinary
	default:
		return bodyNone
	}
}

func shortURL(raw string) string {
	s := raw
	if i := strings.Index(s, "://"); i != -1 {
		s = s[i+3:]
	}
	if len(s) > 40 {
		s = s[:37] + "..."
	}
	return s
}

func (m *model) loadTabData() {
	m.collections, _ = ListCollections()
	m.history, _ = ListHistory()
	m.envs, _ = ListEnvironments()
}

// resolveEnv returns the variable map of the active environment for {{VAR}}
// substitution. Returns an empty (non-nil) map when no environment is active.
func (m model) resolveEnv() map[string]string {
	if m.activeEnv == "" {
		return map[string]string{}
	}
	for _, e := range m.envs {
		if e.Name == m.activeEnv {
			return e.Values
		}
	}
	return map[string]string{}
}

// isTyping reports whether the current focus is inside a free-text input, in
// which case single-letter keys (q, h, j, k, l, n, ...) must be typed rather
// than treated as navigation shortcuts.
func (m model) isTyping() bool {
	if m.creatingMode || m.importMode || m.searchMode || m.envEdit != nil || m.collImporting {
		return true
	}
	if m.overlay != overlayNone {
		return true
	}
	if m.activeTab == tabRequest {
		if m.focusArea == focusURL || m.focusArea == focusEditor {
			return true
		}
	}
	return false
}

// layout describes how the request editor and response panels are arranged,
// derived from the terminal width per the responsive spec.
type layout int

const (
	layoutWide    layout = iota // >=120 cols: editor and response side by side
	layoutStacked               // >=80 cols: editor above, response below
	layoutCompact               // <80 cols: minimal chrome
)

func (m model) layoutMode() layout {
	switch {
	case m.width >= 120:
		return layoutWide
	case m.width >= 80:
		return layoutStacked
	default:
		return layoutCompact
	}
}

// responseLines returns the plain (unstyled) lines of the current response in
// the active view mode. Shared by the renderer and the search index so that
// match line numbers line up with what is displayed.
func (m model) responseLines() []string {
	if m.response == nil {
		return nil
	}
	return strings.Split(m.responseContent(), "\n")
}

func (m model) responseContent() string {
	r := m.response
	if r == nil {
		return ""
	}
	switch m.responseMode {
	case respRaw:
		return r.Body
	case respHeaders:
		return headerLines(r.Headers)
	case respCookies:
		return cookieLines(r.Headers)
	default: // respPretty
		return PrettyJSON(r.Body)
	}
}

func headerLines(h map[string]string) string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(h[k])
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// currentRepo opens the git repository of the selected collection.
func (m model) currentRepo() (*CollectionRepo, bool) {
	if m.collIdx < 0 || m.collIdx >= len(m.collections) {
		return nil, false
	}
	repo, err := OpenCollectionRepo(m.collections[m.collIdx].Name)
	if err != nil {
		return nil, false
	}
	return repo, true
}

// gitRepoPath returns the forward-slash, repo-relative path of a request file
// (git always uses forward slashes regardless of OS).
func gitRepoPath(requestID string) string {
	return "requests/" + requestID + ".json"
}

// relativeTime renders a compact "time ago" string for commit timestamps.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

func cookieLines(h map[string]string) string {
	for k, v := range h {
		if strings.EqualFold(k, "Set-Cookie") {
			var out []string
			for _, p := range strings.Split(v, ",") {
				if p = strings.TrimSpace(p); p != "" {
					out = append(out, p)
				}
			}
			return strings.Join(out, "\n")
		}
	}
	return ""
}

func (m *model) loadCollectionReqs() {
	if m.collIdx >= 0 && m.collIdx < len(m.collections) {
		coll := m.collections[m.collIdx]
		m.collReqs, _ = ListRequests(coll.Name)
	}
}

func (m model) allThemes() []string {
	var names []string
	for _, t := range themes {
		names = append(names, t.name)
	}
	for _, t := range m.customThemes {
		names = append(names, t.name)
	}
	return names
}

func (m model) configMax() int {
	switch m.configSection {
	case 0:
		return len(m.allThemes()) - 1
	case 1:
		return len(colorSchemes) - 1
	case 2:
		return 1
	}
	return 0
}

func (m model) configCurrent() string {
	switch m.configSection {
	case 0:
		return m.allThemes()[m.themeIdx]
	case 1:
		return colorSchemes[m.colorIdx].name
	case 2:
		if m.modeIdx == 0 {
			return "Dark"
		}
		return "Light"
	}
	return ""
}

func (m *model) updateConfig() {
	m.config.Theme = m.themeIdx
	m.config.Color = m.colorIdx
	m.config.Mode = m.modeIdx
	saveConfig(m.config)
}

func (m *model) rebuildStyles() {
	t := m.currentTheme()

	// Apply the selected color scheme as an override on the theme's
	// primary/accent so the "color" axis is meaningful regardless of theme.
	if m.colorIdx >= 0 && m.colorIdx < len(colorSchemes) {
		cs := colorSchemes[m.colorIdx]
		t.primary = cs.primary
		t.accent = cs.accent
	}

	m.styles = newStyles(t)
}

func (m model) currentTheme() theme {
	if m.themeIdx < len(themes) {
		return themes[m.themeIdx]
	}
	idx := m.themeIdx - len(themes)
	if idx >= 0 && idx < len(m.customThemes) {
		return m.customThemes[idx]
	}
	return themes[0]
}

package tui

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"recomphamr2/internal/security"
)

// Model is the complete Bubble Tea TUI state.
type Model struct {
	snapshot   Snapshot
	composer   textareaModel
	transcript viewportModel
	picker     list.Model
	help       helpModel
	keys       keyMap
	entries    []TranscriptEntry
	overlay    overlayKind
	width      int
	height     int
	history    []string
	historyAt  int
	quitArmed  bool
	newOutput  bool
	profile    colorprofile.Profile
}

// Interface aliases keep component ownership explicit without exposing fields.
type textareaModel = textarea.Model
type viewportModel = viewport.Model
type helpModel = help.Model

// New returns a fully initialized Bubble Tea model.
func New(snapshot Snapshot) Model {
	profile := colorprofile.ANSI256
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		profile = colorprofile.ASCII
	}
	composer := newComposer()
	composer.SetVirtualCursor(false)
	model := Model{
		snapshot:   snapshot,
		composer:   composer,
		transcript: newViewport(),
		picker:     newPicker(profile),
		help:       help.New(),
		keys:       newKeyMap(),
		width:      DefaultWidth,
		height:     DefaultHeight,
		profile:    profile,
	}
	styleComponents(&model.composer, &model.picker, &model.help, profile)
	model.historyAt = 0
	model.resize()
	return model
}

// Init starts textarea cursor blinking.
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// Snapshot returns the current immutable app-owned state.
func (m Model) Snapshot() Snapshot {
	return m.snapshot
}

// ComposerValue returns the authoritative textarea value.
func (m Model) ComposerValue() string {
	return m.composer.Value()
}

// Entries returns a copy of the semantic transcript.
func (m Model) Entries() []TranscriptEntry {
	return append([]TranscriptEntry(nil), m.entries...)
}

// Update handles Bubble Tea and app-owned messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = typed.Width, typed.Height
		m.resize()
		return m, nil
	case tea.ColorProfileMsg:
		m.profile = typed.Profile
		styleComponents(&m.composer, &m.picker, &m.help, m.profile)
		return m, nil
	case ColorProfileMsg:
		m.profile = typed.Profile
		styleComponents(&m.composer, &m.picker, &m.help, m.profile)
		return m, nil
	case SnapshotMsg:
		m.snapshot = typed.Snapshot
		return m, nil
	case TranscriptMsg:
		m.appendTranscript(typed.Entries)
		return m, nil
	case ClearTranscriptMsg:
		m.entries = nil
		m.newOutput = false
		m.transcript.SetContent("")
		return m, nil
	case tea.FocusMsg:
		return m, m.composer.Focus()
	case tea.BlurMsg:
		m.composer.Blur()
		return m, nil
	case tea.KeyPressMsg:
		return m.updateKey(typed)
	case tea.PasteMsg:
		var cmd tea.Cmd
		m.composer, cmd = m.composer.Update(msg)
		return m, cmd
	case tea.MouseWheelMsg:
		if len(m.entries) == 0 {
			return m, nil
		}
		var cmd tea.Cmd
		m.transcript, cmd = m.transcript.Update(msg)
		m.syncTranscriptFollow()
		return m, cmd
	default:
		var cmd tea.Cmd
		if m.overlay != overlayNone {
			m.picker, cmd = m.picker.Update(msg)
		} else {
			m.composer, cmd = m.composer.Update(msg)
		}
		return m, cmd
	}
}

func (m Model) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	name := strings.ToLower(msg.String())
	if m.width < MinimumWidth || m.height < MinimumHeight {
		if name == "ctrl+d" {
			return m, emitIntent(IntentQuit, "")
		}
		return m, nil
	}
	switch name {
	case "ctrl+d":
		return m, emitIntent(IntentQuit, "")
	case "ctrl+c":
		if m.snapshot.PendingTool != "" && m.snapshot.PendingTool != "none" || m.snapshot.Mode == "thinking" || m.snapshot.Mode == "streaming" {
			m.quitArmed = false
			return m, emitIntent(IntentCancel, "")
		}
		if m.quitArmed {
			return m, emitIntent(IntentQuit, "")
		}
		m.quitArmed = true
		m.snapshot.Status = "press Ctrl+C again to quit"
		return m, nil
	case "esc":
		if m.overlay != overlayNone {
			m.closeOverlay()
		} else {
			m.quitArmed = false
			m.snapshot.Status = ""
		}
		return m, nil
	}
	m.quitArmed = false
	if m.overlay != overlayNone {
		return m.updateOverlay(msg)
	}
	if name == "shift+enter" || name == "ctrl+j" {
		m.composer.InsertString("\n")
		return m, nil
	}
	switch name {
	case "enter":
		return m.submit()
	case "?":
		if strings.TrimSpace(m.composer.Value()) == "" {
			m.openOverlay(overlayHelp)
			return m, nil
		}
	case "pgup", "pgdown":
		if len(m.entries) > 0 {
			var cmd tea.Cmd
			m.transcript, cmd = m.transcript.Update(msg)
			m.syncTranscriptFollow()
			return m, cmd
		}
		return m, nil
	case "up":
		if m.composer.LineCount() == 1 {
			m.recall(-1)
			return m, nil
		}
	case "down":
		if m.composer.LineCount() == 1 {
			m.recall(1)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.composer, cmd = m.composer.Update(msg)
	if m.composer.Value() == "/" {
		m.openOverlay(overlayCommands)
	}
	return m, cmd
}

func (m Model) updateOverlay(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	name := strings.ToLower(msg.String())
	switch name {
	case "backspace":
		if m.overlay == overlayCommands && m.picker.FilterInput.Value() == "" {
			m.closeOverlay()
			m.composer.Reset()
			return m, nil
		}
	case "enter":
		return m.acceptSelection()
	case "tab":
		return m.completeSelection()
	}
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

func (m Model) submit() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.composer.Value())
	if value == "" || value == "/" {
		return m, nil
	}
	m.history = append(m.history, value)
	m.historyAt = len(m.history)
	m.composer.Reset()
	if strings.HasPrefix(value, "/") {
		return m, emitIntent(IntentCommand, value)
	}
	return m, emitIntent(IntentSubmit, value)
}

func (m Model) acceptSelection() (tea.Model, tea.Cmd) {
	selected, ok := m.picker.SelectedItem().(pickerItem)
	if !ok || selected.blocked {
		return m, nil
	}
	if m.overlay == overlayCommands {
		switch selected.name {
		case "/models":
			m.openOverlay(overlayModels)
			return m, nil
		case "/skills", "/skill":
			m.openOverlay(overlaySkills)
			return m, nil
		case "/mcp":
			m.openOverlay(overlayMCP)
			return m, nil
		case "/help":
			m.openOverlay(overlayHelp)
			return m, nil
		}
		if commandNeedsInput(selected.name) {
			m.composer.SetValue(selected.name + " ")
			m.composer.MoveToEnd()
			m.closeOverlay()
			return m, nil
		}
		m.closeOverlay()
		m.composer.Reset()
		return m, emitIntent(IntentCommand, selected.name)
	}
	if m.overlay == overlayHelp {
		m.composer.SetValue(selected.name + " ")
		m.composer.MoveToEnd()
		m.closeOverlay()
		return m, nil
	}
	m.closeOverlay()
	m.composer.Reset()
	return m, emitIntent(selected.kind, selected.name)
}

func commandNeedsInput(name string) bool {
	return name == "/skill-audit" || name == "/skill-new"
}

func (m Model) completeSelection() (tea.Model, tea.Cmd) {
	selected, ok := m.picker.SelectedItem().(pickerItem)
	if !ok || selected.blocked {
		return m, nil
	}
	m.composer.SetValue(selected.name + " ")
	m.composer.MoveToEnd()
	m.closeOverlay()
	return m, nil
}

func (m *Model) openOverlay(kind overlayKind) {
	m.overlay = kind
	var items []list.Item
	if kind == overlayCommands {
		items = commandItems()
		m.picker.Title = "Command palette"
	} else {
		for _, item := range pickerItems(kind, m.snapshot) {
			items = append(items, item)
		}
		m.picker.Title = overlayTitle(kind)
	}
	_ = m.picker.SetItems(items)
	m.picker.ResetSelected()
	if kind == overlayCommands {
		m.picker.SetFilterText(pickerQuery(m.composer.Value()))
		m.picker.SetFilterState(list.Filtering)
	} else {
		m.picker.SetFilterState(list.Unfiltered)
	}
	m.resize()
}

func (m *Model) closeOverlay() {
	m.overlay = overlayNone
	m.picker.SetFilterState(list.Unfiltered)
	_ = m.composer.Focus()
	m.resize()
}

func overlayTitle(kind overlayKind) string {
	switch kind {
	case overlayModels:
		return "Select model"
	case overlaySkills:
		return "Select skill"
	case overlayMCP:
		return "MCP servers"
	case overlayHelp:
		return "Help"
	default:
		return "Command palette"
	}
}

func (m *Model) recall(delta int) {
	if len(m.history) == 0 {
		return
	}
	next := m.historyAt + delta
	if next < 0 {
		next = 0
	}
	if next > len(m.history) {
		next = len(m.history)
	}
	m.historyAt = next
	if next == len(m.history) {
		m.composer.Reset()
		return
	}
	m.composer.SetValue(m.history[next])
	m.composer.MoveToEnd()
}

func (m *Model) appendTranscript(entries []TranscriptEntry) {
	if len(entries) == 0 {
		return
	}
	follow := len(m.entries) == 0 || m.transcript.AtBottom()
	for _, entry := range entries {
		entry.Text = normalizeBody(entry.Kind, redact(entry.Text, m.snapshot.Secrets))
		m.entries = append(m.entries, entry)
	}
	m.transcript.SetContent(renderTranscript(m.entries, m.laneWidth(), m.profile))
	if follow {
		m.transcript.GotoBottom()
		m.newOutput = false
	} else {
		m.newOutput = true
	}
}

func (m *Model) syncTranscriptFollow() {
	if m.transcript.AtBottom() {
		m.newOutput = false
	}
}

func normalizeBody(kind TranscriptKind, text string) string {
	trimmed := strings.TrimSpace(text)
	prefix := string(kind) + ":"
	if strings.HasPrefix(strings.ToLower(trimmed), prefix) {
		return strings.TrimSpace(trimmed[len(prefix):])
	}
	return trimmed
}

// ParseEntry classifies one existing backend output line for presentation.
func ParseEntry(line string) TranscriptEntry {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	classes := []TranscriptKind{TranscriptUser, TranscriptAssistant, TranscriptTool, TranscriptMCP, TranscriptVerified, TranscriptWarning, TranscriptBlocked, TranscriptUnsupported, TranscriptAttachment}
	for _, kind := range classes {
		prefix := string(kind) + ":"
		if strings.HasPrefix(lower, prefix) {
			return TranscriptEntry{Kind: kind, Text: strings.TrimSpace(trimmed[len(prefix):])}
		}
	}
	return TranscriptEntry{Kind: TranscriptNote, Text: trimmed}
}

func redact(text string, secrets []string) string {
	for _, secret := range secrets {
		text = security.RedactSecret(text, secret)
	}
	return text
}

func emitIntent(kind IntentKind, value string) tea.Cmd {
	return func() tea.Msg { return IntentMsg{Kind: kind, Value: value} }
}

package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

	"recomphamr2/internal/commands"
	"recomphamr2/internal/security"
)

const (
	// DefaultWidth is the canonical golden-render width for tests and docs.
	DefaultWidth = 120
	// DefaultHeight is the canonical terminal height used by renderer tests.
	DefaultHeight = 32
	// CompactWidth is the threshold below which the evidence panel collapses.
	CompactWidth = 96
	// LargePasteThreshold is the byte count that turns paste text into a chip.
	LargePasteThreshold = 1024
)

const (
	brandWide    = "RECOMP HAMR"
	brandCompact = "RecompHamr"
	domainLine   = "RE . decomp . recomp . evidence-backed reconstruction"
	safetyLine   = "local commands run with your user permissions; prompts start work only after submit"
)

const (
	// KeyEnter submits the current composer text.
	KeyEnter = "enter"
	// KeyBackspace deletes the last composer rune.
	KeyBackspace = "backspace"
	// KeyUp recalls the previous prompt history entry.
	KeyUp = "up"
	// KeyDown recalls the next prompt history entry.
	KeyDown = "down"
	// KeyTab completes the current slash command candidate.
	KeyTab = "tab"
	// KeyCtrlC cancels active work or arms quit when idle.
	KeyCtrlC = "ctrl+c"
	// KeyCtrlD quits immediately.
	KeyCtrlD = "ctrl+d"
	// KeyEsc clears transient palette and quit state.
	KeyEsc = "esc"
)

const (
	// ActionNone means an update changed only local UI state.
	ActionNone Action = "none"
	// ActionSubmit means the composer submitted user or slash-command text.
	ActionSubmit Action = "submit"
	// ActionCancel means the UI requested cancellation of active work.
	ActionCancel Action = "cancel"
	// ActionQuit means the UI requested process exit.
	ActionQuit Action = "quit"
)

// Model is the testable terminal shell state.
type Model struct {
	// Transcript is the visible conversation and command output.
	Transcript []string
	// Env is command execution state owned by internal/commands.
	Env commands.Environment
	// Layout stores current render metadata.
	Layout Layout
	// Composer is the multiline prompt buffer.
	Composer string
	// History stores submitted prompt text newest-last.
	History []string
	// HistoryIndex is the active prompt-history cursor, or len(History).
	HistoryIndex int
	// Attachments stores large paste chips until the next submission.
	Attachments []Attachment
	// Status is the footer status text.
	Status string
	// DebugEnabled controls whether redacted debug lines render.
	DebugEnabled bool
	// DebugLog stores redacted debug entries.
	DebugLog []string
	// DebugSecrets are values removed from debug output.
	DebugSecrets []string
	// QuitArmed records the first idle Ctrl+C in the double-press quit flow.
	QuitArmed bool
}

// Layout contains the visible TUI state that can be rendered without Bubble Tea.
type Layout struct {
	// Width is the terminal width in cells.
	Width int
	// Height is the terminal height in cells.
	Height int
	// Mode is the current UI mode label.
	Mode string
	// ActiveModel is the selected model profile label.
	ActiveModel string
	// ActiveSkill is the active skill indicator.
	ActiveSkill string
	// MCPStatus is the MCP gate/status indicator.
	MCPStatus string
	// ContextStatus is the context-budget evidence indicator.
	ContextStatus string
	// PendingTool is the currently visible tool status.
	PendingTool string
	// MemoryStatus is the memory freshness indicator.
	MemoryStatus string
}

// Event is one testable TUI update message.
type Event struct {
	// Key is a symbolic key constant such as KeyEnter.
	Key string
	// Text is inserted into the composer.
	Text string
	// Paste is pasted text, converted to a chip when large or multiline.
	Paste string
	// Width updates Layout.Width when positive.
	Width int
	// Height updates Layout.Height when positive.
	Height int
}

// Action is the side effect requested by Update.
type Action string

// Attachment describes one large paste chip held outside the composer text.
type Attachment struct {
	// Name is the visible chip identifier.
	Name string
	// Content is the pasted text associated with the chip.
	Content string
}

// BubbleModel adapts Model to the Bubble Tea runtime without owning core logic.
type BubbleModel struct {
	// State is the pure TUI shell state rendered by View.
	State Model
	// LastAction records the latest side effect requested by Update.
	LastAction Action
}

// New returns an empty terminal shell model.
func New(env commands.Environment) Model {
	return Model{Env: env, Layout: DefaultLayout(), HistoryIndex: 0}
}

// NewBubble returns a Bubble Tea adapter around a fresh Model.
func NewBubble(env commands.Environment) BubbleModel {
	return BubbleModel{State: New(env), LastAction: ActionNone}
}

// DefaultLayout returns RecompHamr's evidence-first terminal layout defaults.
func DefaultLayout() Layout {
	return Layout{
		Width:         DefaultWidth,
		Height:        DefaultHeight,
		Mode:          "plan",
		ActiveModel:   "unverified",
		ActiveSkill:   "none",
		MCPStatus:     "gated",
		ContextStatus: "local budget pending",
		PendingTool:   "none",
		MemoryStatus:  "refreshed",
	}
}

// Update applies one terminal event and returns the requested side effect.
func (m Model) Update(event Event) (Model, Action) {
	if event.Width > 0 {
		m.Layout.Width = event.Width
	}
	if event.Height > 0 {
		m.Layout.Height = event.Height
	}
	if event.Paste != "" {
		m = m.Paste(event.Paste)
	}
	if event.Text != "" {
		m.Composer += event.Text
		m.QuitArmed = false
	}
	if event.Key == "" {
		return m, ActionNone
	}
	return m.handleKey(event.Key)
}

// Init satisfies tea.Model without launching background work.
func (b BubbleModel) Init() tea.Cmd {
	return nil
}

// Update translates Bubble Tea messages into pure Model events.
func (b BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	event, ok := bubbleEvent(msg)
	if !ok {
		return b, nil
	}
	state, action := b.State.Update(event)
	b.State = state
	b.LastAction = action
	return b, nil
}

// View renders the current Bubble Tea model.
func (b BubbleModel) View() string {
	return b.State.Render()
}

// Submit dispatches slash commands or appends plain user text.
func (m Model) Submit(text string) Model {
	text = strings.TrimSpace(text)
	if text == "" && len(m.Attachments) == 0 {
		return m
	}
	m.History = append(m.History, text)
	m.HistoryIndex = len(m.History)
	if strings.HasPrefix(text, "/") {
		out, env := commands.Execute(m.Env, text)
		m.Env = env
		m.Transcript = append(m.Transcript, out)
		m.Attachments = nil
		m.Composer = ""
		return m
	}
	m.Transcript = append(m.Transcript, "user: "+submissionText(text, m.Attachments))
	m.Attachments = nil
	m.Composer = ""
	return m
}

// Paste inserts small single-line text or creates a large-paste chip.
func (m Model) Paste(text string) Model {
	if isLargePaste(text) {
		name := fmt.Sprintf("paste-%d", len(m.Attachments)+1)
		m.Attachments = append(m.Attachments, Attachment{Name: name, Content: text})
		m.Transcript = append(m.Transcript, fmt.Sprintf("paste: %s (%d bytes)", name, len(text)))
		return m
	}
	m.Composer += text
	return m
}

// Palette returns slash-command completions for the current composer.
func (m Model) Palette() []string {
	text := strings.TrimSpace(m.Composer)
	if !strings.HasPrefix(text, "/") {
		return nil
	}
	return CompleteCommand(strings.Fields(text + " ")[0])
}

// PaletteRows returns registry-backed command palette rows for rendering.
func (m Model) PaletteRows() []string {
	token := strings.TrimSpace(m.Composer)
	if !strings.HasPrefix(token, "/") {
		return nil
	}
	fields := strings.Fields(token + " ")
	prefix := fields[0]
	var rows []string
	for i, cmd := range commandMatches(prefix) {
		pointer := " "
		if i == 0 {
			pointer = ">"
		}
		rows = append(rows, fmt.Sprintf("%s %-14s %s  usage: %s", pointer, cmd.Name, cmd.Summary, cmd.Usage))
	}
	return rows
}

// Debug records a redacted debug line when debug mode is enabled.
func (m Model) Debug(text string) Model {
	if !m.DebugEnabled {
		return m
	}
	m.DebugLog = append(m.DebugLog, redact(text, m.DebugSecrets))
	return m
}

// Render returns the visible transcript.
func (m Model) Render() string {
	return m.RenderWithLayout(m.Layout)
}

// RenderWithLayout returns the full RecompHamr initiative layout.
func (m Model) RenderWithLayout(layout Layout) string {
	if layout.Width <= 0 {
		layout.Width = DefaultWidth
	}
	if layout.Height <= 0 {
		layout.Height = DefaultHeight
	}
	if layout.Width < CompactWidth {
		return m.renderCompact(layout)
	}
	return m.renderWide(layout)
}

// CompleteCommand returns matching slash command names.
func CompleteCommand(prefix string) []string {
	matches := commandMatches(prefix)
	var out []string
	for _, cmd := range matches {
		out = append(out, cmd.Name)
	}
	return out
}

// Improvements documents intentional differences from OpenCode-style terminal agents.
func Improvements() []string {
	return []string{
		"evidence rail keeps memory, skill, MCP, and tool state visible for reverse-engineering work",
		"right-side evidence deck separates verified context from chat transcript to reduce claim drift",
		"compact mode collapses panels into status bands so narrow terminals remain usable",
		"RecompHamr-owned visual tokens keep the UI distinct from OpenCode while preserving terminal polish",
	}
}

func (m Model) renderWide(layout Layout) string {
	var b strings.Builder
	writeHeader(&b, layout, false)
	writeDivider(&b, layout.Width)
	fmt.Fprintf(&b, "signals                         transcript                                  evidence\n")
	fmt.Fprintf(&b, "memory %-22s %s\n", chip(layout.MemoryStatus), transcriptLine(m.Transcript, 0))
	fmt.Fprintf(&b, "skill  %-22s %s\n", chip(layout.ActiveSkill), transcriptLine(m.Transcript, 1))
	fmt.Fprintf(&b, "mcp    %-22s %s\n", chip(layout.MCPStatus), transcriptLine(m.Transcript, 2))
	fmt.Fprintf(&b, "tool   %-22s %s\n", chip(layout.PendingTool), transcriptLine(m.Transcript, 3))
	fmt.Fprintf(&b, "context %-21s %s\n", chip(layout.ContextStatus), transcriptLine(m.Transcript, 4))
	if len(m.Transcript) == 0 {
		fmt.Fprintf(&b, "ready  %-22s %s\n", chip("verified idle"), "Ask RecompHamr, run /help, or activate a skill.")
	}
	writePalette(&b, m)
	writeFooter(&b, m)
	writeComposer(&b, m)
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderCompact(layout Layout) string {
	var b strings.Builder
	writeHeader(&b, layout, true)
	fmt.Fprintf(&b, "status %s %s %s %s %s\n", chip("memory:"+layout.MemoryStatus), chip("skill:"+layout.ActiveSkill), chip("mcp:"+layout.MCPStatus), chip("tool:"+layout.PendingTool), chip("context:"+layout.ContextStatus))
	writeDivider(&b, layout.Width)
	for _, line := range m.Transcript {
		fmt.Fprintf(&b, "%s\n", transcriptBlock(line))
	}
	writePalette(&b, m)
	writeFooter(&b, m)
	writeComposer(&b, m)
	return strings.TrimRight(b.String(), "\n")
}

func writeHeader(b *strings.Builder, layout Layout, compact bool) {
	brand := brandWide
	if compact {
		brand = brandCompact
	}
	fmt.Fprintf(b, "%s\n", brand)
	fmt.Fprintf(b, "%s\n", domainLine)
	fmt.Fprintf(b, "mode %s  model %s\n", chip(layout.Mode), chip(layout.ActiveModel))
	fmt.Fprintf(b, "safety %s\n", safetyLine)
}

func writeDivider(b *strings.Builder, width int) {
	if width <= 0 {
		width = DefaultWidth
	}
	if width > 96 {
		width = 96
	}
	if width < 24 {
		width = 24
	}
	fmt.Fprintf(b, "%s\n", strings.Repeat("-", width))
}

func chip(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "unverified"
	}
	return "[" + text + "]"
}

func transcriptLine(lines []string, index int) string {
	if index >= len(lines) {
		return ""
	}
	return transcriptBlock(lines[index])
}

func transcriptBlock(line string) string {
	label := "note"
	switch {
	case strings.HasPrefix(line, "user:"):
		label = "user"
	case strings.HasPrefix(line, "assistant:"):
		label = "assistant"
	case strings.HasPrefix(line, "tool:"):
		label = "tool"
	case strings.HasPrefix(line, "mcp ") || strings.HasPrefix(line, "mcp:"):
		label = "mcp"
	case strings.HasPrefix(line, "blocked:"):
		label = "blocked"
	case strings.HasPrefix(line, "unsupported:"):
		label = "unsupported"
	case strings.HasPrefix(line, "unverified:"):
		label = "unverified"
	case strings.HasPrefix(line, "status:"):
		label = "status"
	case strings.HasPrefix(line, "paste:"):
		label = "attachment"
	}
	return fmt.Sprintf("%-11s %s", label, line)
}

func (m Model) handleKey(key string) (Model, Action) {
	switch key {
	case KeyEnter:
		before := len(m.Transcript)
		m = m.Submit(m.Composer)
		if len(m.Transcript) == before {
			return m, ActionNone
		}
		return m, ActionSubmit
	case KeyBackspace:
		m.Composer = trimLastRune(m.Composer)
		return m, ActionNone
	case KeyUp:
		m = m.recall(-1)
		return m, ActionNone
	case KeyDown:
		m = m.recall(1)
		return m, ActionNone
	case KeyTab:
		m = m.completeComposer()
		return m, ActionNone
	case KeyCtrlC:
		return m.ctrlC()
	case KeyCtrlD:
		m.Status = "quit"
		return m, ActionQuit
	case KeyEsc:
		m.QuitArmed = false
		m.Status = ""
		return m, ActionNone
	default:
		m.Status = "unsupported key: " + key
		return m, ActionNone
	}
}

func (m Model) completeComposer() Model {
	text := strings.TrimSpace(m.Composer)
	if !strings.HasPrefix(text, "/") {
		return m
	}
	fields := strings.Fields(text)
	prefix := text
	suffix := ""
	if len(fields) > 0 {
		prefix = fields[0]
		if len(fields) > 1 {
			suffix = " " + strings.Join(fields[1:], " ")
		}
	}
	matches := CompleteCommand(prefix)
	if len(matches) == 0 {
		m.Status = "unverified: no command matches " + prefix
		return m
	}
	m.Composer = matches[0] + suffix
	if suffix == "" {
		m.Composer += " "
	}
	m.Status = "completed command: " + matches[0]
	return m
}

func (m Model) ctrlC() (Model, Action) {
	if m.Layout.PendingTool != "none" || m.Layout.Mode == "thinking" || m.Layout.Mode == "streaming" {
		m.Layout.PendingTool = "none"
		m.Layout.Mode = "idle"
		m.Status = "cancelled"
		m.Transcript = append(m.Transcript, "status: cancelled")
		m.QuitArmed = false
		return m, ActionCancel
	}
	if m.QuitArmed {
		m.Status = "quit"
		return m, ActionQuit
	}
	m.QuitArmed = true
	m.Status = "press Ctrl+C again to quit"
	return m, ActionNone
}

func (m Model) recall(delta int) Model {
	if len(m.History) == 0 {
		return m
	}
	next := m.HistoryIndex + delta
	if next < 0 {
		next = 0
	}
	if next > len(m.History) {
		next = len(m.History)
	}
	m.HistoryIndex = next
	if next == len(m.History) {
		m.Composer = ""
		return m
	}
	m.Composer = m.History[next]
	return m
}

func writeFooter(b *strings.Builder, m Model) {
	if m.Status != "" {
		fmt.Fprintf(b, "status > %s\n", m.Status)
	}
	if m.DebugEnabled && len(m.DebugLog) > 0 {
		fmt.Fprintf(b, "debug > %s\n", m.DebugLog[len(m.DebugLog)-1])
	}
	fmt.Fprintf(b, "hints  / commands  Tab complete  Ctrl+C cancel/quit  Ctrl+D exit\n")
}

func writePalette(b *strings.Builder, m Model) {
	rows := m.PaletteRows()
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(b, "commands\n")
	for _, row := range rows {
		fmt.Fprintf(b, "%s\n", row)
	}
}

func writeComposer(b *strings.Builder, m Model) {
	lines := strings.Split(composerView(m), "\n")
	for i, line := range lines {
		if i == 0 {
			fmt.Fprintf(b, "composer > %s\n", line)
			continue
		}
		fmt.Fprintf(b, "           %s\n", line)
	}
}

func composerView(m Model) string {
	text := m.Composer
	for _, attachment := range m.Attachments {
		chip := fmt.Sprintf("[%s %d bytes]", attachment.Name, len(attachment.Content))
		if text == "" {
			text = chip
		} else {
			text += " " + chip
		}
	}
	return text
}

func submissionText(text string, attachments []Attachment) string {
	out := strings.TrimSpace(text)
	for _, attachment := range attachments {
		chip := fmt.Sprintf("[%s %d bytes]", attachment.Name, len(attachment.Content))
		if out == "" {
			out = chip
		} else {
			out += " " + chip
		}
	}
	return out
}

func isLargePaste(text string) bool {
	return len(text) >= LargePasteThreshold || strings.Contains(text, "\n")
}

func trimLastRune(text string) string {
	if text == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(text)
	return text[:len(text)-size]
}

func redact(text string, secrets []string) string {
	out := text
	for _, secret := range secrets {
		out = security.RedactSecret(out, secret)
	}
	return out
}

func bubbleEvent(msg tea.Msg) (Event, bool) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		return Event{Width: typed.Width, Height: typed.Height}, true
	case tea.KeyMsg:
		if len(typed.Runes) > 0 {
			return Event{Text: string(typed.Runes)}, true
		}
		return Event{Key: bubbleKey(typed.String())}, true
	default:
		return Event{}, false
	}
}

func bubbleKey(key string) string {
	switch strings.ToLower(key) {
	case KeyEnter, KeyBackspace, KeyUp, KeyDown, KeyCtrlC, KeyCtrlD, KeyEsc:
		return key
	case KeyTab:
		return KeyTab
	default:
		return key
	}
}

func commandMatches(prefix string) []commands.Command {
	var out []commands.Command
	for _, cmd := range commands.Registry() {
		if strings.HasPrefix(cmd.Name, prefix) {
			out = append(out, cmd)
		}
	}
	return out
}

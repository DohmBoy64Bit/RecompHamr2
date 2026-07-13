package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
func (b BubbleModel) View() tea.View {
	view := tea.NewView(b.State.RenderStyled())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.ReportFocus = true
	view.WindowTitle = "RecompHamr"
	view.Cursor = tea.NewCursor(styledComposerCursorX(b.State), styledComposerCursorY(b.State))
	view.Cursor.Shape = tea.CursorBar
	view.Cursor.Blink = true
	return view
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

// RenderStyled returns the ANSI-styled Bubble Tea view content.
func (m Model) RenderStyled() string {
	return m.RenderStyledWithLayout(m.Layout)
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

// RenderStyledWithLayout returns styled content for Bubble Tea rendering.
func (m Model) RenderStyledWithLayout(layout Layout) string {
	if layout.Width <= 0 {
		layout.Width = DefaultWidth
	}
	if layout.Height <= 0 {
		layout.Height = DefaultHeight
	}
	return m.renderBubble(layout)
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
	return m.renderPlain(layout, false)
}

func (m Model) renderCompact(layout Layout) string {
	return m.renderPlain(layout, true)
}

func (m Model) renderPlain(layout Layout, compact bool) string {
	var b strings.Builder
	if len(m.Transcript) == 0 {
		b.WriteString(startupPlain(m, layout, compact))
	} else {
		b.WriteString(chatPlain(m, layout, compact))
	}
	if len(m.PaletteRows()) > 0 {
		b.WriteString("\n")
		b.WriteString(palettePlain(m, layout))
	}
	return strings.TrimRight(b.String(), "\n")
}

func startupPlain(m Model, layout Layout, compact bool) string {
	width := renderWidth(layout.Width)
	brand := brandWide
	if compact {
		brand = brandCompact
	}
	lines := []string{
		centerText(width, brand),
		centerText(width, domainLine),
		"",
		centerText(width, composerPrompt(m)),
		centerText(width, statusBar(layout)),
		centerText(width, "/ commands   Tab complete   Ctrl+C cancel/quit   Ctrl+D exit"),
		"",
		centerText(width, "Tip: keep evidence in .rehamr/REPHAMR_STATE.md"),
	}
	return strings.Join(lines, "\n")
}

func chatPlain(m Model, layout Layout, compact bool) string {
	width := renderWidth(layout.Width)
	var b strings.Builder
	for _, line := range visibleTranscript(m.Transcript, 8) {
		fmt.Fprintf(&b, "%s\n", transcriptCard(line, width, compact))
	}
	if m.Status != "" {
		fmt.Fprintf(&b, "%s\n", transcriptCard("status: "+m.Status, width, compact))
	}
	if m.DebugEnabled && len(m.DebugLog) > 0 {
		fmt.Fprintf(&b, "%s\n", transcriptCard("status: debug "+m.DebugLog[len(m.DebugLog)-1], width, compact))
	}
	fmt.Fprintf(&b, "\n%s\n", composerPrompt(m))
	fmt.Fprintf(&b, "%s\n", statusBar(layout))
	fmt.Fprintf(&b, "/ commands   Tab complete   Ctrl+C cancel/quit   Ctrl+D exit")
	return b.String()
}

func palettePlain(m Model, layout Layout) string {
	width := renderWidth(layout.Width)
	rows := m.PaletteRows()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", centerText(width, "Command Palette"))
	for _, row := range rows {
		fmt.Fprintf(&b, "%s\n", centerText(width, row))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderBubble(layout Layout) string {
	if len(m.Transcript) == 0 {
		return m.renderBubbleStartup(layout)
	}
	return m.renderBubbleChat(layout)
}

func (m Model) renderBubbleStartup(layout Layout) string {
	width, height := bubbleSize(layout)
	panelWidth := launcherPanelWidth(width)
	panel := lipgloss.JoinVertical(
		lipgloss.Left,
		tuiStyleLogo(panelWidth).Render(brandWide),
		tuiStyleMuted().Width(panelWidth).Align(lipgloss.Center).Render(domainLine),
		"",
		tuiStyleComposerPanel(panelWidth).Render(composerPrompt(m)+"\n\n"+statusBar(layout)),
		tuiStyleHints().Width(panelWidth).Render("Tab complete    / commands    Ctrl+C cancel/quit    Ctrl+D exit"),
		"",
		tuiStyleTip().Width(panelWidth).Render("Tip: use /init-re to create reverse-engineering memory before long sessions."),
	)
	if len(m.PaletteRows()) > 0 {
		panel = lipgloss.JoinVertical(lipgloss.Left, paletteBubble(m, width), panel)
	}
	top := launcherTopPadding(height)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, strings.Repeat("\n", top)+panel)
}

func (m Model) renderBubbleChat(layout Layout) string {
	width, height := bubbleSize(layout)
	bodyHeight := height - 5
	if bodyHeight < 6 {
		bodyHeight = 6
	}
	body := transcriptBubble(m, width, bodyHeight)
	composer := tuiStyleComposerPanel(width).Render(composerPrompt(m) + "\n\n" + statusBar(layout))
	footer := tuiStyleHints().Render("/ commands    Tab complete    Ctrl+C cancel/quit    Ctrl+D exit")
	content := lipgloss.JoinVertical(lipgloss.Left, body, composer, footer)
	if len(m.PaletteRows()) > 0 {
		content = overlayPalette(content, paletteBubble(m, width), width)
	}
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content)
}

func transcriptBubble(m Model, width int, height int) string {
	var lines []string
	for _, line := range visibleTranscript(m.Transcript, height) {
		lines = append(lines, styleTranscriptLine(line, width))
	}
	if m.Status != "" {
		lines = append(lines, styleTranscriptLine("status: "+m.Status, width))
	}
	if len(lines) == 0 {
		lines = append(lines, tuiStyleMuted().Render("No transcript yet."))
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(strings.Join(lines, "\n\n"))
}

func paletteBubble(m Model, width int) string {
	rows := m.PaletteRows()
	panelWidth := width - 28
	if panelWidth < 44 {
		panelWidth = width
	}
	if panelWidth > 76 {
		panelWidth = 76
	}
	var body []string
	body = append(body, tuiStylePaletteTitle().Render("COMMAND PALETTE")+"  "+tuiStyleMuted().Render("esc"))
	for _, row := range rows {
		if strings.HasPrefix(row, ">") {
			body = append(body, tuiStyleSelected().Width(panelWidth-4).Render(strings.TrimSpace(row[1:])))
			continue
		}
		body = append(body, tuiStylePaletteRow().Render(strings.TrimSpace(row)))
	}
	return tuiStyleOverlay(panelWidth).Render(strings.Join(body, "\n"))
}

func overlayPalette(content string, palette string, width int) string {
	return lipgloss.JoinVertical(lipgloss.Center, palette, content)
}

func styleTranscriptLine(line string, width int) string {
	label := strings.Fields(transcriptBlock(line))[0]
	text := transcriptBlock(line)
	style := tuiStyleTranscript(width)
	switch label {
	case "user":
		style = tuiStyleUser(width)
	case "assistant":
		style = tuiStyleAssistant(width)
	case "tool", "mcp":
		style = tuiStyleTool(width)
	case "blocked":
		style = tuiStyleBlockedCard(width)
	case "unsupported", "unverified":
		style = tuiStyleWarningCard(width)
	}
	return style.Render(text)
}

func composerPrompt(m Model) string {
	text := composerView(m)
	if strings.TrimSpace(text) == "" {
		return `Ask RecompHamr... "map this function"`
	}
	return "composer > " + text
}

func statusBar(layout Layout) string {
	return fmt.Sprintf("Build * %s * %s * skill %s * mcp %s * context %s", layout.ActiveModel, layout.Mode, layout.ActiveSkill, layout.MCPStatus, layout.ContextStatus)
}

func visibleTranscript(lines []string, limit int) []string {
	if limit <= 0 || len(lines) <= limit {
		return append([]string(nil), lines...)
	}
	return append([]string(nil), lines[len(lines)-limit:]...)
}

func transcriptCard(line string, width int, compact bool) string {
	text := transcriptBlock(line)
	if compact && len(text) > width {
		return text[:width]
	}
	return text
}

func centerText(width int, text string) string {
	if len(text) >= width {
		return text
	}
	return strings.Repeat(" ", (width-len(text))/2) + text
}

func bubbleSize(layout Layout) (int, int) {
	width := layout.Width
	if width <= 0 {
		width = DefaultWidth
	}
	height := layout.Height
	if height <= 0 {
		height = DefaultHeight
	}
	if width < 40 {
		width = 40
	}
	return width, height
}

func launcherPanelWidth(width int) int {
	panelWidth := width - 16
	if panelWidth > 84 {
		panelWidth = 84
	}
	if panelWidth < 44 {
		panelWidth = width - 4
	}
	if panelWidth < 36 {
		panelWidth = 36
	}
	return panelWidth
}

func launcherTopPadding(height int) int {
	if height <= 18 {
		return 1
	}
	top := height / 5
	if top > 6 {
		return 6
	}
	return top
}

func chip(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "unverified"
	}
	return "[" + text + "]"
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

func renderWidth(width int) int {
	if width <= 0 {
		width = DefaultWidth
	}
	if width > 110 {
		width = 110
	}
	if width < 32 {
		width = 32
	}
	return width
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
	case tea.KeyPressMsg:
		key := typed.Key()
		if key.Text != "" {
			return Event{Text: key.Text}, true
		}
		return Event{Key: bubbleKey(typed.String())}, true
	case tea.PasteMsg:
		return Event{Paste: typed.Content}, true
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

func tuiStyleSelected() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Bold(true)
}

func tuiStyleLogo(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("214")).Bold(true).Align(lipgloss.Center)
}

func tuiStyleMuted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
}

func tuiStyleHints() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
}

func tuiStyleTip() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
}

func tuiStyleComposerPanel(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("45")).
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("235"))
}

func tuiStyleTranscript(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width - 6).MarginLeft(2).Foreground(lipgloss.Color("250"))
}

func tuiStyleUser(width int) lipgloss.Style {
	return tuiStyleTranscript(width).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("45")).PaddingLeft(1)
}

func tuiStyleAssistant(width int) lipgloss.Style {
	return tuiStyleTranscript(width).Foreground(lipgloss.Color("120")).PaddingLeft(3)
}

func tuiStyleTool(width int) lipgloss.Style {
	return tuiStyleTranscript(width).Foreground(lipgloss.Color("109")).PaddingLeft(3)
}

func tuiStyleBlockedCard(width int) lipgloss.Style {
	return tuiStyleTranscript(width).Foreground(lipgloss.Color("196")).Bold(true).PaddingLeft(3)
}

func tuiStyleWarningCard(width int) lipgloss.Style {
	return tuiStyleTranscript(width).Foreground(lipgloss.Color("220")).Bold(true).PaddingLeft(3)
}

func tuiStylePaletteTitle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
}

func tuiStylePaletteRow() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
}

func tuiStyleOverlay(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Border(lipgloss.NormalBorder(), false, true, false, true).
		BorderForeground(lipgloss.Color("238"))
}

func composerCursorX(m Model) int {
	lines := strings.Split(composerView(m), "\n")
	return len("composer > ") + len(lines[len(lines)-1])
}

func composerCursorY(m Model) int {
	return strings.Count(m.Render(), "\n")
}

func styledComposerCursorX(m Model) int {
	if len(m.Transcript) == 0 {
		width, _ := bubbleSize(m.Layout)
		panelWidth := launcherPanelWidth(width)
		panelLeft := (width - panelWidth) / 2
		return panelLeft + 4 + len(`Ask RecompHamr... "map this function"`)
	}
	return composerCursorX(m)
}

func styledComposerCursorY(m Model) int {
	if len(m.Transcript) == 0 {
		_, height := bubbleSize(m.Layout)
		return launcherTopPadding(height) + 4
	}
	return composerCursorY(m)
}

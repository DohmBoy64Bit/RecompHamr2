package tui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

type theme struct {
	brand    lipgloss.Style
	accent   lipgloss.Style
	body     lipgloss.Style
	muted    lipgloss.Style
	success  lipgloss.Style
	warning  lipgloss.Style
	blocked  lipgloss.Style
	composer lipgloss.Style
}

func makeTheme(profile colorprofile.Profile) theme {
	complete := lipgloss.Complete(profile)
	semanticColor := func(ansi, ansi256, truecolor string) color.Color {
		return complete(lipgloss.Color(ansi), lipgloss.Color(ansi256), lipgloss.Color(truecolor))
	}
	return theme{
		brand:    lipgloss.NewStyle().Foreground(semanticColor("3", "208", "#ff9d2e")).Bold(true),
		accent:   lipgloss.NewStyle().Foreground(semanticColor("6", "44", "#39d9e6")),
		body:     lipgloss.NewStyle().Foreground(semanticColor("7", "252", "#d7d7d7")),
		muted:    lipgloss.NewStyle().Foreground(semanticColor("8", "245", "#8b9490")),
		success:  lipgloss.NewStyle().Foreground(semanticColor("2", "78", "#73d987")),
		warning:  lipgloss.NewStyle().Foreground(semanticColor("3", "221", "#e6c75a")),
		blocked:  lipgloss.NewStyle().Foreground(semanticColor("1", "203", "#ff6b6b")).Bold(true),
		composer: lipgloss.NewStyle().Background(semanticColor("0", "235", "#171b19")).Foreground(semanticColor("7", "254", "#e5e7e6")),
	}
}

type frame struct {
	content    string
	cursorX    int
	cursorY    int
	hideCursor bool
}

// View renders one declarative Bubble Tea frame.
func (m Model) View() tea.View {
	m.resize()
	rendered := m.renderFrame()
	if m.profile == colorprofile.ASCII {
		rendered.content = ansi.Strip(rendered.content)
	}
	view := tea.NewView(rendered.content)
	view.AltScreen = true
	view.ReportFocus = true
	view.WindowTitle = "RecompHamr"
	if len(m.entries) > 0 {
		view.MouseMode = tea.MouseModeCellMotion
	}
	if !rendered.hideCursor {
		view.Cursor = tea.NewCursor(rendered.cursorX, rendered.cursorY)
		view.Cursor.Shape = tea.CursorBar
		view.Cursor.Blink = true
	}
	return view
}

// Render returns deterministic content through the live component tree.
func Render(snapshot Snapshot, entries []TranscriptEntry, width, height int) string {
	model := New(snapshot)
	model.width, model.height = width, height
	model.resize()
	model.appendTranscript(entries)
	return model.renderFrame().content
}

func (m *Model) resize() {
	if m.width <= 0 {
		m.width = DefaultWidth
	}
	if m.height <= 0 {
		m.height = DefaultHeight
	}
	lane := m.laneWidth()
	m.composer.SetWidth(max(20, lane-4))
	m.help.SetWidth(lane)
	composerHeight := max(3, m.composer.Height()+2)
	overlayHeight := 0
	if m.overlay != overlayNone {
		overlayHeight = min(16, max(7, m.height/2))
		m.picker.SetSize(lane, overlayHeight)
	}
	viewportHeight := m.height - composerHeight - overlayHeight - 5
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	m.transcript.SetWidth(lane)
	m.transcript.SetHeight(viewportHeight)
	m.transcript.SetContent(renderTranscript(m.entries, lane, m.profile))
}

func (m Model) laneWidth() int {
	if m.width > 120 {
		return 112
	}
	if m.width >= 80 {
		return max(40, m.width-8)
	}
	return max(40, m.width-4)
}

func (m Model) renderFrame() frame {
	if m.width < MinimumWidth || m.height < MinimumHeight {
		text := fmt.Sprintf("RecompHamr needs at least %dx%d  current %dx%d\nCtrl+D exit", MinimumWidth, MinimumHeight, m.width, m.height)
		return frame{content: lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, text), hideCursor: true}
	}
	canvas := make([]string, m.height)
	lane := m.laneWidth()
	x := max(0, (m.width-lane)/2)
	if len(m.entries) == 0 {
		return m.renderStartup(canvas, x, lane)
	}
	return m.renderChat(canvas, x, lane)
}

func (m Model) renderStartup(canvas []string, x, lane int) frame {
	t := makeTheme(m.profile)
	brand := wordmark(t, m.width, m.profile)
	brandHeight := lipgloss.Height(brand)
	top := max(1, (m.height-(brandHeight+10))/3)
	putCentered(canvas, top, m.width, brand)
	putCentered(canvas, top+brandHeight, m.width, t.muted.Render("RECOMP HAMR  ·  evidence-backed reconstruction"))
	composerY := top + brandHeight + 3
	putBlock(canvas, composerY, x, composerBlock(m, lane, t))
	hints := t.muted.Render(m.help.View(m.keys))
	putCentered(canvas, composerY+3, m.width, hints)
	if memoryNeedsInit(m.snapshot.MemoryStatus) {
		putCentered(canvas, composerY+5, m.width, t.warning.Render("Tip: /init-re creates project memory."))
	}
	cursor := m.composer.Cursor()
	return frameWithCursor(finishCanvas(canvas, m.width), cursor, x+2, composerY, false)
}

func (m Model) renderChat(canvas []string, x, lane int) frame {
	t := makeTheme(m.profile)
	composer := composerBlock(m, lane, t)
	composerLines := strings.Split(composer, "\n")
	helpLine := t.muted.Render(m.help.View(m.keys))
	statusLines := 0
	feedback := strings.TrimSpace(m.snapshot.Status)
	if m.newOutput {
		feedback = "new output  PgDn to follow"
	}
	if feedback != "" {
		statusLines = 1
	}
	composerY := m.height - len(composerLines) - 1 - statusLines
	putBlock(canvas, composerY, x, composer)
	if statusLines == 1 {
		putBlock(canvas, composerY+len(composerLines), x, t.warning.Render(truncate(feedback, lane)))
	}
	putBlock(canvas, m.height-1, x, helpLine)
	bodyBottom := composerY
	if m.overlay != overlayNone {
		overlay := m.picker.View()
		overlayLines := strings.Split(overlay, "\n")
		overlayY := max(0, composerY-len(overlayLines))
		putBlock(canvas, overlayY, x, overlay)
		bodyBottom = overlayY
	}
	m.transcript.SetHeight(max(3, bodyBottom-2))
	putBlock(canvas, 1, x, m.transcript.View())
	cursor := m.composer.Cursor()
	return frameWithCursor(finishCanvas(canvas, m.width), cursor, x+2, composerY, m.overlay != overlayNone)
}

func frameWithCursor(content string, cursor *tea.Cursor, offsetX, offsetY int, hide bool) frame {
	result := frame{content: content, hideCursor: hide}
	if cursor == nil {
		result.hideCursor = true
		return result
	}
	result.cursorX = offsetX + cursor.Position.X
	result.cursorY = offsetY + cursor.Position.Y
	return result
}

func composerBlock(m Model, width int, t theme) string {
	value := m.composer.View()
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = t.accent.Render("│") + " " + t.composer.Width(max(1, width-2)).Render(truncate(line, width-2))
	}
	state := strings.TrimSpace(strings.Join([]string{m.snapshot.ActiveModel, readiness(m.snapshot)}, "  "))
	lines = append(lines, t.accent.Render("│")+" "+t.muted.Render(truncate(state, width-2)))
	return strings.Join(lines, "\n")
}

func wordmark(t theme, width int, profile colorprofile.Profile) string {
	if width < CompactWidth || profile == colorprofile.ASCII {
		return t.body.Bold(true).Render("RECOMP") + " " + t.brand.Render("HAMR")
	}
	glyphs := map[rune][]string{
		'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
		'E': {"█████", "█    ", "████ ", "█    ", "█████"},
		'C': {" ████", "█    ", "█    ", "█    ", " ████"},
		'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
		'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
		'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
		'H': {"█   █", "█   █", "█████", "█   █", "█   █"},
		'A': {" ███ ", "█   █", "█████", "█   █", "█   █"},
	}
	renderPart := func(text string, row int, style lipgloss.Style) string {
		parts := make([]string, 0, len(text))
		for _, letter := range text {
			parts = append(parts, glyphs[letter][row])
		}
		return style.Render(strings.Join(parts, " "))
	}
	lines := make([]string, 5)
	for row := range lines {
		lines[row] = renderPart("RECOMP", row, t.body.Bold(true)) + "   " + renderPart("HAMR", row, t.brand)
	}
	return strings.Join(lines, "\n")
}

func readiness(snapshot Snapshot) string {
	if snapshot.PendingTool != "" && snapshot.PendingTool != "none" {
		return "working"
	}
	if snapshot.Mode == "" {
		return "ready"
	}
	return snapshot.Mode
}

func memoryNeedsInit(status string) bool {
	lower := strings.ToLower(status)
	return strings.Contains(lower, "missing") || strings.Contains(lower, "unsupported")
}

func renderTranscript(entries []TranscriptEntry, width int, profile colorprofile.Profile) string {
	t := makeTheme(profile)
	var blocks []string
	for _, entry := range entries {
		labelStyle := t.muted
		switch entry.Kind {
		case TranscriptUser:
			labelStyle = t.accent
		case TranscriptAssistant:
			labelStyle = t.success
		case TranscriptWarning, TranscriptUnsupported:
			labelStyle = t.warning
		case TranscriptBlocked:
			labelStyle = t.blocked
		case TranscriptVerified:
			labelStyle = t.success
		}
		label := labelStyle.Render(fmt.Sprintf("%-12s", string(entry.Kind)))
		bodyWidth := max(10, width-14)
		body := wrapText(entry.Text, bodyWidth)
		lines := strings.Split(body, "\n")
		if (entry.Kind == TranscriptTool || entry.Kind == TranscriptMCP) && len(lines) > 12 {
			lines = append(lines[:11], "output truncated")
		}
		for i, line := range lines {
			if i == 0 {
				lines[i] = label + line
			} else {
				lines[i] = strings.Repeat(" ", 12) + line
			}
		}
		blocks = append(blocks, strings.Join(lines, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

func wrapText(text string, width int) string {
	var out []string
	for _, source := range strings.Split(text, "\n") {
		for lipgloss.Width(source) > width {
			part := ansi.Truncate(source, width, "")
			out = append(out, part)
			source = strings.TrimSpace(strings.TrimPrefix(source, part))
			if part == "" {
				break
			}
		}
		out = append(out, source)
	}
	return strings.Join(out, "\n")
}

func truncate(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	}
	return ansi.Truncate(text, max(1, width), "…")
}

func putCentered(canvas []string, y, width int, text string) {
	x := max(0, (width-lipgloss.Width(text))/2)
	putBlock(canvas, y, x, text)
}

func putBlock(canvas []string, y, x int, block string) {
	for index, line := range strings.Split(block, "\n") {
		row := y + index
		if row < 0 || row >= len(canvas) {
			continue
		}
		canvas[row] = strings.Repeat(" ", max(0, x)) + line
	}
}

func finishCanvas(canvas []string, width int) string {
	for index, line := range canvas {
		canvas[index] = truncate(line, width)
	}
	return strings.Join(canvas, "\n")
}

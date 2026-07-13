package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func transcriptText(m Model) string {
	var lines []string
	for _, line := range m.Transcript {
		lines = append(lines, transcriptBlock(m.redactVisible(line)))
	}
	if m.Status != "" {
		lines = append(lines, transcriptBlock(m.redactVisible("status: "+m.Status)))
	}
	return strings.Join(lines, "\n\n")
}

func transcriptBubble(m Model, width int, height int) string {
	return transcriptBubbleWithTheme(m, width, height, defaultTheme())
}

func transcriptBubbleWithTheme(m Model, width int, height int, theme tuiTheme) string {
	var lines []string
	entryLimit := height / 3
	if entryLimit < 1 {
		entryLimit = 1
	}
	for _, line := range visibleTranscriptWindow(m.Transcript, entryLimit, m.TranscriptOffset) {
		lines = append(lines, styleTranscriptLineWithTheme(m.redactVisible(line), width, theme))
	}
	if m.NewOutput {
		lines = append(lines, theme.warning(width).Render("new output  PgDn to follow"))
	}
	if m.Status != "" {
		lines = append(lines, styleTranscriptLineWithTheme(m.redactVisible("status: "+m.Status), width, theme))
	}
	if len(lines) == 0 {
		lines = append(lines, theme.muted().Render("No transcript yet."))
	}
	return theme.transcriptFrame(width, height).Render(strings.Join(lines, "\n\n"))
}

func styleTranscriptLine(line string, width int) string {
	return styleTranscriptLineWithTheme(line, width, defaultTheme())
}

func styleTranscriptLineWithTheme(line string, width int, theme tuiTheme) string {
	label := strings.Fields(transcriptBlock(line))[0]
	text := boundedTranscriptBlock(line, width)
	style := theme.transcript(width)
	switch label {
	case "user":
		style = theme.user(width)
	case "assistant":
		style = theme.assistant(width)
	case "tool", "mcp":
		style = theme.tool(width)
	case "blocked":
		style = theme.blocked(width)
	case "warning", "unsupported", "unverified", "status":
		style = theme.warning(width)
	case "verification":
		style = theme.assistant(width)
	}
	return style.Render(text)
}

func visibleTranscriptWindow(lines []string, limit int, offset int) []string {
	end := len(lines) - offset
	if end < 0 {
		end = 0
	}
	start := end - limit
	if start < 0 {
		start = 0
	}
	return append([]string(nil), lines[start:end]...)
}

func boundedTranscriptBlock(line string, width int) string {
	const maximumLines = 12
	parts := strings.Split(transcriptBlock(line), "\n")
	truncated := len(parts) > maximumLines
	if truncated {
		parts = parts[:maximumLines]
	}
	lineWidth := width - 8
	if lineWidth < 20 {
		lineWidth = 20
	}
	for i := range parts {
		parts[i] = ansi.Truncate(parts[i], lineWidth, "…")
	}
	if truncated {
		parts = append(parts, "output truncated")
	}
	return strings.Join(parts, "\n")
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
		return ansi.Truncate(text, width, "…")
	}
	return text
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
	case strings.HasPrefix(line, "verification:") || strings.HasPrefix(line, "verified:"):
		label = "verification"
	case strings.HasPrefix(line, "blocked:"):
		label = "blocked"
	case strings.HasPrefix(line, "unsupported:"):
		label = "unsupported"
	case strings.HasPrefix(line, "unverified:"):
		label = "unverified"
	case strings.HasPrefix(line, "warning:"):
		label = "warning"
	case strings.HasPrefix(line, "status:"):
		label = "status"
	case strings.HasPrefix(line, "paste:"):
		label = "attachment"
	}
	return fmt.Sprintf("%-11s %s", label, line)
}

func (m Model) redactVisible(text string) string {
	return redact(text, m.DebugSecrets)
}

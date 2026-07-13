package tui

import (
	"fmt"
	"strings"
)

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
	if rows := m.overlayRows(); len(rows) > 0 {
		b.WriteString("\n")
		b.WriteString(overlayPlain(m.overlayKind(), rows, layout))
	}
	return strings.TrimRight(b.String(), "\n")
}

func startupPlain(m Model, layout Layout, compact bool) string {
	width := renderWidth(layout.Width)
	brand := brandWide
	domain := "evidence-backed reconstruction"
	if compact {
		brand = brandCompact
		domain = "RE / decomp / recomp"
	}
	lines := []string{
		centerText(width, brand),
		centerText(width, domain),
		"",
		centerText(width, composerPrompt(m)),
		centerText(width, startupStatus(layout)),
		centerText(width, "tab complete   / commands   ? help"),
	}
	if tip := actionableStartupTip(layout); tip != "" {
		lines = append(lines, "", centerText(width, tip))
	}
	return strings.Join(lines, "\n")
}

func chatPlain(m Model, layout Layout, compact bool) string {
	width := renderWidth(layout.Width)
	var b strings.Builder
	for _, line := range visibleTranscript(m.Transcript, 8) {
		fmt.Fprintf(&b, "%s\n", transcriptCard(m.redactVisible(line), width, compact))
	}
	if m.Status != "" {
		fmt.Fprintf(&b, "%s\n", transcriptCard(m.redactVisible("status: "+m.Status), width, compact))
	}
	if m.DebugEnabled && len(m.DebugLog) > 0 {
		fmt.Fprintf(&b, "%s\n", transcriptCard(m.redactVisible("status: debug "+m.DebugLog[len(m.DebugLog)-1]), width, compact))
	}
	fmt.Fprintf(&b, "\n%s\n", composerPrompt(m))
	fmt.Fprintf(&b, "%s\n", statusBar(layout))
	fmt.Fprintf(&b, "/ commands   Tab complete   Ctrl+C cancel/quit   Ctrl+D exit")
	return b.String()
}

func overlayPlain(kind string, rows []string, layout Layout) string {
	width := renderWidth(layout.Width)
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", centerText(width, titleCase(overlayTitle(kind))))
	for _, row := range rows {
		fmt.Fprintf(&b, "%s\n", centerText(width, row))
	}
	return strings.TrimRight(b.String(), "\n")
}

func titleCase(text string) string {
	parts := strings.Fields(strings.ToLower(text))
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

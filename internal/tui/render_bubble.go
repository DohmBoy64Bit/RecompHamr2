package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderStyled returns the ANSI-styled Bubble Tea view content.
func (m Model) RenderStyled() string {
	return m.RenderStyledWithLayout(m.Layout)
}

// RenderStyledWithLayout returns styled content for Bubble Tea rendering.
func (m Model) RenderStyledWithLayout(layout Layout) string {
	return m.styledScreen(layout).content
}

func (m Model) renderBubbleStartup(layout Layout) string {
	width, height := bubbleSize(layout)
	theme := newTUITheme(layout.ColorProfile)
	panelWidth := launcherPanelWidth(width)
	groups := []string{
		startupWordmark(width, panelWidth, theme),
		theme.muted().Width(panelWidth).Align(lipgloss.Center).Render(startupDomain(width)),
		"",
	}
	if rows := m.overlayRows(); len(rows) > 0 {
		groups = append(groups, overlayBubbleWithTheme(m, m.overlayKind(), rows, m.PaletteIndex, panelWidth, theme))
	}
	groups = append(groups,
		theme.composer(panelWidth).Render(startupComposerWithTheme(m, layout, theme)),
		theme.hints().Width(panelWidth).Align(lipgloss.Center).Render("tab complete    / commands    ? help"),
	)
	if tip := actionableStartupTip(layout); tip != "" {
		groups = append(groups, "", theme.tip().Width(panelWidth).Align(lipgloss.Center).Render(tip))
	}
	panel := lipgloss.JoinVertical(
		lipgloss.Left,
		groups...,
	)
	top := launcherTopPadding(height)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, strings.Repeat("\n", top)+panel)
}

func (m Model) renderBubbleChat(layout Layout) string {
	width, height := bubbleSize(layout)
	theme := newTUITheme(layout.ColorProfile)
	bodyHeight := height - 5
	body := transcriptBubbleWithTheme(m, width, bodyHeight, theme)
	composer := theme.composer(width).Render(composerPrompt(m) + "\n\n" + startupStatus(layout))
	footer := theme.hints().Render("/ commands    PgUp/PgDn scroll    Ctrl+C cancel")
	parts := []string{body}
	if rows := m.overlayRows(); len(rows) > 0 {
		parts = append(parts, overlayBubbleWithTheme(m, m.overlayKind(), rows, m.PaletteIndex, width, theme))
	}
	parts = append(parts, composer, footer)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content)
}

func paletteBubble(m Model, width int) string {
	return overlayBubble(m, "commands", m.PaletteRows(), m.PaletteIndex, width)
}

func overlayBubble(m Model, kind string, rows []string, selected int, width int) string {
	return overlayBubbleWithTheme(m, kind, rows, selected, width, defaultTheme())
}

func overlayBubbleWithTheme(m Model, kind string, rows []string, selected int, width int, theme tuiTheme) string {
	panelWidth := width - 28
	if panelWidth < 44 {
		panelWidth = width
	}
	if panelWidth > 76 {
		panelWidth = 76
	}
	visible := 7
	if width < 80 {
		visible = 5
	}
	start := 0
	if selected >= visible {
		start = selected - visible + 1
	}
	end := start + visible
	if end > len(rows) {
		end = len(rows)
	}
	var body []string
	header := overlayTitle(kind) + "  " + fmt.Sprintf("%d/%d", len(rows), len(rows))
	body = append(body, theme.paletteTitle().Render(header)+"  "+theme.muted().Render("esc"))
	if len(rows) == 0 {
		body = append(body, theme.warning(panelWidth).Render("unverified: no matches"))
	}
	for i, row := range rows[start:end] {
		actual := start + i
		clean := strings.TrimSpace(strings.TrimLeft(row, ">*! "))
		if actual == selected || strings.HasPrefix(row, ">") {
			body = append(body, theme.selected().Width(panelWidth-4).Render(clean))
			continue
		}
		if strings.HasPrefix(row, "!") {
			body = append(body, theme.warning(panelWidth).Render(clean))
			continue
		}
		body = append(body, theme.paletteRow().Render(clean))
	}
	if kind == "commands" {
		body = append(body, theme.muted().Render(m.paletteDetail()))
	}
	return theme.overlay(panelWidth).Render(strings.Join(body, "\n"))
}

func startupWordmark(width int, panelWidth int, theme tuiTheme) string {
	if width < 80 {
		return theme.logo(panelWidth).Render(theme.brand().Render(brandCompact))
	}
	wordmark := lipgloss.NewStyle().Foreground(theme.color(7, "255", "#E6E6E6")).Bold(true).Render("RECOMP ") + theme.brand().Render("HAMR")
	return theme.logo(panelWidth).Render(wordmark)
}

func startupDomain(width int) string {
	if width < 80 {
		return "RE / decomp / recomp"
	}
	return "evidence-backed reconstruction"
}

func startupComposerWithTheme(m Model, layout Layout, theme tuiTheme) string {
	return strings.Join([]string{
		composerPrompt(m),
		"",
		theme.composerMeta().Render(startupStatus(layout)),
	}, "\n")
}

func startupStatus(layout Layout) string {
	return strings.TrimSpace(strings.Join([]string{layout.Mode, layout.ActiveModel, runtimeState(layout)}, "  "))
}

func runtimeState(layout Layout) string {
	if layout.PendingTool != "" && layout.PendingTool != "none" {
		return "working"
	}
	if layout.Mode == "thinking" || layout.Mode == "streaming" {
		return "working"
	}
	return "ready"
}

func actionableStartupTip(layout Layout) string {
	memory := strings.ToLower(layout.MemoryStatus)
	if strings.Contains(memory, "missing") || strings.Contains(memory, "unsupported") {
		return "Tip: /init-re creates project memory."
	}
	return ""
}

func overlayTitle(kind string) string {
	switch kind {
	case "models":
		return "MODEL PICKER"
	case "skills":
		return "SKILL PICKER"
	case "mcp":
		return "MCP CONTROLS"
	case "help":
		return "HELP"
	default:
		return "COMMAND PALETTE"
	}
}

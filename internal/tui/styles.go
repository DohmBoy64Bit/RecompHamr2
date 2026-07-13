package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

type tuiTheme struct {
	profile colorprofile.Profile
}

func newTUITheme(profile colorprofile.Profile) tuiTheme {
	if profile == colorprofile.Unknown {
		profile = colorprofile.ANSI256
	}
	return tuiTheme{profile: profile}
}

func (t tuiTheme) color(ansi int, ansi256 string, truecolor string) color.Color {
	return lipgloss.Complete(t.profile)(lipgloss.Color(string(rune('0'+ansi))), lipgloss.Color(ansi256), lipgloss.Color(truecolor))
}

func (t tuiTheme) selected() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true).Reverse(true)
	if t.profile > colorprofile.ASCII {
		style = style.Foreground(t.color(0, "16", "#101010")).Background(t.color(6, "45", "#21D4E8"))
	}
	return style
}

func (t tuiTheme) logo(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Bold(true).Align(lipgloss.Center)
}
func (t tuiTheme) brand() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(3, "214", "#FF9D2E")).Bold(true)
}
func (t tuiTheme) muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(7, "244", "#8A918A"))
}
func (t tuiTheme) hints() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(2, "120", "#72E06A")).Bold(true)
}
func (t tuiTheme) tip() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(3, "220", "#F5D547"))
}
func (t tuiTheme) composer(width int) lipgloss.Style {
	style := lipgloss.NewStyle().Width(width).Padding(0, 2).Border(lipgloss.NormalBorder(), false, false, false, true).Bold(false)
	if t.profile > colorprofile.ASCII {
		style = style.BorderForeground(t.color(6, "45", "#21D4E8")).Foreground(t.color(7, "255", "#E6E6E6")).Background(t.color(0, "235", "#171A18"))
	}
	return style
}
func (t tuiTheme) composerMeta() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(7, "250", "#C6C6C6"))
}
func (t tuiTheme) transcriptFrame(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Height(height)
}
func (t tuiTheme) transcript(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width - 6).MarginLeft(2).Foreground(t.color(7, "250", "#C6C6C6"))
}
func (t tuiTheme) user(width int) lipgloss.Style {
	return t.transcript(width).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(t.color(6, "45", "#21D4E8")).PaddingLeft(1)
}
func (t tuiTheme) assistant(width int) lipgloss.Style {
	return t.transcript(width).Foreground(t.color(2, "120", "#72E06A")).PaddingLeft(3)
}
func (t tuiTheme) tool(width int) lipgloss.Style {
	return t.transcript(width).Foreground(t.color(6, "109", "#66B8C2")).PaddingLeft(3)
}
func (t tuiTheme) blocked(width int) lipgloss.Style {
	return t.transcript(width).Foreground(t.color(1, "203", "#FF5C57")).Bold(true).PaddingLeft(3)
}
func (t tuiTheme) warning(width int) lipgloss.Style {
	return t.transcript(width).Foreground(t.color(3, "220", "#F5D547")).Bold(true).PaddingLeft(3)
}
func (t tuiTheme) paletteTitle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(2, "120", "#72E06A")).Bold(true)
}
func (t tuiTheme) paletteRow() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.color(7, "250", "#C6C6C6"))
}
func (t tuiTheme) overlay(width int) lipgloss.Style {
	style := lipgloss.NewStyle().Width(width).Padding(1, 2).Border(lipgloss.NormalBorder(), false, true, false, true)
	if t.profile > colorprofile.ASCII {
		style = style.Background(t.color(0, "235", "#171A18")).BorderForeground(t.color(7, "238", "#444A46"))
	}
	return style
}

func defaultTheme() tuiTheme { return newTUITheme(colorprofile.ANSI256) }

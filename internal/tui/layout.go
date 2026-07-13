package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func centerText(width int, text string) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	return strings.Repeat(" ", (width-textWidth)/2) + text
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

func renderWidth(width int) int {
	if width <= 0 {
		width = DefaultWidth
	}
	if width > 110 {
		return 110
	}
	if width < 32 {
		return 32
	}
	return width
}

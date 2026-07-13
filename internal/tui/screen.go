package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type styledScreen struct {
	content    string
	cursorX    int
	cursorY    int
	hideCursor bool
}

func (m Model) styledScreen(layout Layout) styledScreen {
	if layout.Width <= 0 {
		layout.Width = DefaultWidth
	}
	if layout.Height <= 0 {
		layout.Height = DefaultHeight
	}
	if terminalTooSmall(layout) {
		return tooSmallScreen(layout)
	}
	if len(m.Transcript) == 0 {
		return m.startupScreen(layout)
	}
	return m.chatScreen(layout)
}

func terminalTooSmall(layout Layout) bool {
	return layout.Width < MinimumWidth || layout.Height < MinimumHeight
}

func tooSmallScreen(layout Layout) styledScreen {
	width, height := bubbleSize(layout)
	theme := newTUITheme(layout.ColorProfile)
	message := "RecompHamr needs a larger terminal\n\n" +
		fmt.Sprintf("current %dx%d  required %dx%d", layout.Width, layout.Height, MinimumWidth, MinimumHeight) +
		"\n\nresize or Ctrl+D exit"
	return styledScreen{
		content:    lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, theme.warning(width).Render(message)),
		hideCursor: true,
	}
}

func (m Model) startupScreen(layout Layout) styledScreen {
	width, height := bubbleSize(layout)
	content := m.renderBubbleStartup(layout)
	panelWidth := launcherPanelWidth(width)
	panelLeft := (width - panelWidth) / 2
	cursorX := panelLeft + 3 + lipgloss.Width(composerPrompt(m))
	if cursorX >= width {
		cursorX = width - 2
	}
	return styledScreen{
		content: content,
		cursorX: cursorX,
		cursorY: launcherTopPadding(height) + 3,
	}
}

func (m Model) chatScreen(layout Layout) styledScreen {
	width, height := bubbleSize(layout)
	bodyHeight := height - 5
	content := m.renderBubbleChat(layout)
	cursorX := composerCursorX(m)
	if cursorX >= width {
		cursorX = width - 2
	}
	return styledScreen{
		content: content,
		cursorX: cursorX,
		cursorY: bodyHeight + 1,
	}
}

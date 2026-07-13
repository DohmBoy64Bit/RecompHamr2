package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func composerPrompt(m Model) string {
	text := composerView(m)
	if strings.TrimSpace(text) == "" {
		return `Ask RecompHamr... "map this function"`
	}
	return "composer > " + text
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

func composerCursorX(m Model) int {
	lines := strings.Split(composerView(m), "\n")
	return lipgloss.Width("composer > ") + lipgloss.Width(lines[len(lines)-1])
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

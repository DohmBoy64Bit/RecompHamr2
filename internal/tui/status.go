package tui

import (
	"fmt"
	"strings"
)

func statusBar(layout Layout) string {
	return fmt.Sprintf("Build * %s * %s * memory %s * skill %s * mcp %s * context %s", layout.ActiveModel, layout.Mode, layout.MemoryStatus, layout.ActiveSkill, layout.MCPStatus, layout.ContextStatus)
}

func chip(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "unverified"
	}
	return "[" + text + "]"
}

package tui

import (
	"fmt"
	"strings"

	"recomphamr2/internal/commands"
)

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
	if m.overlayKind() != "commands" {
		return nil
	}
	fields := strings.Fields(token + " ")
	prefix := fields[0]
	var rows []string
	for i, cmd := range commandMatches(prefix) {
		pointer := " "
		if i == m.PaletteIndex {
			pointer = ">"
		}
		rows = append(rows, fmt.Sprintf("%s %-14s %s", pointer, cmd.Name, cmd.Summary))
	}
	return rows
}

func (m Model) paletteDetail() string {
	fields := strings.Fields(strings.TrimSpace(m.Composer) + " ")
	if len(fields) == 0 {
		return ""
	}
	matches := commandMatches(fields[0])
	if len(matches) == 0 {
		return ""
	}
	index := m.PaletteIndex
	if index < 0 || index >= len(matches) {
		index = 0
	}
	return "usage: " + matches[index].Usage + "  side effects: " + matches[index].SideEffects
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

func commandMatches(prefix string) []commands.Command {
	var out []commands.Command
	for _, cmd := range commands.Registry() {
		if strings.HasPrefix(cmd.Name, prefix) {
			out = append(out, cmd)
		}
	}
	return out
}

func (m Model) overlayKind() string {
	trimmed := strings.TrimSpace(m.Composer)
	switch {
	case trimmed == "/models" || strings.HasPrefix(trimmed, "/models "):
		return "models"
	case trimmed == "/skills" || trimmed == "/skill" || strings.HasPrefix(trimmed, "/skill "):
		return "skills"
	case trimmed == "/mcp" || strings.HasPrefix(trimmed, "/mcp "):
		return "mcp"
	case trimmed == "/help" || strings.HasPrefix(trimmed, "/help "):
		return "help"
	case strings.HasPrefix(trimmed, "/"):
		return "commands"
	default:
		return ""
	}
}

func (m Model) overlayRows() []string {
	switch m.overlayKind() {
	case "commands":
		return m.PaletteRows()
	case "models":
		return pickerRows(commands.ModelPickerRows(m.Env))
	case "skills":
		return pickerRows(commands.SkillPickerRows(m.Env))
	case "mcp":
		return pickerRows(commands.MCPPickerRows(m.Env))
	case "help":
		return helpRows()
	default:
		return nil
	}
}

func pickerRows(rows []commands.PickerRow) []string {
	out := make([]string, 0, len(rows))
	appendRow := func(row commands.PickerRow) {
		marker := " "
		if row.Active {
			marker = "*"
		}
		if row.Blocked {
			marker = "!"
		}
		out = append(out, fmt.Sprintf("%s %-18s %s  %s", marker, row.Name, row.Summary, row.Detail))
	}
	for _, row := range rows {
		if row.Blocked {
			appendRow(row)
		}
	}
	for _, row := range rows {
		if !row.Blocked {
			appendRow(row)
		}
	}
	return out
}

func helpRows() []string {
	return []string{
		"  /                  open slash command palette  filters command registry",
		"  Tab                complete selected command   no side effects",
		"  Enter              submit composer             commands may mutate documented state",
		"  Ctrl+C             cancel or arm quit           no model work while idle",
		"  Ctrl+D             quit immediately            clean Bubble Tea exit",
	}
}

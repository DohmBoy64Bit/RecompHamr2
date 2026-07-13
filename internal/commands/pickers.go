package commands

import (
	"fmt"
	"sort"
	"strings"

	"recomphamr2/internal/mcp"
	"recomphamr2/internal/skills"
)

// PickerRow describes one read-only TUI picker row.
type PickerRow struct {
	// Name is the primary visible identifier.
	Name string
	// Summary is the short user-facing row text.
	Summary string
	// Detail is additional verified local state.
	Detail string
	// Active marks the current selection or active runtime state.
	Active bool
	// Blocked marks a row that reports unavailable local state.
	Blocked bool
}

// ModelPickerRows returns configured model rows without mutating config.
func ModelPickerRows(env Environment) []PickerRow {
	if env.Config == nil {
		return []PickerRow{{Name: "blocked", Summary: "config is not loaded", Detail: "run /doctor or /init-re", Blocked: true}}
	}
	names := env.Config.ModelNames()
	rows := make([]PickerRow, 0, len(names))
	for _, name := range names {
		profile := env.Config.Models[name]
		rows = append(rows, PickerRow{
			Name:    name,
			Summary: profile.LLM,
			Detail:  fmt.Sprintf("context %d url %s", profile.ContextSize, summarizeURL(profile.URL)),
			Active:  name == env.Config.Active,
		})
	}
	return rows
}

// SkillPickerRows returns embedded and custom skill rows without activating one.
func SkillPickerRows(env Environment) []PickerRow {
	active := activeSet(env.ActiveSkills)
	embedded := skills.Embedded()
	rows := make([]PickerRow, 0, len(embedded)+1)
	for _, skill := range embedded {
		rows = append(rows, PickerRow{Name: skill.Name, Summary: "available", Detail: skill.Source + " skill", Active: active[skill.Name]})
	}
	custom, err := skills.LoadCustom(env.CustomSkillsDir)
	if err != nil {
		rows = append(rows, PickerRow{Name: "blocked", Summary: "custom skills unavailable", Detail: err.Error(), Blocked: true})
	} else {
		for _, skill := range custom {
			rows = append(rows, PickerRow{Name: skill.Name, Summary: "available", Detail: skill.Source + " skill", Active: active[skill.Name]})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Blocked != rows[j].Blocked {
			return !rows[i].Blocked
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

// MCPPickerRows returns MCP server rows from the manager or built-in registry.
func MCPPickerRows(env Environment) []PickerRow {
	if env.MCP != nil {
		statuses := env.MCP.AllStatus()
		rows := make([]PickerRow, 0, len(statuses))
		for _, status := range statuses {
			detail := fmt.Sprintf("tools %d", status.Tools)
			if status.Err != "" {
				detail += " error " + status.Err
			}
			rows = append(rows, PickerRow{Name: status.Name, Summary: string(status.State), Detail: detail, Active: status.State == mcp.StateConnected, Blocked: status.State == mcp.StateError})
		}
		return rows
	}
	servers := mcp.Builtins()
	rows := make([]PickerRow, 0, len(servers))
	for _, server := range servers {
		detail := "stdio " + server.Command
		if server.URL != "" {
			detail = "http " + summarizeURL(server.URL)
		}
		if server.Autostart {
			detail += " autostart"
		}
		if server.RequireSkill {
			detail += " skill-gated"
		}
		rows = append(rows, PickerRow{Name: server.Name, Summary: "disconnected", Detail: detail})
	}
	return rows
}

func activeSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		out[strings.TrimSpace(name)] = true
	}
	return out
}

func summarizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "unverified"
	}
	if before, _, ok := strings.Cut(raw, "?"); ok {
		raw = before
	}
	if len(raw) > 48 {
		return raw[:45] + "..."
	}
	return raw
}

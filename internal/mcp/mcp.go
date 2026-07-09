package mcp

import (
	"os"
	"strings"
)

// ServerConfig describes one MCP server registration.
type ServerConfig struct {
	// Name is the stable server identifier.
	Name string
	// Command is the stdio command used to launch the server.
	Command string
	// Args are stdio command arguments.
	Args []string
	// URL is the streamable HTTP endpoint for the server.
	URL string
	// AllowedTools is the configured tool allowlist.
	AllowedTools []string
	// Autostart reports whether the server should connect automatically.
	Autostart bool
	// RequireSkill gates tools behind a matching active skill.
	RequireSkill bool
}

// Builtins returns RecompHamr built-in MCP server configs with env overrides.
func Builtins() []ServerConfig {
	servers := []ServerConfig{
		{Name: "ghidra", Command: "ghidra-mcp", RequireSkill: true},
		{Name: "n64-debug-mcp", Command: "n64-debug-mcp", RequireSkill: true},
		{Name: "pcrecomp", Command: "pcrecomp-mcp", RequireSkill: true},
		{Name: "mcp-pine", Command: "mcp-pine", RequireSkill: true},
		{Name: "objdiff", Command: "objdiff-mcp", RequireSkill: true},
		{Name: "pcsx2", Command: "pcsx2-mcp", RequireSkill: true},
		{Name: "bizhawk", Command: "bizhawk-mcp", RequireSkill: true},
		{Name: "sega2asm", Command: "sega2asm-mcp", RequireSkill: true},
	}
	for i := range servers {
		applyEnv(&servers[i])
	}
	return servers
}

// VisibleTools returns tools only when a matching skill is active.
func VisibleTools(server ServerConfig, activeSkills []string, tools []string) []string {
	for _, skill := range activeSkills {
		if SkillAllowsServer(skill, server.Name) {
			return append([]string(nil), tools...)
		}
	}
	return nil
}

// SkillAllowsServer reports whether skill unlocks server.
func SkillAllowsServer(skill string, server string) bool {
	skill = strings.ToLower(strings.TrimSpace(skill))
	server = strings.ToLower(strings.TrimSpace(server))
	if skill == "" || server == "" {
		return false
	}
	if skill == server {
		return true
	}
	return skillServerMap[skill] == server
}

func applyEnv(server *ServerConfig) {
	key := strings.ToUpper(strings.ReplaceAll(server.Name, "-", "_"))
	if command := os.Getenv("RECOMPHAMR_MCP_" + key + "_COMMAND"); command != "" {
		server.Command = command
	}
	if url := os.Getenv("RECOMPHAMR_MCP_" + key + "_URL"); url != "" {
		server.URL = url
	}
	if tools := os.Getenv("RECOMPHAMR_MCP_" + key + "_TOOLS"); tools != "" {
		server.AllowedTools = strings.Split(tools, ",")
	}
	server.Autostart = os.Getenv("RECOMPHAMR_MCP_AUTOSTART") == "1"
}

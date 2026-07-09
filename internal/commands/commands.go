package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"recomphamr2/internal/config"
	"recomphamr2/internal/doctor"
	"recomphamr2/internal/mcp"
	"recomphamr2/internal/parity"
	"recomphamr2/internal/project"
	"recomphamr2/internal/skills"
	"recomphamr2/internal/tools"
)

// Command documents one slash command.
type Command struct {
	// Name is the slash-prefixed command token.
	Name string
	// Summary is the one-line help text.
	Summary string
	// Usage is the accepted command syntax.
	Usage string
	// SideEffects documents filesystem, network, config, or state changes.
	SideEffects string
	// Examples are user-facing invocations.
	Examples []string
	// Errors documents expected non-success output classes.
	Errors []string
}

// Environment provides filesystem context for command execution.
type Environment struct {
	// ProjectDir is the workspace root for project commands.
	ProjectDir string
	// CustomSkillsDir is the directory used for user-authored skills.
	CustomSkillsDir string
	// Config is the loaded runtime config for model commands.
	Config *config.Config
	// ActiveSkills is the current skill activation list.
	ActiveSkills []string
	// MCP is the optional MCP manager used by /mcp lifecycle commands.
	MCP *mcp.Manager
	// FetchSkill fetches a remote skill body for /skill-new; nil uses HTTP(S).
	FetchSkill SkillFetcher
}

// SkillFetcher returns a remote skill document body for /skill-new.
type SkillFetcher func(rawURL string) (string, error)

var (
	commandMkdirAll  = os.MkdirAll
	commandWriteFile = os.WriteFile
)

// Registry returns all parity slash commands.
func Registry() []Command {
	out := make([]Command, len(commandCatalog))
	for i, cmd := range commandCatalog {
		out[i] = cmd
		out[i].Examples = append([]string(nil), cmd.Examples...)
		out[i].Errors = append([]string(nil), cmd.Errors...)
	}
	return out
}

// Help renders detailed command help.
func Help() string {
	cmds := Registry()
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	var b strings.Builder
	b.WriteString("RecompHamr slash commands\n")
	for _, cmd := range cmds {
		fmt.Fprintf(&b, "%s\t%s\tusage: %s\tside effects: %s\n", cmd.Name, cmd.Summary, cmd.Usage, cmd.SideEffects)
	}
	return b.String()
}

// Markdown renders command reference documentation from the registry.
func Markdown() string {
	var b strings.Builder
	b.WriteString("# Slash Command Reference\n\n")
	b.WriteString("Generated from `internal/commands.Registry`.\n\n")
	for _, cmd := range Registry() {
		fmt.Fprintf(&b, "## `%s`\n\n%s\n\n", cmd.Name, cmd.Summary)
		fmt.Fprintf(&b, "- Usage: `%s`\n", cmd.Usage)
		fmt.Fprintf(&b, "- Side effects: %s\n", cmd.SideEffects)
		if len(cmd.Examples) > 0 {
			b.WriteString("- Examples: ")
			for i, example := range cmd.Examples {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "`%s`", example)
			}
			b.WriteString("\n")
		}
		if len(cmd.Errors) > 0 {
			b.WriteString("- Errors: " + strings.Join(cmd.Errors, "; ") + "\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// Execute runs a supported slash command against env.
func Execute(env Environment, text string) (string, Environment) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "unsupported: empty command", env
	}
	if !strings.HasPrefix(fields[0], "/") {
		return "unsupported: slash command must start with /", env
	}
	switch fields[0] {
	case "/help":
		if len(fields) == 2 {
			return helpFor(fields[1]), env
		}
		if len(fields) > 2 {
			return "usage: /help [command]", env
		}
		return Help(), env
	case "/clear":
		return "conversation cleared", env
	case "/models":
		return runModels(env, fields)
	case "/skills":
		return runSkills(env)
	case "/skill":
		if len(fields) != 2 {
			return "usage: /skill <name>", env
		}
		skill, err := skills.Resolve(fields[1], env.CustomSkillsDir)
		if err != nil {
			return "unverified: " + err.Error(), env
		}
		env.ActiveSkills = append(env.ActiveSkills, skill.Name)
		return "loaded skill: " + skill.Name, env
	case "/skill-audit":
		if len(fields) != 2 {
			return "usage: /skill-audit <name>", env
		}
		return skills.Audit(fields[1]), env
	case "/skill-new":
		return runSkillNew(env, fields)
	case "/init-re":
		if _, err := project.Init(env.ProjectDir); err != nil {
			return "blocked: " + err.Error(), env
		}
		return "initialized .rehamr workspace", env
	case "/status-re":
		return project.Status(env.ProjectDir), env
	case "/doctor":
		return doctor.Run(doctor.Options{ProjectDir: env.ProjectDir, CustomSkillsDir: env.CustomSkillsDir}).String(), env
	case "/mcp":
		return runMCP(env, fields), env
	default:
		return "unsupported: unknown command " + fields[0], env
	}
}

func runModels(env Environment, fields []string) (string, Environment) {
	if env.Config == nil {
		return "blocked: config is not loaded", env
	}
	if len(fields) > 2 {
		return "usage: /models [name]", env
	}
	if len(fields) == 2 {
		if err := env.Config.SetActive(fields[1]); err != nil {
			return "unverified: " + err.Error(), env
		}
		return "active model: " + fields[1], env
	}
	names := make([]string, 0, len(env.Config.Models))
	for name := range env.Config.Models {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("models:")
	for _, name := range names {
		mark := " "
		if name == env.Config.Active {
			mark = " *"
		}
		fmt.Fprintf(&b, "\n%s %s", mark, name)
	}
	return b.String(), env
}

func runSkills(env Environment) (string, Environment) {
	embedded := skills.Embedded()
	custom, err := skills.LoadCustom(env.CustomSkillsDir)
	if err != nil {
		return "blocked: " + err.Error(), env
	}
	var b strings.Builder
	fmt.Fprintf(&b, "skills: %d embedded, %d custom, %d active, tools: %d built-in, commands: %d parity", len(embedded), len(custom), len(env.ActiveSkills), len(tools.Schemas()), len(parity.SlashCommands))
	if len(env.ActiveSkills) > 0 {
		b.WriteString("\nactive: " + strings.Join(env.ActiveSkills, ", "))
	}
	return b.String(), env
}

func helpFor(name string) string {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	for _, cmd := range Registry() {
		if cmd.Name == name {
			return commandHelp(cmd)
		}
	}
	return "unsupported: unknown command " + name
}

func commandHelp(cmd Command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s - %s\nusage: %s\nside effects: %s", cmd.Name, cmd.Summary, cmd.Usage, cmd.SideEffects)
	if len(cmd.Examples) > 0 {
		b.WriteString("\nexamples:")
		for _, example := range cmd.Examples {
			b.WriteString("\n  " + example)
		}
	}
	if len(cmd.Errors) > 0 {
		b.WriteString("\nerrors:")
		for _, err := range cmd.Errors {
			b.WriteString("\n  " + err)
		}
	}
	return b.String()
}

func runSkillNew(env Environment, fields []string) (string, Environment) {
	if len(fields) != 2 {
		return "usage: /skill-new <url>", env
	}
	parsed, err := url.ParseRequestURI(fields[1])
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "unverified: invalid URL " + fields[1], env
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "unverified: /skill-new supports only http and https URLs", env
	}
	fetch := env.FetchSkill
	if fetch == nil {
		fetch = defaultFetchSkill
	}
	body, err := fetch(fields[1])
	if err != nil {
		return "blocked: " + err.Error(), env
	}
	draft, err := skills.NewDraft(fields[1], body)
	if err != nil {
		return "unverified: " + err.Error(), env
	}
	cacheDir := filepath.Join(env.ProjectDir, ".rehamr", "fetched")
	if err := commandMkdirAll(cacheDir, 0o700); err != nil {
		return "blocked: " + err.Error(), env
	}
	cachePath := filepath.Join(cacheDir, draft.Name+".md")
	if err := commandWriteFile(cachePath, []byte(body), 0o600); err != nil {
		return "blocked: " + err.Error(), env
	}
	return fmt.Sprintf("skill draft: %s\nclass: %s\nfetched: %s\ntarget: %s\nnext: confirm content, then write approved skill to %s", draft.Name, draft.Classification.Class, filepath.ToSlash(cachePath), draft.TargetPath, draft.TargetPath), env
}

func defaultFetchSkill(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("skill fetch returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func runMCP(env Environment, fields []string) string {
	servers := mcp.Builtins()
	if len(fields) == 1 {
		if env.MCP != nil {
			return env.MCP.FormatStatus()
		}
		names := make([]string, 0, len(servers))
		for _, server := range servers {
			names = append(names, server.Name)
		}
		sort.Strings(names)
		return fmt.Sprintf("mcp servers: %d registered\n%s\nusage: /mcp connect|disconnect|tools|enable|disable <server> [tool]", len(servers), strings.Join(names, ", "))
	}
	switch fields[1] {
	case "connect":
		if len(fields) != 3 {
			return "usage: /mcp connect <server>"
		}
		if env.MCP == nil {
			return "unsupported: /mcp connect requires MCP manager wiring"
		}
		if err := env.MCP.Connect(context.Background(), fields[2]); err != nil {
			return "blocked: " + err.Error()
		}
		return "mcp connected: " + fields[2]
	case "disconnect":
		if len(fields) != 3 {
			return "usage: /mcp disconnect <server>"
		}
		if env.MCP == nil {
			return "unsupported: /mcp disconnect requires MCP manager wiring"
		}
		if err := env.MCP.Disconnect(fields[2]); err != nil {
			return "blocked: " + err.Error()
		}
		return "mcp disconnected: " + fields[2]
	case "tools":
		if len(fields) != 3 {
			return "usage: /mcp tools <server>"
		}
		if env.MCP != nil {
			return env.MCP.FormatTools(fields[2])
		}
		return mcpToolsStatus(fields[2], servers, env.ActiveSkills)
	case "enable", "disable":
		if len(fields) != 4 {
			return "usage: /mcp " + fields[1] + " <server> <tool | *>"
		}
		if env.MCP == nil {
			return "unsupported: /mcp " + fields[1] + " requires MCP manager wiring"
		}
		if err := env.MCP.SetToolEnabled(fields[2], fields[3], fields[1] == "enable"); err != nil {
			return "blocked: " + err.Error()
		}
		return "mcp " + fields[1] + "d: " + fields[2] + "." + fields[3]
	default:
		return "usage: /mcp [connect|disconnect|tools|enable|disable] <server> [tool]"
	}
}

func mcpToolsStatus(name string, servers []mcp.ServerConfig, activeSkills []string) string {
	for _, server := range servers {
		if server.Name != name {
			continue
		}
		visible := mcp.VisibleTools(server, activeSkills, server.AllowedTools)
		if len(visible) == 0 {
			return "mcp " + name + ": tools gated by active skill or not configured"
		}
		sort.Strings(visible)
		return "mcp " + name + " tools: " + strings.Join(visible, ", ")
	}
	return "unverified: unknown MCP server " + name
}

var commandCatalog = []Command{
	{"/clear", "reset the visible conversation", "/clear", "clears transient conversation state", []string{"/clear"}, []string{"none"}},
	{"/models", "list or switch model profiles", "/models [name]", "lists config profiles or updates `.rehamr/config.yaml` active profile", []string{"/models", "/models lmstudio-amd"}, []string{"blocked when config is not loaded", "unverified when the requested profile is missing"}},
	{"/skills", "list embedded and custom skills", "/skills", "reads embedded skills and optional custom skill directory", []string{"/skills"}, []string{"blocked when custom skill directory cannot be read"}},
	{"/skill", "load a skill by name", "/skill <name>", "updates active skill state for the current session", []string{"/skill ghidra-mcp"}, []string{"usage when name is missing", "unverified when the skill cannot be resolved"}},
	{"/skill-audit", "classify a skill template", "/skill-audit <name>", "none", []string{"/skill-audit n64-debug-mcp"}, []string{"usage when name is missing"}},
	{"/skill-new", "fetch and classify a new skill", "/skill-new <url>", "fetches HTTP(S) skill markdown, caches `.rehamr/fetched/<name>.md`, and reports the manual approval target under `.rehamr/skills/`", []string{"/skill-new https://example.com/SKILL.md"}, []string{"usage when URL is missing", "unverified when URL is invalid or content is too short", "blocked when fetch or cache write fails"}},
	{"/init-re", "create `.rehamr/` workspace", "/init-re", "creates config, project memory, MCP config, evidence, function, format, recomp, decomp, and skill files", []string{"/init-re"}, []string{"blocked when workspace creation fails"}},
	{"/status-re", "summarize RE workspace", "/status-re", "reads `.rehamr/` state and reports missing tracked files", []string{"/status-re"}, []string{"unsupported when workspace is not initialized"}},
	{"/doctor", "run local diagnostics", "/doctor", "reads local runtime, workspace, config, memory, skill, tool, MCP registration, and operational install/update/release file state without network probes", []string{"/doctor"}, []string{"unsupported when workspace, config, or memory is not initialized", "blocked when local config, workspace, custom skill, or operational release file state cannot be read"}},
	{"/mcp", "show or manage MCP servers", "/mcp [connect|disconnect|tools|enable|disable] <server> [tool]", "lists built-in MCP registrations; uses MCP manager wiring for connect, disconnect, tool listing, and tool mutation", []string{"/mcp", "/mcp connect ghidra", "/mcp tools ghidra"}, []string{"usage for malformed subcommands", "unsupported when MCP manager wiring is absent", "blocked for unknown server or lifecycle failure"}},
	{"/help", "show command help", "/help [command]", "none", []string{"/help", "/help /models"}, []string{"unsupported for unknown command"}},
}

package commands

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"recomphamr2/internal/config"
	"recomphamr2/internal/mcp"
)

func TestRegistryAndHelp(t *testing.T) {
	registry := Registry()
	if got := len(registry); got != 11 {
		t.Fatalf("len(Registry()) = %d, want 11", got)
	}
	registry[0].Examples[0] = "mutated"
	if Registry()[0].Examples[0] == "mutated" {
		t.Fatal("Registry() leaked mutable examples")
	}
	help, markdown := Help(), Markdown()
	for _, want := range []string{"/help", "/models", "/mcp", "side effects", "usage:"} {
		if !strings.Contains(help, want) {
			t.Fatalf("Help() missing %s: %s", want, help)
		}
	}
	for _, want := range []string{"# Slash Command Reference", "Generated from", "## `/clear`", "- Examples:", "- Errors:"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("Markdown() missing %s: %s", want, markdown)
		}
	}
	if out, _ := Execute(Environment{}, "/help models"); !strings.Contains(out, "/models -") || !strings.Contains(out, "examples:") {
		t.Fatalf("/help models output = %q", out)
	}
}

func TestPickerRows(t *testing.T) {
	if rows := ModelPickerRows(Environment{}); len(rows) != 1 || !rows[0].Blocked || !strings.Contains(rows[0].Summary, "config") {
		t.Fatalf("blocked model rows = %+v", rows)
	}
	cfg := config.Default()
	profile := cfg.Models[cfg.Active]
	profile.URL = "http://localhost:1234/v1?api_key=secret"
	cfg.Models[cfg.Active] = profile
	cfg.Models["empty-url"] = config.Profile{LLM: "empty", URL: "", ContextSize: 1024}
	cfg.Models["long-url"] = config.Profile{LLM: "long", URL: "http://localhost:1234/this/path/is/long/enough/to/truncate/value", ContextSize: 2048}
	modelRows := ModelPickerRows(Environment{Config: cfg})
	if len(modelRows) < 3 || !strings.Contains(modelRows[0].Detail, "url unverified") {
		t.Fatalf("model rows = %+v", modelRows)
	}
	var activeFound, queryRedacted, truncated bool
	for _, row := range modelRows {
		activeFound = activeFound || row.Active
		queryRedacted = queryRedacted || (row.Name == cfg.Active && !strings.Contains(row.Detail, "api_key"))
		truncated = truncated || strings.Contains(row.Detail, "...")
	}
	if !activeFound || !queryRedacted || !truncated {
		t.Fatalf("model row evidence active=%v queryRedacted=%v truncated=%v rows=%+v", activeFound, queryRedacted, truncated, modelRows)
	}

	if rows := SkillPickerRows(Environment{CustomSkillsDir: string(rune(0))}); len(rows) == 0 || !rows[len(rows)-1].Blocked {
		t.Fatalf("skill rows missing blocked custom row: %+v", rows)
	}
	customDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(customDir, "custom.md"), []byte("# Custom\nUse this skill when custom work is needed.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	skillRows := SkillPickerRows(Environment{CustomSkillsDir: customDir, ActiveSkills: []string{"custom"}})
	var customActive bool
	for _, row := range skillRows {
		customActive = customActive || row.Name == "custom" && row.Active && strings.Contains(row.Detail, "custom")
	}
	if !customActive {
		t.Fatalf("custom active skill row missing: %+v", skillRows)
	}

	t.Setenv("RECOMPHAMR_MCP_GHIDRA_URL", "http://localhost:1234/mcp?token=secret")
	t.Setenv("RECOMPHAMR_MCP_AUTOSTART", "1")
	builtinMCP := MCPPickerRows(Environment{})
	if len(builtinMCP) == 0 || !strings.Contains(builtinMCP[0].Summary, "disconnected") {
		t.Fatalf("builtin MCP rows = %+v", builtinMCP)
	}
	var ghidraHTTP bool
	for _, row := range builtinMCP {
		ghidraHTTP = ghidraHTTP || row.Name == "ghidra" && strings.Contains(row.Detail, "http") && !strings.Contains(row.Detail, "token")
	}
	if !ghidraHTTP {
		t.Fatalf("ghidra HTTP MCP row missing: %+v", builtinMCP)
	}

	manager := mcp.NewManager([]mcp.ServerConfig{{Name: "ok"}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return &commandFakeMCPClient{tools: []mcp.ToolDef{{Name: "decompile"}}}, nil
	}))
	if err := manager.Connect(context.Background(), "ok"); err != nil {
		t.Fatal(err)
	}
	managerRows := MCPPickerRows(Environment{MCP: manager})
	if len(managerRows) != 1 || !managerRows[0].Active || !strings.Contains(managerRows[0].Detail, "tools 1") {
		t.Fatalf("manager MCP rows = %+v", managerRows)
	}
	errManager := mcp.NewManager([]mcp.ServerConfig{{Name: "bad"}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return nil, errors.New("boom")
	}))
	_ = errManager.Connect(context.Background(), "bad")
	errorRows := MCPPickerRows(Environment{MCP: errManager})
	if len(errorRows) != 1 || !errorRows[0].Blocked || !strings.Contains(errorRows[0].Detail, "boom") {
		t.Fatalf("error MCP rows = %+v", errorRows)
	}
}

func TestExecuteCommands(t *testing.T) {
	cfg, _, err := config.Bootstrap(t.TempDir())
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	skillBody := "# Rules\nUse this skill when testing the skill-new workflow. Stop condition: done with evidence.\n"
	env := Environment{
		ProjectDir: t.TempDir(),
		Config:     cfg,
		FetchSkill: func(string) (string, error) { return skillBody, nil },
	}
	cases := []string{"/help", "/help /mcp", "/clear", "/models", "/skills", "/skill-audit n64-debug-mcp", "/init-re", "/status-re", "/doctor", "/mcp", "/mcp tools ghidra", "/skill-new https://example.com/skill.md"}
	for _, text := range cases {
		out, next := Execute(env, text)
		env = next
		if out == "" {
			t.Fatalf("Execute(%q) returned empty output", text)
		}
		if text == "/doctor" && !strings.Contains(out, "RecompHamr doctor") {
			t.Fatalf("/doctor output = %q", out)
		}
	}
	out, next := Execute(env, "/skill ghidra-mcp")
	if !strings.Contains(out, "loaded skill") || len(next.ActiveSkills) == 0 {
		t.Fatalf("/skill output=%q env=%+v", out, next)
	}
	out, _ = Execute(env, "/models ollama-amd")
	if !strings.Contains(out, "active model") {
		t.Fatalf("/models switch output=%q", out)
	}
	out, _ = Execute(Environment{ActiveSkills: []string{"ghidra"}}, "/mcp tools ghidra")
	if !strings.Contains(out, "gated") {
		t.Fatalf("empty MCP tools output=%q", out)
	}
}

func TestSkillNewWorkflow(t *testing.T) {
	body := "# Rules\nUse this skill when importing reviewed markdown. Stop condition: done with evidence.\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	defaultOut, _ := Execute(Environment{ProjectDir: t.TempDir()}, "/skill-new "+server.URL+"/Skill.md")
	if !strings.Contains(defaultOut, "skill draft: skill") {
		t.Fatalf("default /skill-new output = %q", defaultOut)
	}

	projectDir := t.TempDir()
	out, _ := Execute(Environment{
		ProjectDir: projectDir,
		FetchSkill: func(rawURL string) (string, error) {
			if rawURL != "https://example.com/Skill.md" {
				t.Fatalf("FetchSkill() URL = %q", rawURL)
			}
			return body, nil
		},
	}, "/skill-new https://example.com/Skill.md")
	for _, want := range []string{"skill draft: skill", "class: micro_skill", "fetched:", "target: .rehamr/skills/skill.md"} {
		if !strings.Contains(out, want) {
			t.Fatalf("/skill-new output missing %q: %q", want, out)
		}
	}
	cached, err := os.ReadFile(filepath.Join(projectDir, ".rehamr", "fetched", "skill.md"))
	if err != nil || string(cached) != body {
		t.Fatalf("cached skill = %q, %v", string(cached), err)
	}

	for _, tc := range []struct {
		name string
		env  Environment
		text string
		want string
	}{
		{"scheme", Environment{}, "/skill-new ftp://example.com/skill.md", "unverified:"},
		{"fetch", Environment{FetchSkill: func(string) (string, error) { return "", errors.New("fetch failed") }}, "/skill-new https://example.com/skill.md", "blocked:"},
		{"short", Environment{FetchSkill: func(string) (string, error) { return "short", nil }}, "/skill-new https://example.com/skill.md", "unverified:"},
		{"cache", Environment{ProjectDir: filepath.Join(t.TempDir(), "file"), FetchSkill: func(string) (string, error) { return body, nil }}, "/skill-new https://example.com/skill.md", "blocked:"},
	} {
		if tc.name == "cache" {
			if err := os.WriteFile(tc.env.ProjectDir, []byte("x"), 0o600); err != nil {
				t.Fatalf("WriteFile() cache blocker error = %v", err)
			}
		}
		out, _ := Execute(tc.env, tc.text)
		if !strings.Contains(out, tc.want) {
			t.Fatalf("%s Execute() = %q, want %q", tc.name, out, tc.want)
		}
	}

	origWriteFile := commandWriteFile
	defer func() { commandWriteFile = origWriteFile }()
	commandWriteFile = func(string, []byte, os.FileMode) error { return errors.New("write failed") }
	out, _ = Execute(Environment{ProjectDir: t.TempDir(), FetchSkill: func(string) (string, error) { return body, nil }}, "/skill-new https://example.com/skill.md")
	if !strings.Contains(out, "blocked:") {
		t.Fatalf("write failure /skill-new output = %q", out)
	}
}

func TestDefaultFetchSkill(t *testing.T) {
	body := "# Rules\nUse this skill when fetching over HTTP. Stop condition: done with evidence.\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail.md" {
			http.Error(w, "nope", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	got, err := defaultFetchSkill(server.URL + "/skill.md")
	if err != nil || got != body {
		t.Fatalf("defaultFetchSkill() = %q, %v", got, err)
	}
	if _, err := defaultFetchSkill(server.URL + "/fail.md"); err == nil {
		t.Fatal("defaultFetchSkill() accepted HTTP failure")
	}
	if _, err := defaultFetchSkill("http://%"); err == nil {
		t.Fatal("defaultFetchSkill() accepted invalid URL")
	}
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: errorReadCloser{}}, nil
	})
	if _, err := defaultFetchSkill("https://example.com/skill.md"); err == nil {
		t.Fatal("defaultFetchSkill() accepted read failure")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errorReadCloser) Close() error {
	return nil
}

func TestExecuteErrors(t *testing.T) {
	env := Environment{ProjectDir: t.TempDir()}
	badProject := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(badProject, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	for _, text := range []string{"", "models", "/help x y", "/help missing", "/models", "/models a b", "/skill", "/skill missing", "/skill-audit", "/skill-new", "/skill-new ://bad", "/unknown"} {
		out, _ := Execute(env, text)
		if !strings.Contains(out, "unsupported") && !strings.Contains(out, "usage:") && !strings.Contains(out, "blocked:") && !strings.Contains(out, "unverified:") {
			t.Fatalf("Execute(%q) = %q, want explicit non-success", text, out)
		}
	}
	cfg, _, err := config.Bootstrap(t.TempDir())
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	out, _ := Execute(Environment{Config: cfg}, "/models missing")
	if !strings.Contains(out, "unverified:") {
		t.Fatalf("unknown model output = %q", out)
	}
	out, _ = Execute(Environment{Config: cfg}, "/models a b")
	if !strings.Contains(out, "usage: /models [name]") {
		t.Fatalf("too many models args output = %q", out)
	}
	out, _ = Execute(Environment{ProjectDir: badProject}, "/init-re")
	if !strings.Contains(out, "blocked:") {
		t.Fatalf("bad init output = %q", out)
	}
}

func TestMCPSubcommandValidation(t *testing.T) {
	cases := map[string]string{
		"/mcp connect":             "usage: /mcp connect <server>",
		"/mcp connect ghidra":      "unsupported:",
		"/mcp disconnect":          "usage: /mcp disconnect <server>",
		"/mcp disconnect ghidra":   "unsupported:",
		"/mcp tools":               "usage: /mcp tools <server>",
		"/mcp tools missing":       "unverified:",
		"/mcp enable":              "usage: /mcp enable <server> <tool | *>",
		"/mcp enable ghidra *":     "unsupported:",
		"/mcp disable":             "usage: /mcp disable <server> <tool | *>",
		"/mcp disable ghidra tool": "unsupported:",
		"/mcp what":                "usage: /mcp [connect|disconnect|tools|enable|disable]",
	}
	for text, want := range cases {
		out, _ := Execute(Environment{}, text)
		if !strings.Contains(out, want) {
			t.Fatalf("Execute(%q) = %q, want %q", text, out, want)
		}
	}
}

func TestMCPManagerCommands(t *testing.T) {
	manager := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra", RequireSkill: true}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return &commandFakeMCPClient{tools: []mcp.ToolDef{{Name: "decompile"}, {Name: "xref"}}}, nil
	}))
	env := Environment{MCP: manager, ActiveSkills: []string{"ghidra-mcp"}}
	out, env := Execute(env, "/mcp")
	if !strings.Contains(out, "ghidra disconnected") {
		t.Fatalf("/mcp status = %q", out)
	}
	out, env = Execute(env, "/mcp connect ghidra")
	if !strings.Contains(out, "mcp connected: ghidra") {
		t.Fatalf("/mcp connect = %q", out)
	}
	out, env = Execute(env, "/mcp tools ghidra")
	if !strings.Contains(out, "decompile") || !strings.Contains(out, "xref") {
		t.Fatalf("/mcp tools = %q", out)
	}
	out, env = Execute(env, "/mcp disable ghidra xref")
	if !strings.Contains(out, "mcp disabled: ghidra.xref") {
		t.Fatalf("/mcp disable = %q", out)
	}
	out, env = Execute(env, "/mcp tools ghidra")
	if strings.Contains(out, "xref") {
		t.Fatalf("/mcp tools after disable = %q", out)
	}
	out, env = Execute(env, "/mcp enable ghidra xref")
	if !strings.Contains(out, "mcp enabled: ghidra.xref") {
		t.Fatalf("/mcp enable = %q", out)
	}
	out, env = Execute(env, "/mcp disconnect ghidra")
	if !strings.Contains(out, "mcp disconnected: ghidra") {
		t.Fatalf("/mcp disconnect = %q", out)
	}
	out, _ = Execute(Environment{MCP: mcp.NewManager(nil, nil)}, "/mcp connect missing")
	if !strings.Contains(out, "blocked:") {
		t.Fatalf("/mcp unknown connect = %q", out)
	}
	out, _ = Execute(Environment{MCP: mcp.NewManager(nil, nil)}, "/mcp disconnect missing")
	if !strings.Contains(out, "blocked:") {
		t.Fatalf("/mcp unknown disconnect = %q", out)
	}
	out, _ = Execute(Environment{MCP: mcp.NewManager(nil, nil)}, "/mcp enable missing tool")
	if !strings.Contains(out, "blocked:") {
		t.Fatalf("/mcp unknown enable = %q", out)
	}
}

type commandFakeMCPClient struct {
	tools []mcp.ToolDef
}

func (c *commandFakeMCPClient) Tools() []mcp.ToolDef {
	return append([]mcp.ToolDef(nil), c.tools...)
}

func (c *commandFakeMCPClient) CallTool(context.Context, string, map[string]interface{}) (mcp.CallToolResult, error) {
	return mcp.CallToolResult{}, nil
}

func (c *commandFakeMCPClient) Close() error {
	return nil
}

func TestSkillsCustomDirectoryFailure(t *testing.T) {
	out, _ := Execute(Environment{CustomSkillsDir: string(rune(0))}, "/skills")
	if !strings.Contains(out, "blocked:") {
		t.Fatalf("/skills custom failure output = %q", out)
	}
}

func TestMCPToolsVisibleWhenEnvAllowsAndSkillActive(t *testing.T) {
	t.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", "decompile,function")
	out, _ := Execute(Environment{ActiveSkills: []string{"ghidra"}}, "/mcp tools ghidra")
	if !strings.Contains(out, "decompile") || !strings.Contains(out, "function") {
		t.Fatalf("visible MCP tools output = %q", out)
	}
}

func TestModelsAndSkillsRichSuccessOutput(t *testing.T) {
	cfg, _, err := config.Bootstrap(t.TempDir())
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	customDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(customDir, "custom.md"), []byte("# Custom"), 0o600); err != nil {
		t.Fatalf("WriteFile() custom skill error = %v", err)
	}
	env := Environment{Config: cfg, CustomSkillsDir: customDir, ActiveSkills: []string{"custom"}}
	out, _ := Execute(env, "/models")
	if !strings.Contains(out, "* "+cfg.Active) {
		t.Fatalf("/models output = %q, want active marker", out)
	}
	out, _ = Execute(env, "/skills")
	for _, want := range []string{"1 custom", "1 active", "active: custom"} {
		if !strings.Contains(out, want) {
			t.Fatalf("/skills output missing %q: %q", want, out)
		}
	}
}

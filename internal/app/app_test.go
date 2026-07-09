package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"recomphamr2/internal/config"
	"recomphamr2/internal/llm"
	"recomphamr2/internal/project"
	"recomphamr2/internal/tui"
)

func TestRunDiagnostic(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{DiagnosticMode}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "phase-0: complete") || !strings.Contains(stdout.String(), "product wiring available") {
		t.Fatalf("diagnostic output = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("diagnostic stderr = %q, want empty", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{HelpMode}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "usage: recomphamr [--diagnostic|--help]") || !strings.Contains(stdout.String(), "no arguments") {
		t.Fatalf("help output = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("help stderr = %q, want empty", stderr.String())
	}
}

func TestRunComposesRuntime(t *testing.T) {
	cwd := chdirTemp(t)
	defer cwd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() exit code = %d stderr=%q", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"RecompHamr product runtime",
		"config: created active=lmstudio-amd profiles=4",
		"memory: unsupported: REPHAMR_STATE.md is missing",
		"mcp: manager wired servers=8 autoconnect=false",
		"agent: wired for later interactive turns; no model call made during startup",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("runtime output missing %q:\n%s", want, out)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("runtime stderr = %q, want empty", stderr.String())
	}
}

func TestRunUnsupportedArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--bad"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run() exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("unsupported stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "usage: recomphamr") {
		t.Fatalf("unsupported stderr = %q", stderr.String())
	}
}

func TestRunReportsRuntimeCompositionBlocker(t *testing.T) {
	orig := bootstrapConfig
	defer func() { bootstrapConfig = orig }()
	bootstrapConfig = func(string) (*config.Config, bool, error) {
		return nil, false, errors.New("config blocked")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 2 || !strings.Contains(stderr.String(), "blocked: config blocked") {
		t.Fatalf("Run() code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestComposeRuntimeLoadsMemory(t *testing.T) {
	dir := t.TempDir()
	ws, err := project.Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.WriteFile(ws.State, []byte(strings.Repeat("abcd", 20)), 0o600); err != nil {
		t.Fatalf("WriteFile() memory error = %v", err)
	}

	runtime, err := ComposeRuntime(Options{ProjectDir: dir, MemoryMaxBytes: 16})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	if runtime.ConfigCreated || runtime.Commands.Config == nil || runtime.Commands.MCP == nil || runtime.MCP == nil {
		t.Fatalf("ComposeRuntime() incomplete runtime = %#v", runtime)
	}
	if !strings.Contains(runtime.MemoryStatus, "verified bytes=16 truncated=true") {
		t.Fatalf("MemoryStatus = %q", runtime.MemoryStatus)
	}
	if runtime.TUI.Layout.Mode != "ready" || runtime.TUI.Layout.ActiveModel != runtime.Config.Active {
		t.Fatalf("TUI layout = %#v", runtime.TUI.Layout)
	}
	if got := runtime.TUI.Submit("/models").Transcript[1]; !strings.Contains(got, "* lmstudio-amd") {
		t.Fatalf("runtime command dispatch = %q", got)
	}
}

func TestComposeRuntimeCustomSkillsDefault(t *testing.T) {
	dir := t.TempDir()
	runtime, err := ComposeRuntime(Options{ProjectDir: dir})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	want := filepath.Join(dir, config.DirName, "skills")
	if runtime.Commands.CustomSkillsDir != want {
		t.Fatalf("CustomSkillsDir = %q, want %q", runtime.Commands.CustomSkillsDir, want)
	}
	if !runtime.ConfigCreated {
		t.Fatal("ComposeRuntime() did not report config creation")
	}
}

func TestComposeRuntimeMemoryReadBlocked(t *testing.T) {
	orig := loadMemory
	defer func() { loadMemory = orig }()
	loadMemory = func(string, int) (project.Memory, error) {
		return project.Memory{}, errors.New("memory blocked")
	}
	_, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "memory blocked") {
		t.Fatalf("ComposeRuntime() memory blocker error = %v", err)
	}
}

func TestComposeRuntimeActiveProfileBlocked(t *testing.T) {
	orig := bootstrapConfig
	defer func() { bootstrapConfig = orig }()
	bootstrapConfig = func(projectDir string) (*config.Config, bool, error) {
		cfg := config.Default()
		cfg.Active = "missing"
		cfg.Dir = filepath.Join(projectDir, config.DirName)
		return cfg, false, nil
	}
	_, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "active profile") {
		t.Fatalf("ComposeRuntime() active profile error = %v", err)
	}
}

func TestRuntimeSummaryHandlesNilConfig(t *testing.T) {
	summary := (Runtime{ProjectDir: "x", MemoryStatus: "unsupported", TUI: tui.Model{Layout: tui.Layout{Mode: "ready"}}}).Summary()
	if !strings.Contains(summary, "active=unverified profiles=0") {
		t.Fatalf("Summary() = %q", summary)
	}
}

func TestLoadOptionalMemoryMissing(t *testing.T) {
	mem, status, err := loadOptionalMemory(t.TempDir(), 0)
	if err != nil {
		t.Fatalf("loadOptionalMemory() error = %v", err)
	}
	if mem.MaxBytes == 0 || !strings.HasPrefix(status, "unsupported:") {
		t.Fatalf("loadOptionalMemory() mem=%#v status=%q", mem, status)
	}
}

func TestRunSmokeSlashCommand(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	report, err := RunSmoke(context.Background(), SmokeSpec{Runtime: runtime, Prompt: "/models", RenderWidth: tui.CompactWidth - 1})
	if err != nil {
		t.Fatalf("RunSmoke(/models) error = %v", err)
	}
	if !strings.Contains(report.CommandOutput, "* lmstudio-amd") || !strings.Contains(report.Render, "models:") {
		t.Fatalf("slash smoke report = %#v", report)
	}
}

func TestRunSmokePromptToolLoopAndMemory(t *testing.T) {
	dir := t.TempDir()
	ws, err := project.Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.WriteFile(ws.State, []byte("memory fact: use fake evidence"), 0o600); err != nil {
		t.Fatalf("WriteFile() memory error = %v", err)
	}
	runtime, err := ComposeRuntime(Options{ProjectDir: dir})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	model := &scriptModel{t: t, replies: []llm.Message{
		{Role: "assistant", Tools: []llm.ToolCall{{ID: "call-1", Name: "read_file", Arguments: map[string]any{"path": "README.md"}}}},
		{Role: "assistant", Content: "verified: fake runtime smoke complete"},
	}}
	report, err := RunSmoke(context.Background(), SmokeSpec{
		Runtime:     runtime,
		Prompt:      "inspect README",
		Model:       model,
		RunTool:     func(context.Context, llm.ToolCall) (string, error) { return "fake tool result", nil },
		RenderWidth: tui.CompactWidth - 1,
	})
	if err != nil {
		t.Fatalf("RunSmoke(prompt) error = %v", err)
	}
	if !model.sawMemory {
		t.Fatal("fake model did not receive project memory")
	}
	for _, want := range []string{"assistant: verified: fake runtime smoke complete", "tool: fake tool result", "RecompHamr initiative"} {
		if !strings.Contains(report.Render, want) {
			t.Fatalf("prompt smoke render missing %q:\n%s", want, report.Render)
		}
	}
	if len(report.Messages) < 4 || report.Cancelled {
		t.Fatalf("prompt smoke report = %#v", report)
	}
}

func TestRunSmokeCancellation(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	report, err := RunSmoke(context.Background(), SmokeSpec{
		Runtime:         runtime,
		Prompt:          "cancel me",
		Model:           &scriptModel{t: t, replies: []llm.Message{{Role: "assistant", Content: "should not run"}}},
		RunTool:         func(context.Context, llm.ToolCall) (string, error) { return "", nil },
		CancelBeforeRun: true,
	})
	if !errors.Is(err, context.Canceled) || !report.Cancelled {
		t.Fatalf("RunSmoke(cancel) report=%#v error=%v", report, err)
	}
}

func TestRunSmokeErrors(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	if _, err := RunSmoke(nil, SmokeSpec{Runtime: runtime}); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("RunSmoke(empty) error = %v", err)
	}
	if _, err := RunSmoke(context.Background(), SmokeSpec{Runtime: runtime, Prompt: "hello"}); err == nil || !strings.Contains(err.Error(), "model and tool runner") {
		t.Fatalf("RunSmoke(missing fake deps) error = %v", err)
	}
}

func TestSmokeLinesEmptyMessage(t *testing.T) {
	lines := smokeLines([]llm.Message{{Role: "assistant"}})
	if len(lines) != 1 || lines[0] != "assistant: (empty)" {
		t.Fatalf("smokeLines() = %v", lines)
	}
}

func chdirTemp(t *testing.T) func() {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	return func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore Chdir() error = %v", err)
		}
	}
}

type scriptModel struct {
	t         *testing.T
	replies   []llm.Message
	calls     int
	sawMemory bool
}

func (m *scriptModel) Next(_ context.Context, messages []llm.Message) (llm.Message, error) {
	m.t.Helper()
	if len(messages) > 0 && strings.Contains(messages[0].Content, "memory fact") {
		m.sawMemory = true
	}
	if m.calls >= len(m.replies) {
		m.t.Fatalf("scriptModel exhausted after %d calls", m.calls)
	}
	reply := m.replies[m.calls]
	m.calls++
	return reply, nil
}

func BenchmarkComposeRuntimeStartup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runtime, err := ComposeRuntime(Options{ProjectDir: b.TempDir()})
		if err != nil {
			b.Fatalf("ComposeRuntime() error = %v", err)
		}
		if runtime.Config == nil || runtime.MCP == nil {
			b.Fatal("ComposeRuntime returned incomplete runtime")
		}
	}
}

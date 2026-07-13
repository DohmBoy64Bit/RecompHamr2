package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"recomphamr2/internal/agent"
	"recomphamr2/internal/config"
	"recomphamr2/internal/llm"
	"recomphamr2/internal/mcp"
	"recomphamr2/internal/project"
	"recomphamr2/internal/tools"
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
	if !strings.Contains(stdout.String(), "usage: recomphamr [--diagnostic|--help|--summary]") || !strings.Contains(stdout.String(), "launch the interactive terminal app") {
		t.Fatalf("help output = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("help stderr = %q, want empty", stderr.String())
	}
}

func TestRunLaunchesInteractiveRuntime(t *testing.T) {
	cwd := chdirTemp(t)
	defer cwd()
	origRun := runProgram
	origModel := newAgentModel
	origTools := newToolRunner
	defer func() {
		runProgram = origRun
		newAgentModel = origModel
		newToolRunner = origTools
	}()
	launched := false
	runProgram = func(model tea.Model, stdout io.Writer, stderr io.Writer) error {
		launched = true
		if _, ok := model.(liveModel); !ok {
			t.Fatalf("launch model = %T, want liveModel", model)
		}
		fmt.Fprintln(stdout, "launched")
		return nil
	}
	newAgentModel = func(Runtime) (agent.Model, error) {
		return &scriptModel{t: t, replies: []llm.Message{{Role: "assistant", Content: "ok"}}}, nil
	}
	newToolRunner = func(Runtime) agent.ToolRunner {
		return func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() exit code = %d stderr=%q", code, stderr.String())
	}
	if !launched || !strings.Contains(stdout.String(), "launched") {
		t.Fatalf("interactive launch stdout=%q launched=%v", stdout.String(), launched)
	}
	if stderr.Len() != 0 {
		t.Fatalf("runtime stderr = %q, want empty", stderr.String())
	}
}

func TestRunSummaryComposesRuntime(t *testing.T) {
	cwd := chdirTemp(t)
	defer cwd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{SummaryMode}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() exit code = %d stderr=%q", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"RecompHamr product runtime",
		"config: created active=lmstudio-amd profiles=4",
		"memory: unsupported: REPHAMR_STATE.md is missing",
		"mcp: manager wired servers=8 autoconnect=false",
		"agent: wired for interactive turns; no model call made during summary",
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

func TestRunUnsupportedMultipleArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help", "--bad"}, &stdout, &stderr)

	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "--summary") {
		t.Fatalf("Run(multiple) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
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

func TestRunSummaryReportsRuntimeCompositionBlocker(t *testing.T) {
	orig := bootstrapConfig
	defer func() { bootstrapConfig = orig }()
	bootstrapConfig = func(string) (*config.Config, bool, error) {
		return nil, false, errors.New("summary config blocked")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{SummaryMode}, &stdout, &stderr)

	if code != 2 || !strings.Contains(stderr.String(), "blocked: summary config blocked") {
		t.Fatalf("Run(summary) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunReportsModelConstructionBlocker(t *testing.T) {
	cwd := chdirTemp(t)
	defer cwd()
	origModel := newAgentModel
	defer func() { newAgentModel = origModel }()
	newAgentModel = func(Runtime) (agent.Model, error) { return nil, errors.New("model blocked") }
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 2 || !strings.Contains(stderr.String(), "blocked: model blocked") {
		t.Fatalf("Run(model blocker) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunReportsLaunchBlocker(t *testing.T) {
	origRun := runProgram
	defer func() { runProgram = origRun }()
	runProgram = func(tea.Model, io.Writer, io.Writer) error { return errors.New("launch blocked") }
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 2 || !strings.Contains(stderr.String(), "blocked: launch blocked") {
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
	if len(runtime.TUI.Transcript) != 0 {
		t.Fatalf("startup transcript = %#v, want empty launcher state", runtime.TUI.Transcript)
	}
	if got := runtime.TUI.Submit("/models").Transcript[0]; !strings.Contains(got, "* lmstudio-amd") {
		t.Fatalf("runtime command dispatch = %q", got)
	}
}

func TestComposeRuntimeStartsOnLauncher(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	view := runtime.TUI.RenderStyledWithLayout(tui.Layout{
		Width:         120,
		Height:        32,
		Mode:          runtime.TUI.Layout.Mode,
		ActiveModel:   runtime.TUI.Layout.ActiveModel,
		ActiveSkill:   runtime.TUI.Layout.ActiveSkill,
		MCPStatus:     runtime.TUI.Layout.MCPStatus,
		ContextStatus: runtime.TUI.Layout.ContextStatus,
		PendingTool:   runtime.TUI.Layout.PendingTool,
		MemoryStatus:  runtime.TUI.Layout.MemoryStatus,
	})
	plainView := ansi.Strip(view)
	for _, want := range []string{"RECOMP HAMR", "Ask RecompHamr", "ready  lmstudio-amd  ready"} {
		if !strings.Contains(plainView, want) {
			t.Fatalf("startup launcher missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "runtime: local product shell composed") || strings.Contains(view, "note        runtime:") {
		t.Fatalf("startup launcher leaked runtime transcript note:\n%s", view)
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

func TestComposeRuntimeLoadsPersistentMCPConfig(t *testing.T) {
	dir := t.TempDir()
	mcpDir := filepath.Join(dir, config.DirName)
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	body := `{"servers":{"custom":{"url":"http://example.invalid","allowed_tools":["ping"],"require_skill":false}}}`
	if err := os.WriteFile(filepath.Join(mcpDir, "mcp.json"), []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile(mcp.json) error = %v", err)
	}
	runtime, err := ComposeRuntime(Options{ProjectDir: dir})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	if status := runtime.MCP.Status("custom"); status.Name != "custom" || status.State != mcp.StateDisconnected {
		t.Fatalf("custom MCP status = %#v", status)
	}
	if !strings.Contains(runtime.Summary(), "servers=9") {
		t.Fatalf("Summary() missing persistent server count:\n%s", runtime.Summary())
	}
}

func TestComposeRuntimeMCPConfigBlocked(t *testing.T) {
	dir := t.TempDir()
	mcpDir := filepath.Join(dir, config.DirName)
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, "mcp.json"), []byte(`{"unknown":true}`), 0o600); err != nil {
		t.Fatalf("WriteFile(mcp.json) error = %v", err)
	}
	_, err := ComposeRuntime(Options{ProjectDir: dir})
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("ComposeRuntime() mcp config error = %v", err)
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

func TestRuntimeSummaryReportsConnectedMCP(t *testing.T) {
	manager := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra", URL: "http://example.invalid"}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return fakeMCPClient{}, nil
	}))
	if err := manager.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	summary := (Runtime{ProjectDir: "x", MCP: manager, MemoryStatus: "unsupported", TUI: tui.Model{Layout: tui.Layout{Mode: "ready"}}}).Summary()
	if !strings.Contains(summary, "autoconnect=connected=1") {
		t.Fatalf("Summary() = %q", summary)
	}
	if got := autoconnectStatus(nil); got != "none" {
		t.Fatalf("autoconnectStatus(nil) = %q", got)
	}
}

func TestRuntimeMCPConfigsLoadError(t *testing.T) {
	orig := loadMCPConfig
	defer func() { loadMCPConfig = orig }()
	loadMCPConfig = func(string) (mcp.ConfigFile, error) {
		return mcp.ConfigFile{}, errors.New("mcp config blocked")
	}
	if _, err := runtimeMCPConfigs(t.TempDir()); err == nil || !strings.Contains(err.Error(), "mcp config blocked") {
		t.Fatalf("runtimeMCPConfigs() error = %v", err)
	}
}

func TestLaunchBuildsLiveProgram(t *testing.T) {
	origRun := runProgram
	defer func() { runProgram = origRun }()
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	called := false
	runProgram = func(model tea.Model, stdout io.Writer, stderr io.Writer) error {
		called = true
		live, ok := model.(liveModel)
		if !ok || live.model == nil || live.tools == nil {
			t.Fatalf("live model = %#v ok=%v", model, ok)
		}
		return nil
	}
	if err := Launch(runtime, io.Discard, io.Discard); err != nil {
		t.Fatalf("Launch() error = %v", err)
	}
	if !called {
		t.Fatal("Launch() did not run program")
	}
}

func TestLaunchRequiresRuntimeConfig(t *testing.T) {
	err := Launch(Runtime{}, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "runtime config") {
		t.Fatalf("Launch() error = %v", err)
	}
}

func TestLiveModelSlashPromptCancelAndQuit(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	model := liveModel{
		BubbleModel: tui.BubbleModel{State: runtime.TUI, LastAction: tui.ActionNone},
		runtime:     runtime,
		model:       &scriptModel{t: t, replies: []llm.Message{{Role: "assistant", Content: "assistant ready"}}},
		tools:       func(context.Context, llm.ToolCall) (string, error) { return "", nil },
	}
	updated, cmd := model.Update(keyText("/models"))
	model = updated.(liveModel)
	updated, cmd = model.Update(keyCode(tea.KeyEnter))
	model = updated.(liveModel)
	if cmd != nil || !strings.Contains(model.View().Content, "active model:") {
		t.Fatalf("slash update cmd=%v view=\n%s", cmd, model.View().Content)
	}
	for _, picker := range []string{"/skills", "/mcp"} {
		model.BubbleModel.State.Composer = picker
		updated, cmd = model.Update(keyCode(tea.KeyEnter))
		model = updated.(liveModel)
		if cmd != nil || model.BubbleModel.LastIntent.Value == "" {
			t.Fatalf("picker %s cmd=%v intent=%+v", picker, cmd, model.BubbleModel.LastIntent)
		}
	}
	model.BubbleModel.State.Composer = "/zzz"
	updated, cmd = model.Update(keyCode(tea.KeyEnter))
	model = updated.(liveModel)
	if cmd != nil || !strings.Contains(model.BubbleModel.State.Transcript[len(model.BubbleModel.State.Transcript)-1], "unknown command") {
		t.Fatalf("unknown slash command did not stay local: %+v", model.BubbleModel.State.Transcript)
	}
	updated, _ = model.Update(keyText("hello"))
	model = updated.(liveModel)
	updated, cmd = model.Update(keyCode(tea.KeyEnter))
	model = updated.(liveModel)
	if cmd == nil || model.BubbleModel.State.Layout.Mode != "thinking" {
		t.Fatalf("prompt update cmd=%v mode=%q", cmd, model.BubbleModel.State.Layout.Mode)
	}
	result := cmd().(agentResult)
	updated, _ = model.Update(result)
	model = updated.(liveModel)
	if !strings.Contains(model.View().Content, "assistant: assistant ready") {
		t.Fatalf("agent result view=\n%s", model.View().Content)
	}
	updated, _ = model.Update(keyCtrl('c'))
	model = updated.(liveModel)
	updated, cmd = model.Update(keyCtrl('c'))
	model = updated.(liveModel)
	if cmd == nil || model.BubbleModel.State.Status != "quit" {
		t.Fatalf("quit cmd=%v status=%q", cmd, model.BubbleModel.State.Status)
	}
}

func TestLiveModelExposesMCPToolsForActiveSkill(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	manager := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra", RequireSkill: true}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return fakeMCPClient{}, nil
	}))
	if err := manager.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	runtime.MCP = manager
	runtime.TUI.Env.MCP = manager
	model := &toolAwareScriptModel{scriptModel: scriptModel{t: t, replies: []llm.Message{{Role: "assistant", Tools: []llm.ToolCall{{ID: "mcp-1", Name: "ghidra.decompile", Arguments: map[string]any{"address": "0x1000"}}}}, {Role: "assistant", Content: "mcp done"}}}}
	live := liveModel{
		BubbleModel: tui.BubbleModel{State: runtime.TUI.Submit("/skill ghidra-mcp"), LastAction: tui.ActionNone},
		runtime:     runtime,
		model:       model,
		tools:       liveToolRunner(runtime),
	}
	updated, _ := live.Update(keyText("use mcp"))
	live = updated.(liveModel)
	updated, cmd := live.Update(keyCode(tea.KeyEnter))
	live = updated.(liveModel)
	result := cmd().(agentResult)
	if result.err != nil {
		t.Fatalf("mcp agent result error = %v messages=%#v", result.err, result.messages)
	}
	updated, _ = live.Update(result)
	live = updated.(liveModel)

	transcript := strings.Join(live.BubbleModel.State.Transcript, "\n")
	if !model.sawTool("ghidra.decompile") || !strings.Contains(transcript, "tool: mcp decompiled") || !strings.Contains(transcript, "assistant: mcp done") {
		t.Fatalf("mcp transcript=\n%s tools=%v messages=%#v", transcript, model.toolNames, result.messages)
	}
}

func TestLiveModelInitAndBlockedResult(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	model := liveModel{BubbleModel: tui.BubbleModel{State: runtime.TUI}, runtime: runtime}
	if cmd := model.Init(); cmd != nil {
		t.Fatalf("Init() cmd=%v, want nil", cmd)
	}
	updated, _ := model.Update(agentResult{err: errors.New("backend failed")})
	model = updated.(liveModel)
	if model.BubbleModel.State.Status != "blocked" || !strings.Contains(model.View().Content, "blocked: backend failed") {
		t.Fatalf("blocked result status=%q view=\n%s", model.BubbleModel.State.Status, model.View().Content)
	}
}

func TestLiveModelCancellationCancelsAgentContext(t *testing.T) {
	runtime, err := ComposeRuntime(Options{ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatalf("ComposeRuntime() error = %v", err)
	}
	block := make(chan struct{})
	model := liveModel{
		BubbleModel: tui.BubbleModel{State: runtime.TUI, LastAction: tui.ActionNone},
		runtime:     runtime,
		model: modelFunc(func(ctx context.Context, _ []llm.Message) (llm.Message, error) {
			<-block
			return llm.Message{}, ctx.Err()
		}),
		tools: func(context.Context, llm.ToolCall) (string, error) { return "", nil },
	}
	updated, _ := model.Update(keyText("cancel"))
	model = updated.(liveModel)
	updated, cmd := model.Update(keyCode(tea.KeyEnter))
	model = updated.(liveModel)
	updated, _ = model.Update(keyCtrl('c'))
	model = updated.(liveModel)
	close(block)
	result := cmd().(agentResult)
	updated, _ = model.Update(result)
	model = updated.(liveModel)
	if model.BubbleModel.State.Status != "cancelled" {
		t.Fatalf("cancel status = %q result=%v", model.BubbleModel.State.Status, result.err)
	}
}

func TestStreamModelNext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"hello"}}]}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"choices":[],"usage":{"completion_tokens":1,"prompt_tokens":2}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: [DONE]`)
	}))
	defer server.Close()
	model := streamModel{client: llm.NewClient(server.URL, "fake", "")}

	reply, err := model.Next(context.Background(), []llm.Message{{Role: "user", Content: "hi"}})

	if err != nil || reply.Content != "hello" {
		t.Fatalf("Next() reply=%#v error=%v", reply, err)
	}
}

func TestStreamModelWithToolsCopiesDefinitions(t *testing.T) {
	base := streamModel{}
	toolsIn := []llm.Tool{{Type: "function", Function: llm.FunctionDef{Name: "first"}}}
	withTools := base.WithTools(toolsIn).(streamModel)
	toolsIn[0].Function.Name = "mutated"
	if len(withTools.tools) != 1 || withTools.tools[0].Function.Name != "first" {
		t.Fatalf("WithTools() tools = %#v", withTools.tools)
	}
}

func TestStreamModelNextErrors(t *testing.T) {
	if _, err := (streamModel{}).Next(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "LLM client") {
		t.Fatalf("Next(nil client) error = %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusInternalServerError)
	}))
	defer server.Close()
	model := streamModel{client: llm.NewClient(server.URL, "fake", "")}
	if _, err := model.Next(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("Next(server error) = %v", err)
	}
}

func TestFinalFromStreamDefensiveErrors(t *testing.T) {
	empty := make(chan llm.Event)
	close(empty)
	if _, err := finalFromStream(context.Background(), empty); err == nil || !strings.Contains(err.Error(), "without final") {
		t.Fatalf("finalFromStream(empty) error = %v", err)
	}
	done := make(chan llm.Event)
	close(done)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := finalFromStream(ctx, done); !errors.Is(err, context.Canceled) {
		t.Fatalf("finalFromStream(cancelled) error = %v", err)
	}
	errs := make(chan llm.Event, 1)
	errs <- llm.Event{Kind: llm.EventError, Err: errors.New("stream failed")}
	close(errs)
	if _, err := finalFromStream(context.Background(), errs); err == nil || !strings.Contains(err.Error(), "stream failed") {
		t.Fatalf("finalFromStream(error) = %v", err)
	}
}

func TestRunTeaProgramWithImmediateQuit(t *testing.T) {
	origInput := teaInput
	defer func() { teaInput = origInput }()
	teaInput = strings.NewReader("")
	var stdout bytes.Buffer
	if err := runTeaProgram(quitTeaModel{}, &stdout, io.Discard); err != nil {
		t.Fatalf("runTeaProgram() error = %v", err)
	}
}

func TestLiveAgentModelActiveProfileError(t *testing.T) {
	cfg := config.Default()
	cfg.Active = "missing"
	_, err := liveAgentModel(Runtime{Config: cfg})
	if err == nil || !strings.Contains(err.Error(), "active profile") {
		t.Fatalf("liveAgentModel() error = %v", err)
	}
}

func TestLiveToolDefsAndRunner(t *testing.T) {
	runtime := Runtime{ProjectDir: t.TempDir()}
	runner := liveToolRunner(runtime)
	file := filepath.Join(runtime.ProjectDir, "file.txt")
	calls := []llm.ToolCall{
		{Name: tools.WriteFileName, Arguments: map[string]any{"path": file, "content": "old"}},
		{Name: tools.ReadFileName, Arguments: map[string]any{"path": file}},
		{Name: tools.EditFileName, Arguments: map[string]any{"path": file, "old_string": "old", "new_string": "new"}},
		{Name: tools.PowerShellName, Arguments: map[string]any{"cmd": "", "timeout_seconds": 1}},
		{Name: tools.BashName, Arguments: map[string]any{"cmd": "", "timeout_seconds": "1"}},
		{Name: tools.RepomixrName, Arguments: map[string]any{"url": "bad"}},
		{Name: tools.RecompReferenceName, Arguments: map[string]any{"url": "ftp://bad"}},
	}
	for _, call := range calls {
		if _, err := runner(context.Background(), call); err != nil {
			t.Fatalf("runner(%s) error = %v", call.Name, err)
		}
	}
	if _, err := runner(context.Background(), llm.ToolCall{Name: "missing"}); err == nil || !strings.Contains(err.Error(), "unsupported tool") {
		t.Fatalf("runner(unknown) error = %v", err)
	}
	defs := liveToolDefs(runtime, nil)
	if len(defs) != len(tools.Schemas()) || defs[0].Function.Parameters["type"] != "object" {
		t.Fatalf("liveToolDefs() = %#v", defs)
	}
	if got := defaultArg(llm.ToolCall{}, "output_dir", "fallback"); got != "fallback" {
		t.Fatalf("defaultArg() = %q", got)
	}
}

func TestLiveToolDefsAndRunnerIncludeMCP(t *testing.T) {
	manager := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra", RequireSkill: true}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return fakeMCPClient{result: mcp.CallToolResult{Content: []mcp.ContentItem{{Type: "text", Text: "custom"}}}}, nil
	}))
	if err := manager.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	runtime := Runtime{ProjectDir: t.TempDir(), MCP: manager}
	defs := liveToolDefs(runtime, []string{"ghidra-mcp"})
	found := false
	for _, def := range defs {
		if def.Function.Name == "ghidra.decompile" {
			found = true
		}
	}
	if !found {
		t.Fatalf("liveToolDefs() did not include ghidra.decompile: %#v", defs)
	}
	runner := liveToolRunner(runtime)
	out, err := runner(context.Background(), llm.ToolCall{Name: "ghidra.decompile", Arguments: map[string]any{"address": "0x1000"}})
	if err != nil || out != "custom" {
		t.Fatalf("mcp runner output=%q error=%v", out, err)
	}
	managerErr := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra"}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return fakeMCPClient{result: mcp.CallToolResult{Content: []mcp.ContentItem{{Type: "text", Text: "bad"}}, IsError: true}}, nil
	}))
	if err := managerErr.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect(error manager) = %v", err)
	}
	out, err = liveToolRunner(Runtime{ProjectDir: t.TempDir(), MCP: managerErr})(context.Background(), llm.ToolCall{Name: "ghidra.decompile"})
	if err != nil || !strings.HasPrefix(out, "(mcp: tool returned error:") {
		t.Fatalf("mcp error output=%q error=%v", out, err)
	}
	managerEmpty := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra"}}, mcp.ConnectorFunc(func(context.Context, mcp.ServerConfig) (mcp.Client, error) {
		return fakeMCPClient{result: mcp.CallToolResult{Content: []mcp.ContentItem{{Type: "image", Text: "ignored"}}}}, nil
	}))
	if err := managerEmpty.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect(empty manager) = %v", err)
	}
	out, err = liveToolRunner(Runtime{ProjectDir: t.TempDir(), MCP: managerEmpty})(context.Background(), llm.ToolCall{Name: "ghidra.decompile"})
	if err != nil || out != "(mcp: no text content)" {
		t.Fatalf("mcp empty output=%q error=%v", out, err)
	}
	_, err = liveToolRunner(Runtime{ProjectDir: t.TempDir(), MCP: managerEmpty})(context.Background(), llm.ToolCall{Name: "ghidra.disabled"})
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("disabled mcp error = %v", err)
	}
}

func TestAutoconnectMCP(t *testing.T) {
	manager := mcp.NewManager([]mcp.ServerConfig{{Name: "ghidra", URL: "http://example.invalid"}, {Name: "objdiff", URL: "http://example.invalid"}, {Name: "pcrecomp"}}, mcp.ConnectorFunc(func(ctx context.Context, config mcp.ServerConfig) (mcp.Client, error) {
		if config.Name == "objdiff" {
			return nil, errors.New("blocked connect")
		}
		return fakeMCPClient{}, nil
	}))
	autoconnectMCP(context.Background(), manager, []mcp.ServerConfig{{Name: "ghidra", URL: "http://example.invalid", Autostart: true}, {Name: "objdiff", URL: "http://example.invalid", Autostart: true}, {Name: "pcrecomp", Autostart: true}, {Name: "bizhawk"}})
	if manager.Status("ghidra").State != mcp.StateConnected || manager.Status("objdiff").State != mcp.StateError || manager.Status("pcrecomp").State != mcp.StateConnected {
		t.Fatalf("autoconnect statuses: %#v", manager.AllStatus())
	}
	autoconnectMCP(context.Background(), nil, nil)
}

func TestArgumentHelpers(t *testing.T) {
	call := llm.ToolCall{Arguments: map[string]any{
		"string": "x",
		"float":  1.5,
		"int":    2,
		"nil":    nil,
		"label":  stringLabel("named"),
		"other":  []string{"a"},
	}}
	for key, want := range map[string]string{"string": "x", "float": "1.5", "int": "2", "nil": "", "label": "named", "other": "[a]", "missing": ""} {
		if got := arg(call, key); got != want {
			t.Fatalf("arg(%s) = %q, want %q", key, got, want)
		}
	}
	if timeoutArg(llm.ToolCall{}) != 0 || timeoutArg(llm.ToolCall{Arguments: map[string]any{"timeout_seconds": "bad"}}) != 0 {
		t.Fatal("timeoutArg invalid values should return default marker")
	}
	if timeoutArg(llm.ToolCall{Arguments: map[string]any{"timeout_seconds": 2.0}}) != 2*time.Second {
		t.Fatal("timeoutArg did not parse seconds")
	}
	if got := defaultArg(llm.ToolCall{Arguments: map[string]any{"output_dir": "custom"}}, "output_dir", "fallback"); got != "custom" {
		t.Fatalf("defaultArg(custom) = %q", got)
	}
}

func TestVisibleAgentLinesSkipExistingAndSystem(t *testing.T) {
	messages := []llm.Message{
		{Role: "system", Content: "memory"},
		{Role: "user", Content: "old"},
		{Role: "assistant", Content: "old answer"},
		{Role: "tool", Content: "tool result"},
		{Role: "assistant", Tools: []llm.ToolCall{{ID: "1", Name: "read_file"}}},
	}
	if got := visibleCount(messages); got != 3 {
		t.Fatalf("visibleCount() = %d", got)
	}
	lines := visibleAgentLines(messages, 1)
	if len(lines) != 2 || !strings.Contains(lines[0], "tool: tool result") || !strings.Contains(lines[1], "tool call") {
		t.Fatalf("visibleAgentLines() = %#v", lines)
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
	for _, want := range []string{"assistant: verified: fake runtime smoke complete", "tool: fake tool result", "RecompHamr"} {
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

type modelFunc func(context.Context, []llm.Message) (llm.Message, error)

func (m modelFunc) Next(ctx context.Context, messages []llm.Message) (llm.Message, error) {
	return m(ctx, messages)
}

type stringLabel string

func (s stringLabel) String() string {
	return string(s)
}

type toolAwareScriptModel struct {
	scriptModel
	toolNames []string
}

func (m *toolAwareScriptModel) WithTools(tools []llm.Tool) agent.Model {
	m.toolNames = m.toolNames[:0]
	for _, tool := range tools {
		m.toolNames = append(m.toolNames, tool.Function.Name)
	}
	return m
}

func (m *toolAwareScriptModel) sawTool(name string) bool {
	for _, toolName := range m.toolNames {
		if toolName == name {
			return true
		}
	}
	return false
}

type fakeMCPClient struct {
	result mcp.CallToolResult
}

func (c fakeMCPClient) Tools() []mcp.ToolDef {
	return []mcp.ToolDef{{
		Name:        "decompile",
		Description: "Decompile a function at an address.",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]interface{}{"address": map[string]interface{}{"type": "string"}},
			Required:   []string{"address"},
		},
	}}
}

func (c fakeMCPClient) CallTool(_ context.Context, name string, _ map[string]interface{}) (mcp.CallToolResult, error) {
	if name != "decompile" {
		return mcp.CallToolResult{}, errors.New("unexpected tool " + name)
	}
	if len(c.result.Content) > 0 || c.result.IsError {
		return c.result, nil
	}
	return mcp.CallToolResult{Content: []mcp.ContentItem{{Type: "text", Text: "mcp decompiled"}}}, nil
}

func (c fakeMCPClient) Close() error {
	return nil
}

type quitTeaModel struct{}

func (quitTeaModel) Init() tea.Cmd {
	return tea.Quit
}

func (quitTeaModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return quitTeaModel{}, tea.Quit
}

func (quitTeaModel) View() tea.View {
	return tea.NewView("")
}

func keyText(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}

func keyCode(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func keyCtrl(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Mod: tea.ModCtrl})
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

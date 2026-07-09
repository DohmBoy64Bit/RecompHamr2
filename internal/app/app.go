package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"recomphamr2/internal/agent"
	"recomphamr2/internal/commands"
	"recomphamr2/internal/config"
	"recomphamr2/internal/llm"
	"recomphamr2/internal/mcp"
	"recomphamr2/internal/project"
	"recomphamr2/internal/tui"
)

const (
	// DiagnosticMode prints local implementation status without starting product runtime.
	DiagnosticMode = "--diagnostic"
	// HelpMode prints command-line usage.
	HelpMode = "--help"
)

var (
	bootstrapConfig = config.Bootstrap
	loadMemory      = project.LoadMemory
	newMCPManager   = mcp.NewManager
	mcpBuiltins     = mcp.Builtins
)

// Options controls product runtime composition.
type Options struct {
	// ProjectDir is the repository or user project directory.
	ProjectDir string
	// CustomSkillsDir is the directory used for custom skills.
	CustomSkillsDir string
	// MemoryMaxBytes caps project memory loaded for prompt context.
	MemoryMaxBytes int
}

// Runtime is the composed local product runtime state.
type Runtime struct {
	// ProjectDir is the runtime workspace directory.
	ProjectDir string
	// Config is the loaded or bootstrapped runtime configuration.
	Config *config.Config
	// ConfigCreated reports whether config bootstrap created config.yaml.
	ConfigCreated bool
	// Memory is the optional loaded project memory.
	Memory project.Memory
	// MemoryStatus records verified, unsupported, or blocked memory state.
	MemoryStatus string
	// Commands is the slash-command execution environment.
	Commands commands.Environment
	// MCP is the MCP manager wired into Commands.
	MCP *mcp.Manager
	// TUI is the pure terminal model ready for an interactive adapter.
	TUI tui.Model
}

// SmokeSpec describes one deterministic interactive runtime smoke run.
type SmokeSpec struct {
	// Runtime is the composed runtime under test.
	Runtime Runtime
	// Prompt is the submitted user prompt or slash command.
	Prompt string
	// Model is the fake model used for plain prompt turns.
	Model agent.Model
	// RunTool is the fake tool runner used for model tool calls.
	RunTool agent.ToolRunner
	// CancelBeforeRun cancels the context before the agent turn starts.
	CancelBeforeRun bool
	// RenderWidth controls the golden render width when positive.
	RenderWidth int
}

// SmokeReport records deterministic fake-runtime evidence.
type SmokeReport struct {
	// Messages is the resulting model transcript for plain prompt turns.
	Messages []llm.Message
	// Render is the TUI render after the smoke interaction.
	Render string
	// CommandOutput is populated for slash-command smoke interactions.
	CommandOutput string
	// Cancelled reports whether cancellation blocked the turn.
	Cancelled bool
}

// Run executes the RecompHamr command-line entrypoint and returns a process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 1 {
		switch args[0] {
		case DiagnosticMode:
			fmt.Fprintln(stdout, "recomphamr diagnostic mode")
			fmt.Fprintln(stdout, "phase: architecture-skeleton")
			fmt.Fprintln(stdout, "phase-0: complete: RecompHamr 1.x reference inventory is recorded")
			fmt.Fprintln(stdout, "runtime: product wiring available; use bare recomphamr to compose local runtime")
			return 0
		case HelpMode:
			fmt.Fprintln(stdout, "usage: recomphamr [--diagnostic|--help]")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "no arguments   compose local product runtime and print launch summary")
			fmt.Fprintln(stdout, "--diagnostic   report foundation status without starting product runtime")
			fmt.Fprintln(stdout, "--help         show this help")
			return 0
		}
	}
	if len(args) == 0 {
		runtime, err := ComposeRuntime(Options{})
		if err != nil {
			fmt.Fprintln(stderr, "blocked: "+err.Error())
			return 2
		}
		fmt.Fprintln(stdout, runtime.Summary())
		return 0
	}
	fmt.Fprintln(stderr, "usage: recomphamr [--diagnostic|--help]")
	return 2
}

// ComposeRuntime wires local config, memory, commands, MCP manager, and TUI state.
func ComposeRuntime(opts Options) (Runtime, error) {
	projectDir := strings.TrimSpace(opts.ProjectDir)
	if projectDir == "" {
		projectDir = "."
	}
	customSkillsDir := strings.TrimSpace(opts.CustomSkillsDir)
	if customSkillsDir == "" {
		customSkillsDir = filepath.Join(projectDir, config.DirName, "skills")
	}
	cfg, created, err := bootstrapConfig(projectDir)
	if err != nil {
		return Runtime{ProjectDir: projectDir}, err
	}
	mem, memStatus, err := loadOptionalMemory(projectDir, opts.MemoryMaxBytes)
	if err != nil {
		return Runtime{ProjectDir: projectDir, Config: cfg, ConfigCreated: created}, err
	}
	manager := newMCPManager(mcpBuiltins(), nil)
	env := commands.Environment{
		ProjectDir:      projectDir,
		CustomSkillsDir: customSkillsDir,
		Config:          cfg,
		MCP:             manager,
	}
	model := tui.New(env)
	profile, err := cfg.ActiveProfile()
	if err != nil {
		return Runtime{ProjectDir: projectDir, Config: cfg, ConfigCreated: created}, err
	}
	model.Layout.Mode = "ready"
	model.Layout.ActiveModel = cfg.Active
	model.Layout.MCPStatus = "manager wired"
	model.Layout.ContextStatus = fmt.Sprintf("context=%d", profile.ContextSize)
	model.Layout.MemoryStatus = memStatus
	model.Transcript = append(model.Transcript, "runtime: local product shell composed")
	return Runtime{
		ProjectDir:    projectDir,
		Config:        cfg,
		ConfigCreated: created,
		Memory:        mem,
		MemoryStatus:  memStatus,
		Commands:      env,
		MCP:           manager,
		TUI:           model,
	}, nil
}

// RunSmoke executes a deterministic fake-runtime prompt or slash command.
func RunSmoke(ctx context.Context, spec SmokeSpec) (SmokeReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	runtime := spec.Runtime
	if strings.TrimSpace(spec.Prompt) == "" {
		return SmokeReport{}, fmt.Errorf("smoke prompt is empty")
	}
	width := spec.RenderWidth
	if width <= 0 {
		width = tui.DefaultWidth
	}
	if strings.HasPrefix(strings.TrimSpace(spec.Prompt), "/") {
		model := runtime.TUI.Submit(spec.Prompt)
		render := model.RenderWithLayout(tui.Layout{Width: width, Height: tui.DefaultHeight, Mode: model.Layout.Mode, ActiveModel: model.Layout.ActiveModel, ActiveSkill: model.Layout.ActiveSkill, MCPStatus: model.Layout.MCPStatus, ContextStatus: model.Layout.ContextStatus, PendingTool: model.Layout.PendingTool, MemoryStatus: model.Layout.MemoryStatus})
		output := ""
		if len(model.Transcript) > 0 {
			output = model.Transcript[len(model.Transcript)-1]
		}
		return SmokeReport{Render: render, CommandOutput: output}, nil
	}
	if spec.Model == nil || spec.RunTool == nil {
		return SmokeReport{}, fmt.Errorf("smoke model and tool runner are required for prompt turns")
	}
	if spec.CancelBeforeRun {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		ctx = cancelCtx
	}
	history := []llm.Message{{Role: "user", Content: spec.Prompt}}
	loop := agent.Loop{
		Model:                  spec.Model,
		RunTool:                spec.RunTool,
		ProjectMemory:          runtime.Memory.Content,
		ProjectMemorySource:    runtime.Memory.Path,
		ProjectMemoryMaxTokens: 2048,
		MaxRounds:              8,
		MaxToolCalls:           8,
	}
	messages, err := loop.Run(ctx, history)
	model := runtime.TUI
	model.Transcript = smokeLines(messages)
	render := model.RenderWithLayout(tui.Layout{Width: width, Height: tui.DefaultHeight, Mode: model.Layout.Mode, ActiveModel: model.Layout.ActiveModel, ActiveSkill: model.Layout.ActiveSkill, MCPStatus: model.Layout.MCPStatus, ContextStatus: model.Layout.ContextStatus, PendingTool: model.Layout.PendingTool, MemoryStatus: model.Layout.MemoryStatus})
	report := SmokeReport{Messages: messages, Render: render, Cancelled: errors.Is(err, context.Canceled)}
	return report, err
}

// Summary renders a deterministic local launch summary.
func (r Runtime) Summary() string {
	var b strings.Builder
	b.WriteString("RecompHamr product runtime\n")
	fmt.Fprintf(&b, "project: %s\n", r.ProjectDir)
	configState := "loaded"
	if r.ConfigCreated {
		configState = "created"
	}
	active := "unverified"
	profiles := 0
	if r.Config != nil {
		active = r.Config.Active
		profiles = len(r.Config.Models)
	}
	fmt.Fprintf(&b, "config: %s active=%s profiles=%d\n", configState, active, profiles)
	fmt.Fprintf(&b, "memory: %s\n", r.MemoryStatus)
	fmt.Fprintf(&b, "mcp: manager wired servers=%d autoconnect=false\n", len(mcpBuiltins()))
	fmt.Fprintf(&b, "tui: %s\n", r.TUI.Layout.Mode)
	fmt.Fprintf(&b, "agent: wired for later interactive turns; no model call made during startup")
	return b.String()
}

func smokeLines(messages []llm.Message) []string {
	lines := make([]string, 0, len(messages))
	for _, msg := range messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" && len(msg.Tools) > 0 {
			content = fmt.Sprintf("%d tool call(s)", len(msg.Tools))
		}
		if content == "" {
			content = "(empty)"
		}
		lines = append(lines, msg.Role+": "+content)
	}
	return lines
}

func loadOptionalMemory(projectDir string, maxBytes int) (project.Memory, string, error) {
	mem, err := loadMemory(projectDir, maxBytes)
	if err == nil {
		status := fmt.Sprintf("verified bytes=%d", len(mem.Content))
		if mem.Truncated {
			status += " truncated=true"
		}
		return mem, status, nil
	}
	if errors.Is(err, project.ErrWorkspaceMissing) || errors.Is(err, project.ErrMemoryMissing) {
		return mem, "unsupported: " + err.Error(), nil
	}
	return mem, "", err
}

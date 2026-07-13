package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"recomphamr2/internal/agent"
	"recomphamr2/internal/commands"
	"recomphamr2/internal/config"
	"recomphamr2/internal/llm"
	"recomphamr2/internal/mcp"
	"recomphamr2/internal/project"
	"recomphamr2/internal/tools"
	"recomphamr2/internal/tui"
)

const (
	// DiagnosticMode prints local implementation status without starting product runtime.
	DiagnosticMode = "--diagnostic"
	// HelpMode prints command-line usage.
	HelpMode = "--help"
	// SummaryMode prints deterministic runtime composition evidence without launching the TUI.
	SummaryMode = "--summary"
)

var (
	bootstrapConfig           = config.Bootstrap
	loadMemory                = project.LoadMemory
	newMCPManager             = mcp.NewManager
	mcpBuiltins               = mcp.Builtins
	loadMCPConfig             = mcp.LoadConfigFile
	runProgram                = runTeaProgram
	teaProgramRun             = defaultTeaProgramRun
	teaInput        io.Reader = os.Stdin
	newAgentModel             = liveAgentModel
	newToolRunner             = liveToolRunner
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
	// TUI is the immutable startup snapshot consumed by the terminal frontend.
	TUI tui.Snapshot
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

type streamModel struct {
	client *llm.Client
	tools  []llm.Tool
}

type liveModel struct {
	TUI     tui.Model
	runtime Runtime
	model   agent.Model
	tools   agent.ToolRunner
	cancel  context.CancelFunc
	history []llm.Message
}

type agentResult struct {
	messages      []llm.Message
	visibleOffset int
	err           error
}

// Run executes the RecompHamr command-line entrypoint and returns a process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 1 {
		switch args[0] {
		case DiagnosticMode:
			fmt.Fprintln(stdout, "recomphamr diagnostic mode")
			fmt.Fprintln(stdout, "phase: architecture-skeleton")
			fmt.Fprintln(stdout, "phase-0: complete: RecompHamr 1.x reference inventory is recorded")
			fmt.Fprintln(stdout, "runtime: product wiring available; use bare recomphamr to launch the terminal app")
			return 0
		case HelpMode:
			fmt.Fprintln(stdout, "usage: recomphamr [--diagnostic|--help|--summary]")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "no arguments   launch the interactive terminal app")
			fmt.Fprintln(stdout, "--diagnostic   report foundation status without starting product runtime")
			fmt.Fprintln(stdout, "--help         show this help")
			fmt.Fprintln(stdout, "--summary      compose local runtime and print launch summary")
			return 0
		case SummaryMode:
			runtime, err := ComposeRuntime(Options{})
			if err != nil {
				fmt.Fprintln(stderr, "blocked: "+err.Error())
				return 2
			}
			fmt.Fprintln(stdout, runtime.Summary())
			return 0
		}
	}
	if len(args) == 0 {
		runtime, err := ComposeRuntime(Options{})
		if err != nil {
			fmt.Fprintln(stderr, "blocked: "+err.Error())
			return 2
		}
		if err := Launch(runtime, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, "blocked: "+err.Error())
			return 2
		}
		return 0
	}
	fmt.Fprintln(stderr, "usage: recomphamr [--diagnostic|--help|--summary]")
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
	configs, err := runtimeMCPConfigs(projectDir)
	if err != nil {
		return Runtime{ProjectDir: projectDir, Config: cfg, ConfigCreated: created}, err
	}
	manager := newMCPManager(configs, mcp.AutoConnector{})
	autoconnectMCP(context.Background(), manager, configs)
	env := commands.Environment{
		ProjectDir:      projectDir,
		CustomSkillsDir: customSkillsDir,
		Config:          cfg,
		MCP:             manager,
	}
	profile, err := cfg.ActiveProfile()
	if err != nil {
		return Runtime{ProjectDir: projectDir, Config: cfg, ConfigCreated: created}, err
	}
	snapshot := tui.Snapshot{Env: env, Mode: "ready", ActiveModel: cfg.Active, ActiveSkill: "none", MCPStatus: "manager wired", ContextStatus: fmt.Sprintf("context=%d", profile.ContextSize), PendingTool: "none", MemoryStatus: memStatus}
	return Runtime{
		ProjectDir:    projectDir,
		Config:        cfg,
		ConfigCreated: created,
		Memory:        mem,
		MemoryStatus:  memStatus,
		Commands:      env,
		MCP:           manager,
		TUI:           snapshot,
	}, nil
}

// Launch starts the interactive Bubble Tea application with live agent wiring.
func Launch(runtime Runtime, stdout io.Writer, stderr io.Writer) error {
	model, err := newAgentModel(runtime)
	if err != nil {
		return err
	}
	runner := newToolRunner(runtime)
	app := liveModel{
		TUI:     tui.New(runtime.TUI),
		runtime: runtime,
		model:   model,
		tools:   runner,
	}
	return runProgram(app, stdout, stderr)
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
		output, env := commands.Execute(runtime.Commands, spec.Prompt)
		snapshot := runtime.TUI
		snapshot.Env = env
		render := tui.Render(snapshot, []tui.TranscriptEntry{tui.ParseEntry(output)}, width, tui.DefaultHeight)
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
	entries := make([]tui.TranscriptEntry, 0, len(messages))
	for _, line := range smokeLines(messages) {
		entries = append(entries, tui.ParseEntry(line))
	}
	render := tui.Render(runtime.TUI, entries, width, tui.DefaultHeight)
	report := SmokeReport{Messages: messages, Render: render, Cancelled: errors.Is(err, context.Canceled)}
	return report, err
}

// Next sends one OpenAI-compatible streaming turn and returns the final assistant message.
func (m streamModel) Next(ctx context.Context, messages []llm.Message) (llm.Message, error) {
	if m.client == nil {
		return llm.Message{}, fmt.Errorf("LLM client is not configured")
	}
	return finalFromStream(ctx, m.client.Chat(ctx, messages, m.tools))
}

func finalFromStream(ctx context.Context, events <-chan llm.Event) (llm.Message, error) {
	var final *llm.Message
	for event := range events {
		if event.Kind == llm.EventError {
			return llm.Message{}, event.Err
		}
		if event.Kind == llm.EventDone {
			final = event.Final
		}
	}
	if err := ctx.Err(); err != nil {
		return llm.Message{}, err
	}
	if final == nil {
		return llm.Message{}, fmt.Errorf("blocked: model stream ended without final message")
	}
	return *final, nil
}

// Init initializes the live Bubble Tea model.
func (m liveModel) Init() tea.Cmd {
	return m.TUI.Init()
}

// Update translates terminal events and runs app-owned agent side effects.
func (m liveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case agentResult:
		return m.applyAgentResult(typed), nil
	case tui.IntentMsg:
		return m.handleIntent(typed)
	default:
		updated, cmd := m.TUI.Update(msg)
		m.TUI = updated.(tui.Model)
		return m, cmd
	}
}

func (m liveModel) handleIntent(intent tui.IntentMsg) (tea.Model, tea.Cmd) {
	switch intent.Kind {
	case tui.IntentSubmit:
		m = m.applyTUI(tui.TranscriptMsg{Entries: []tui.TranscriptEntry{{Kind: tui.TranscriptUser, Text: intent.Value}}})
		return m.startAgentTurn(intent.Value)
	case tui.IntentCommand:
		return m.executeCommand(intent.Value), nil
	case tui.IntentCancel:
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}
		return m, nil
	case tui.IntentModel:
		return m.executeCommand("/models " + intent.Value), nil
	case tui.IntentSkill:
		return m.executeCommand("/skill " + intent.Value), nil
	case tui.IntentMCP:
		return m.executeCommand("/mcp tools " + intent.Value), nil
	case tui.IntentQuit:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m liveModel) executeCommand(text string) liveModel {
	output, env := commands.Execute(m.runtime.Commands, text)
	m.runtime.Commands = env
	snapshot := m.TUI.Snapshot()
	snapshot.Env = env
	if strings.TrimSpace(text) == "/clear" {
		m = m.applyTUI(tui.ClearTranscriptMsg{})
	}
	m = m.applyTUI(tui.SnapshotMsg{Snapshot: snapshot})
	return m.applyTUI(tui.TranscriptMsg{Entries: []tui.TranscriptEntry{tui.ParseEntry(output)}})
}

func (m liveModel) applyTUI(msg tea.Msg) liveModel {
	updated, _ := m.TUI.Update(msg)
	m.TUI = updated.(tui.Model)
	return m
}

// View renders the live Bubble Tea model.
func (m liveModel) View() tea.View {
	return m.TUI.View()
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
	serverCount := len(mcpBuiltins())
	if r.MCP != nil {
		serverCount = len(r.MCP.AllStatus())
	}
	fmt.Fprintf(&b, "mcp: manager wired servers=%d autoconnect=%s\n", serverCount, autoconnectStatus(r.MCP))
	fmt.Fprintf(&b, "tui: %s\n", r.TUI.Mode)
	fmt.Fprintf(&b, "agent: wired for interactive turns; no model call made during summary")
	return b.String()
}

func runTeaProgram(model tea.Model, stdout io.Writer, _ io.Writer) error {
	_, err := teaProgramRun(model, stdout)
	return err
}

func defaultTeaProgramRun(model tea.Model, stdout io.Writer) (tea.Model, error) {
	return tea.NewProgram(model, tea.WithOutput(stdout), tea.WithInput(teaInput)).Run()
}

func liveAgentModel(runtime Runtime) (agent.Model, error) {
	if runtime.Config == nil {
		return nil, fmt.Errorf("runtime config is not loaded")
	}
	profile, err := runtime.Config.ActiveProfile()
	if err != nil {
		return nil, err
	}
	return streamModel{
		client: llm.NewClient(profile.URL, profile.LLM, profile.Key),
		tools:  liveToolDefs(runtime, nil),
	}, nil
}

// WithTools returns a stream model copy configured with the tools for one turn.
func (m streamModel) WithTools(tools []llm.Tool) agent.Model {
	m.tools = append([]llm.Tool(nil), tools...)
	return m
}

func liveToolDefs(runtime Runtime, activeSkills []string) []llm.Tool {
	schemas := tools.Schemas()
	defs := make([]llm.Tool, 0, len(schemas))
	for _, schema := range schemas {
		defs = append(defs, llm.Tool{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        schema.Name,
				Description: schema.Description,
				Parameters:  toolParameters(schema.Parameters),
			},
		})
	}
	if runtime.MCP != nil {
		for _, tool := range runtime.MCP.ToolsForSkills(activeSkills) {
			defs = append(defs, llm.Tool{
				Type: "function",
				Function: llm.FunctionDef{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema.Map(),
				},
			})
		}
	}
	return defs
}

func toolParameters(names []string) map[string]any {
	properties := map[string]any{}
	required := make([]string, 0, len(names))
	for _, name := range names {
		properties[name] = map[string]any{"type": "string"}
		if name != "timeout_seconds" && name != "output_dir" {
			required = append(required, name)
		}
	}
	return map[string]any{"type": "object", "properties": properties, "required": required}
}

func liveToolRunner(runtime Runtime) agent.ToolRunner {
	repoDir := filepath.Join(runtime.ProjectDir, config.DirName, "repos")
	referenceDir := filepath.Join(runtime.ProjectDir, config.DirName, "references")
	return func(ctx context.Context, call llm.ToolCall) (string, error) {
		switch call.Name {
		case tools.PowerShellName:
			return tools.PowerShell(ctx, arg(call, "cmd"), timeoutArg(call)), nil
		case tools.BashName:
			return tools.Bash(ctx, arg(call, "cmd"), timeoutArg(call)), nil
		case tools.ReadFileName:
			return tools.ReadFile(arg(call, "path")), nil
		case tools.WriteFileName:
			return tools.WriteFile(arg(call, "path"), arg(call, "content")), nil
		case tools.EditFileName:
			return tools.EditFile(arg(call, "path"), arg(call, "old_string"), arg(call, "new_string")), nil
		case tools.RepomixrName:
			return tools.Repomixr(ctx, arg(call, "url"), defaultArg(call, "output_dir", repoDir)), nil
		case tools.RecompReferenceName:
			return tools.RecompReference(ctx, http.DefaultClient, arg(call, "url"), defaultArg(call, "output_dir", referenceDir)), nil
		default:
			if runtime.MCP != nil && strings.Contains(call.Name, ".") {
				result, err := runtime.MCP.CallTool(ctx, call.Name, call.Arguments)
				if err != nil {
					return "", err
				}
				text := strings.TrimSpace(result.Text())
				if text == "" {
					text = "(mcp: no text content)"
				}
				if result.IsError {
					text = "(mcp: tool returned error: " + text + ")"
				}
				return text, nil
			}
			return "", tools.Unsupported(call.Name)
		}
	}
}

func (m liveModel) startAgentTurn(prompt string) (liveModel, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	snapshot := m.TUI.Snapshot()
	snapshot.Mode = "thinking"
	snapshot.PendingTool = "agent"
	snapshot.Status = "running prompt"
	m = m.applyTUI(tui.SnapshotMsg{Snapshot: snapshot})
	m.history = append(m.history, llm.Message{Role: "user", Content: prompt})
	history := append([]llm.Message(nil), m.history...)
	visibleOffset := visibleCount(history)
	model := m.model
	if configurable, ok := m.model.(interface{ WithTools([]llm.Tool) agent.Model }); ok {
		model = configurable.WithTools(liveToolDefs(m.runtime, snapshot.Env.ActiveSkills))
	}
	loop := agent.Loop{
		Model:                  model,
		RunTool:                m.tools,
		ProjectMemory:          m.runtime.Memory.Content,
		ProjectMemorySource:    m.runtime.Memory.Path,
		ProjectMemoryMaxTokens: 2048,
		MaxRounds:              8,
		MaxToolCalls:           8,
	}
	return m, func() tea.Msg {
		messages, err := loop.Run(ctx, history)
		if err != nil && errors.Is(err, context.Canceled) {
			return agentResult{messages: messages, visibleOffset: visibleOffset, err: err}
		}
		return agentResult{messages: messages, visibleOffset: visibleOffset, err: err}
	}
}

func autoconnectMCP(ctx context.Context, manager *mcp.Manager, configs []mcp.ServerConfig) {
	if manager == nil {
		return
	}
	for _, config := range configs {
		if config.Autostart {
			_ = manager.Connect(ctx, config.Name)
		}
	}
}

func runtimeMCPConfigs(projectDir string) ([]mcp.ServerConfig, error) {
	cfg, err := loadMCPConfig(filepath.Join(projectDir, config.DirName, "mcp.json"))
	if err != nil {
		return nil, err
	}
	return mcp.MergeConfigs(mcpBuiltins(), cfg), nil
}

func autoconnectStatus(manager *mcp.Manager) string {
	if manager == nil {
		return "none"
	}
	connected := 0
	for _, status := range manager.AllStatus() {
		if status.State == mcp.StateConnected {
			connected++
		}
	}
	if connected == 0 {
		return "false"
	}
	return fmt.Sprintf("connected=%d", connected)
}

func (m liveModel) applyAgentResult(result agentResult) liveModel {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	snapshot := m.TUI.Snapshot()
	snapshot.PendingTool = "none"
	snapshot.Mode = "ready"
	snapshot.Status = ""
	var entries []tui.TranscriptEntry
	if result.err != nil {
		if errors.Is(result.err, context.Canceled) {
			snapshot.Status = "cancelled"
		} else {
			snapshot.Status = "blocked"
			entries = append(entries, tui.TranscriptEntry{Kind: tui.TranscriptBlocked, Text: result.err.Error()})
		}
	}
	m.history = result.messages
	for _, line := range visibleAgentLines(result.messages, result.visibleOffset) {
		entries = append(entries, tui.ParseEntry(line))
	}
	m = m.applyTUI(tui.SnapshotMsg{Snapshot: snapshot})
	m = m.applyTUI(tui.TranscriptMsg{Entries: entries})
	return m
}

func visibleAgentLines(messages []llm.Message, offset int) []string {
	var lines []string
	seen := 0
	for _, msg := range messages {
		if msg.Role == "system" || msg.Role == "user" {
			continue
		}
		if seen < offset {
			seen++
			continue
		}
		for _, line := range smokeLines([]llm.Message{msg}) {
			lines = append(lines, line)
		}
		seen++
	}
	return lines
}

func visibleCount(messages []llm.Message) int {
	count := 0
	for _, msg := range messages {
		if msg.Role != "system" && msg.Role != "user" {
			count++
		}
	}
	return count
}

func arg(call llm.ToolCall, name string) string {
	if call.Arguments == nil {
		return ""
	}
	switch value := call.Arguments[name].(type) {
	case string:
		return value
	case fmt.Stringer:
		return value.String()
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case int:
		return strconv.Itoa(value)
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}
}

func defaultArg(call llm.ToolCall, name string, fallback string) string {
	value := strings.TrimSpace(arg(call, name))
	if value == "" {
		return fallback
	}
	return value
}

func timeoutArg(call llm.ToolCall) time.Duration {
	raw := strings.TrimSpace(arg(call, "timeout_seconds"))
	if raw == "" {
		return 0
	}
	seconds, err := strconv.ParseFloat(raw, 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
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

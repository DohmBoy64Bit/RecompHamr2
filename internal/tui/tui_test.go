package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"recomphamr2/internal/commands"
)

func TestSubmitAndRender(t *testing.T) {
	model := New(commands.Environment{ProjectDir: t.TempDir()})
	model = model.Submit("hello")
	model = model.Submit("/help")
	view := model.Render()
	if !strings.Contains(view, "user: hello") || !strings.Contains(view, "/models") || !strings.Contains(view, "Build *") {
		t.Fatalf("Render() = %q", view)
	}
}

func TestRenderWideLayout(t *testing.T) {
	model := New(commands.Environment{})
	model.Transcript = []string{"user: inspect binary", "assistant: gather evidence"}
	model.Layout = Layout{
		Width:         120,
		Mode:          "evidence",
		ActiveModel:   "lmstudio-amd",
		ActiveSkill:   "n64-decomp",
		MCPStatus:     "ghidra gated",
		ContextStatus: "42k / 131k",
		PendingTool:   "read_file",
		MemoryStatus:  "fresh",
	}
	view := model.Render()
	for _, want := range []string{"user: inspect binary", "assistant: gather evidence", "Build * lmstudio-amd", "skill n64-decomp", "mcp ghidra gated", "/ commands"} {
		if !strings.Contains(view, want) {
			t.Fatalf("wide render missing %q:\n%s", want, view)
		}
	}
}

func TestRenderCompactLayout(t *testing.T) {
	model := New(commands.Environment{})
	model.Transcript = []string{"user: narrow"}
	view := model.RenderWithLayout(Layout{Width: 80, Mode: "run", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "off", ContextStatus: "ok", PendingTool: "none", MemoryStatus: "fresh"})
	for _, want := range []string{"user: narrow", "Build * local", "skill core-re", "mcp off", "/ commands"} {
		if !strings.Contains(view, want) {
			t.Fatalf("compact render missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "WORKBENCH") {
		t.Fatalf("compact render should not use the old debug-board sections:\n%s", view)
	}
}

func TestDefaultLayoutAndImprovements(t *testing.T) {
	layout := DefaultLayout()
	if layout.Width != DefaultWidth || layout.MCPStatus != "gated" || layout.MemoryStatus != "refreshed" {
		t.Fatalf("DefaultLayout() = %+v", layout)
	}
	improvements := Improvements()
	if len(improvements) != 4 {
		t.Fatalf("len(Improvements()) = %d, want 3", len(improvements))
	}
	for _, improvement := range improvements {
		if strings.TrimSpace(improvement) == "" {
			t.Fatalf("empty improvement in %v", improvements)
		}
	}
}

func TestRenderWithLayoutDefaultWidth(t *testing.T) {
	model := New(commands.Environment{})
	view := model.RenderWithLayout(Layout{})
	if !strings.Contains(view, "RECOMP HAMR") || !strings.Contains(view, "Ask RecompHamr") {
		t.Fatalf("zero-width layout should default to wide render:\n%s", view)
	}
}

func TestRenderStartupAndHelperTokens(t *testing.T) {
	model := New(commands.Environment{})
	view := model.Render()
	for _, want := range []string{"RECOMP HAMR", "evidence-backed reconstruction", "Ask RecompHamr", "/ commands", "? help"} {
		if !strings.Contains(view, want) {
			t.Fatalf("startup render missing %q:\n%s", want, view)
		}
	}
	if got := chip(""); got != "[unverified]" {
		t.Fatalf("chip(empty) = %q", got)
	}
	if got := centerText(12, "HAMR"); got != "    HAMR" {
		t.Fatalf("centerText = %q", got)
	}
	if got := centerText(3, "HAMR"); got != "HAMR" {
		t.Fatalf("centerText narrow = %q", got)
	}
	if width, height := bubbleSize(Layout{Width: 10, Height: 0}); width != 40 || height != DefaultHeight {
		t.Fatalf("bubbleSize = %d,%d", width, height)
	}
	if width, height := bubbleSize(Layout{}); width != DefaultWidth || height != DefaultHeight {
		t.Fatalf("bubbleSize default = %d,%d", width, height)
	}
	if got := launcherPanelWidth(120); got != 84 {
		t.Fatalf("launcherPanelWidth wide = %d", got)
	}
	if got := launcherPanelWidth(42); got != 38 {
		t.Fatalf("launcherPanelWidth compact = %d", got)
	}
	if got := launcherPanelWidth(20); got != 36 {
		t.Fatalf("launcherPanelWidth tiny = %d", got)
	}
	if got := launcherTopPadding(40); got != 6 {
		t.Fatalf("launcherTopPadding tall = %d", got)
	}
	if got := launcherTopPadding(32); got != 6 {
		t.Fatalf("launcherTopPadding exact = %d", got)
	}
	if got := launcherTopPadding(20); got != 4 {
		t.Fatalf("launcherTopPadding medium = %d", got)
	}
	if got := launcherTopPadding(18); got != 1 {
		t.Fatalf("launcherTopPadding short = %d", got)
	}
	if got := renderWidth(200); got != 110 {
		t.Fatalf("renderWidth large = %d", got)
	}
	if got := renderWidth(10); got != 32 {
		t.Fatalf("renderWidth small = %d", got)
	}
}

func TestRenderCompactStartup(t *testing.T) {
	model := New(commands.Environment{})
	view := model.RenderWithLayout(Layout{Width: 72, Mode: "ready", ActiveModel: "local", ActiveSkill: "none", MCPStatus: "gated", ContextStatus: "ok", PendingTool: "none", MemoryStatus: "fresh"})
	for _, want := range []string{"RecompHamr", "RE / decomp / recomp", "Ask RecompHamr", "ready  local  ready", "/ commands"} {
		if !strings.Contains(view, want) {
			t.Fatalf("compact startup missing %q:\n%s", want, view)
		}
	}
}

func TestTranscriptBlocks(t *testing.T) {
	lines := []string{
		"user: hi",
		"assistant: hello",
		"tool: output",
		"mcp ghidra connected",
		"mcp: tool returned error",
		"verification: hashes match",
		"blocked: denied",
		"unsupported: later",
		"unverified: missing evidence",
		"status: cancelled",
		"paste: paste-1 (12 bytes)",
		"plain note",
	}
	wantLabels := []string{"user", "assistant", "tool", "mcp", "mcp", "verification", "blocked", "unsupported", "unverified", "status", "attachment", "note"}
	for i, line := range lines {
		got := transcriptBlock(line)
		if !strings.HasPrefix(got, wantLabels[i]) || !strings.Contains(got, line) {
			t.Fatalf("transcriptBlock(%q) = %q", line, got)
		}
	}
	model := New(commands.Environment{})
	model.Transcript = lines
	view := model.RenderWithLayout(Layout{Width: 80, Mode: "ready", ActiveModel: "local", ActiveSkill: "none", MCPStatus: "gated", ContextStatus: "ok", PendingTool: "none", MemoryStatus: "fresh"})
	for _, want := range []string{"verification verification: hashes match", "blocked     blocked: denied", "attachment  paste: paste-1", "note        plain note"} {
		if !strings.Contains(view, want) {
			t.Fatalf("transcript render missing %q:\n%s", want, view)
		}
	}
}

func TestCompleteCommand(t *testing.T) {
	got := CompleteCommand("/mod")
	if len(got) != 1 || got[0] != "/models" {
		t.Fatalf("CompleteCommand() = %v, want /models", got)
	}
	if got := CompleteCommand("/zzz"); len(got) != 0 {
		t.Fatalf("CompleteCommand() = %v, want none", got)
	}
}

func TestPaletteRowsAndTabCompletion(t *testing.T) {
	model := New(commands.Environment{})
	model.Composer = "/m"
	rows := model.PaletteRows()
	if len(rows) == 0 || !strings.Contains(rows[0], "> /models") || strings.Contains(rows[0], "usage:") {
		t.Fatalf("PaletteRows() = %#v", rows)
	}
	view := model.Render()
	if !strings.Contains(view, "Command Palette") || !strings.Contains(view, "> /models") {
		t.Fatalf("palette render = %q", view)
	}
	model, action := model.Update(Event{Key: KeyTab})
	if action != ActionNone || model.Composer != "/models " || !strings.Contains(model.Status, "completed command") {
		t.Fatalf("tab completion action=%s model=%+v", action, model)
	}
	model.Composer = "/unknown"
	model, _ = model.Update(Event{Key: KeyTab})
	if !strings.Contains(model.Status, "unverified") {
		t.Fatalf("tab no-match status=%q", model.Status)
	}
	model.Composer = "plain"
	if rows := model.PaletteRows(); rows != nil {
		t.Fatalf("PaletteRows(plain) = %#v", rows)
	}
	before := model.Composer
	model, _ = model.Update(Event{Key: KeyTab})
	if model.Composer != before {
		t.Fatalf("tab plain changed composer from %q to %q", before, model.Composer)
	}
	model.Composer = "/help /models"
	model, _ = model.Update(Event{Key: KeyTab})
	if model.Composer != "/help /models" {
		t.Fatalf("tab with suffix composer=%q", model.Composer)
	}
}

func TestUpdateTypingSubmitAndHistory(t *testing.T) {
	model := New(commands.Environment{})
	model, action := model.Update(Event{Text: "hello"})
	if action != ActionNone || model.Composer != "hello" {
		t.Fatalf("typing action=%s composer=%q", action, model.Composer)
	}
	model, action = model.Update(Event{Key: KeyEnter})
	if action != ActionSubmit || model.Composer != "" || len(model.History) != 1 || !strings.Contains(model.Render(), "user: hello") {
		t.Fatalf("submit action=%s model=%+v view=%q", action, model, model.Render())
	}
	model, action = model.Update(Event{Key: KeyUp})
	if action != ActionNone || model.Composer != "hello" {
		t.Fatalf("history up action=%s composer=%q", action, model.Composer)
	}
	model, _ = model.Update(Event{Key: KeyDown})
	if model.Composer != "" {
		t.Fatalf("history down composer=%q, want empty", model.Composer)
	}
}

func TestUpdateEditingResizeAndUnsupportedKey(t *testing.T) {
	model := New(commands.Environment{})
	model, action := model.Update(Event{Text: "ab世", Width: 70, Height: 20})
	if action != ActionNone || model.Layout.Width != 70 || model.Layout.Height != 20 {
		t.Fatalf("resize/type action=%s layout=%+v", action, model.Layout)
	}
	model, _ = model.Update(Event{Key: KeyBackspace})
	if model.Composer != "ab" {
		t.Fatalf("backspace composer=%q", model.Composer)
	}
	model, _ = model.Update(Event{Key: KeyBackspace})
	model, _ = model.Update(Event{Key: KeyBackspace})
	model, _ = model.Update(Event{Key: KeyBackspace})
	if model.Composer != "" {
		t.Fatalf("empty backspace composer=%q", model.Composer)
	}
	model, action = model.Update(Event{Key: "f13"})
	if action != ActionNone || !strings.Contains(model.Status, "unsupported key") {
		t.Fatalf("unsupported action=%s status=%q", action, model.Status)
	}
}

func TestCancellationAndQuitKeys(t *testing.T) {
	model := New(commands.Environment{})
	model.Layout.Mode = "thinking"
	model.Layout.PendingTool = "read_file"
	model, action := model.Update(Event{Key: KeyCtrlC})
	if action != ActionCancel || model.Layout.Mode != "idle" || model.Layout.PendingTool != "none" || !strings.Contains(model.Render(), "cancelled") {
		t.Fatalf("cancel action=%s model=%+v view=%q", action, model, model.Render())
	}
	model, action = model.Update(Event{Key: KeyCtrlC})
	if action != ActionNone || !model.QuitArmed || !strings.Contains(model.Status, "again") {
		t.Fatalf("quit arm action=%s model=%+v", action, model)
	}
	model, action = model.Update(Event{Key: KeyEsc})
	if action != ActionNone || model.QuitArmed || model.Status != "" {
		t.Fatalf("escape action=%s model=%+v", action, model)
	}
	model, _ = model.Update(Event{Key: KeyCtrlC})
	model, action = model.Update(Event{Key: KeyCtrlC})
	if action != ActionQuit || model.Status != "quit" {
		t.Fatalf("double ctrl-c action=%s status=%q", action, model.Status)
	}
	model = New(commands.Environment{})
	model, action = model.Update(Event{Key: KeyCtrlD})
	if action != ActionQuit || model.Status != "quit" {
		t.Fatalf("ctrl-d action=%s status=%q", action, model.Status)
	}
}

func TestPasteChipsComposerPaletteAndDebug(t *testing.T) {
	model := New(commands.Environment{})
	model, _ = model.Update(Event{Paste: "small"})
	if model.Composer != "small" {
		t.Fatalf("small paste composer=%q", model.Composer)
	}
	model = model.Paste("line one\nline two")
	if len(model.Attachments) != 1 || !strings.Contains(model.Render(), "[paste-1") {
		t.Fatalf("large paste model=%+v view=%q", model, model.Render())
	}
	model, action := model.Update(Event{Key: KeyEnter})
	if action != ActionSubmit || len(model.Attachments) != 0 || !strings.Contains(model.Transcript[len(model.Transcript)-1], "[paste-1") {
		t.Fatalf("submit paste action=%s model=%+v", action, model)
	}
	model.Composer = "/s"
	if got := model.Palette(); len(got) == 0 {
		t.Fatalf("Palette() = %v, want slash matches", got)
	}
	model.Composer = "plain"
	if got := model.Palette(); got != nil {
		t.Fatalf("Palette() = %v, want nil", got)
	}
	model.DebugSecrets = []string{"secret"}
	model = model.Debug("token secret")
	if len(model.DebugLog) != 0 {
		t.Fatalf("disabled debug log = %v", model.DebugLog)
	}
	model.DebugEnabled = true
	model = model.Debug("token secret")
	if !strings.Contains(model.Render(), "[REDACTED]") || strings.Contains(model.Render(), "token secret") {
		t.Fatalf("debug redaction render=%q", model.Render())
	}
}

func TestRenderMultilineComposerAndEmptySubmit(t *testing.T) {
	model := New(commands.Environment{})
	model.Composer = "one\ntwo"
	view := model.Render()
	if !strings.Contains(view, "composer > one") || !strings.Contains(view, "two") {
		t.Fatalf("multiline composer render=%q", view)
	}
	before := model
	model, action := model.Update(Event{Key: KeyEnter})
	if action != ActionSubmit || len(model.History) != 1 {
		t.Fatalf("multiline submit action=%s history=%v", action, model.History)
	}
	empty := New(commands.Environment{})
	empty, action = empty.Update(Event{Key: KeyEnter})
	if action != ActionNone || len(empty.Transcript) != 0 {
		t.Fatalf("empty submit action=%s model=%+v", action, empty)
	}
	if before.Composer == "" {
		t.Fatal("test setup lost composer")
	}
}

func TestRenderAttachmentOnlyComposer(t *testing.T) {
	model := New(commands.Environment{}).Paste("line one\nline two")
	view := model.Render()
	if !strings.Contains(view, "composer > [paste-1") {
		t.Fatalf("attachment-only composer render=%q", view)
	}
}

func TestPromptHistoryEmptyAndClamp(t *testing.T) {
	model := New(commands.Environment{})
	model, _ = model.Update(Event{Key: KeyUp})
	if model.Composer != "" {
		t.Fatalf("empty history composer=%q", model.Composer)
	}
	model = model.Submit("first")
	model = model.Submit("second")
	model, _ = model.Update(Event{Key: KeyUp})
	model, _ = model.Update(Event{Key: KeyUp})
	model, _ = model.Update(Event{Key: KeyUp})
	if model.Composer != "first" {
		t.Fatalf("history clamp up composer=%q", model.Composer)
	}
	model, _ = model.Update(Event{Key: KeyDown})
	model, _ = model.Update(Event{Key: KeyDown})
	model, _ = model.Update(Event{Key: KeyDown})
	if model.Composer != "" || model.HistoryIndex != len(model.History) {
		t.Fatalf("history clamp down composer=%q index=%d", model.Composer, model.HistoryIndex)
	}
}

func TestHelpersForPasteAndSubmission(t *testing.T) {
	if isLargePaste(strings.Repeat("x", LargePasteThreshold-1)) {
		t.Fatal("paste below threshold should be inline")
	}
	if !isLargePaste(strings.Repeat("x", LargePasteThreshold)) || !isLargePaste("a\nb") {
		t.Fatal("large or multiline paste should be a chip")
	}
	if got := submissionText("", []Attachment{{Name: "paste-1", Content: "abc"}}); got != "[paste-1 3 bytes]" {
		t.Fatalf("submissionText attachment only = %q", got)
	}
	if got := trimLastRune(""); got != "" {
		t.Fatalf("trimLastRune empty = %q", got)
	}
}

func BenchmarkRenderWideLayout(b *testing.B) {
	model := New(commands.Environment{})
	for i := 0; i < 200; i++ {
		model.Transcript = append(model.Transcript, "assistant: verified evidence line")
	}
	model.Layout = Layout{Width: 120, Mode: "run", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "gated", ContextStatus: "32k", PendingTool: "none", MemoryStatus: "fresh"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if view := model.Render(); len(view) == 0 {
			b.Fatal("Render returned empty view")
		}
	}
}

func BenchmarkRenderCompactLayout(b *testing.B) {
	model := New(commands.Environment{})
	for i := 0; i < 200; i++ {
		model.Transcript = append(model.Transcript, "tool: bounded result")
	}
	layout := Layout{Width: 80, Mode: "run", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "gated", ContextStatus: "32k", PendingTool: "none", MemoryStatus: "fresh"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if view := model.RenderWithLayout(layout); len(view) == 0 {
			b.Fatal("RenderWithLayout returned empty view")
		}
	}
}

func TestBubbleModelAdapter(t *testing.T) {
	bubble := NewBubble(commands.Environment{})
	if cmd := bubble.Init(); cmd != nil {
		t.Fatalf("Init() = %v, want nil", cmd)
	}
	updated, cmd := bubble.Update(tea.WindowSizeMsg{Width: 72, Height: 24})
	if cmd != nil {
		t.Fatalf("Update(window) cmd = %v, want nil", cmd)
	}
	bubble = updated.(BubbleModel)
	if bubble.State.Layout.Width != 72 || bubble.State.Layout.Height != 24 || bubble.LastAction != ActionNone {
		t.Fatalf("window update bubble=%+v", bubble)
	}
	updated, _ = bubble.Update(keyText("h"))
	bubble = updated.(BubbleModel)
	if bubble.State.Composer != "h" {
		t.Fatalf("rune update composer=%q", bubble.State.Composer)
	}
	updated, _ = bubble.Update(keyCode(tea.KeyEnter))
	bubble = updated.(BubbleModel)
	if bubble.LastAction != ActionSubmit || !strings.Contains(bubble.View().Content, "user: h") {
		t.Fatalf("enter update bubble=%+v view=%q", bubble, bubble.View().Content)
	}
	updated, _ = bubble.Update(struct{}{})
	if updated.(BubbleModel).LastAction != bubble.LastAction {
		t.Fatalf("ignored message changed action: before=%s after=%s", bubble.LastAction, updated.(BubbleModel).LastAction)
	}
	if key := bubbleKey("home"); key != "home" {
		t.Fatalf("bubbleKey(home) = %q", key)
	}
	if key := bubbleKey("tab"); key != KeyTab {
		t.Fatalf("bubbleKey(tab) = %q", key)
	}
}

func TestBubbleTeaV2StyledViewFieldsAndPaste(t *testing.T) {
	bubble := NewBubble(commands.Environment{})
	bubble.State.Composer = "/m"
	view := bubble.View()
	if !view.AltScreen || view.MouseMode != tea.MouseModeNone || !view.ReportFocus || view.WindowTitle != "RecompHamr" {
		t.Fatalf("view fields not set for Bubble Tea v2: %+v", view)
	}
	if view.Cursor == nil || view.Cursor.Shape != tea.CursorBar || !view.Cursor.Blink {
		t.Fatalf("cursor not configured: %+v", view.Cursor)
	}
	if view.Cursor.X <= 0 || view.Cursor.Y <= 0 {
		t.Fatalf("startup cursor should land in launcher composer, got %+v", view.Cursor)
	}
	if !strings.Contains(view.Content, "\x1b[") || !strings.Contains(view.Content, "COMMAND PALETTE") {
		t.Fatalf("styled content missing ANSI or palette:\n%q", view.Content)
	}
	updated, _ := bubble.Update(tea.PasteMsg{Content: "line one\nline two"})
	bubble = updated.(BubbleModel)
	if len(bubble.State.Attachments) != 1 || !strings.Contains(bubble.State.Render(), "paste-1") {
		t.Fatalf("paste message did not create chip: %+v", bubble.State)
	}
}

func TestBubbleTeaStartupLayoutIsAnchored(t *testing.T) {
	model := New(commands.Environment{})
	model.Layout = Layout{Width: 120, Height: 32, Mode: "ready", ActiveModel: "lmstudio-amd", ActiveSkill: "none", MCPStatus: "manager wired", ContextStatus: "context=32768", PendingTool: "none", MemoryStatus: "fresh"}
	view := model.RenderStyled()
	plain := stripANSI(view)
	if strings.HasPrefix(plain, "\n\n\n\n\n\n\n") {
		t.Fatalf("startup layout has excessive top drift:\n%q", plain[:80])
	}
	for _, want := range []string{"RECOMP HAMR", "evidence-backed reconstruction", "Ask RecompHamr", "ready  lmstudio-amd  ready", "? help"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("startup layout missing %q:\n%s", want, plain)
		}
	}
	bubble := BubbleModel{State: model}
	teaView := bubble.View()
	if teaView.Cursor == nil || teaView.Cursor.Y != launcherTopPadding(model.Layout.Height)+3 {
		t.Fatalf("startup cursor = %+v", teaView.Cursor)
	}
	compact := stripANSI(model.RenderStyledWithLayout(Layout{Width: 72, Height: 24, Mode: "ready", ActiveModel: "local", ActiveSkill: "none", MCPStatus: "gated", ContextStatus: "ok", PendingTool: "none", MemoryStatus: "fresh"}))
	if !strings.Contains(compact, brandCompact) || strings.Contains(compact, brandWide) {
		t.Fatalf("compact styled startup brand mismatch:\n%s", compact)
	}
	for _, tc := range []struct {
		name   string
		layout Layout
	}{
		{"wide", model.Layout},
		{"80x24", Layout{Width: 80, Height: 24, Mode: "ready", ActiveModel: "local", PendingTool: "none", MemoryStatus: "fresh"}},
		{"60col", Layout{Width: 60, Height: 20, Mode: "ready", ActiveModel: "local", PendingTool: "none", MemoryStatus: "fresh"}},
	} {
		got := startupGolden(stripANSI(model.RenderStyledWithLayout(tc.layout)))
		want, err := os.ReadFile(filepath.Join("testdata", "startup_"+tc.name+".golden"))
		if err != nil {
			t.Fatal(err)
		}
		if got != strings.TrimSpace(string(want)) {
			t.Fatalf("%s startup golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", tc.name, got, want)
		}
	}
	missing := Layout{Width: 80, Height: 24, Mode: "ready", ActiveModel: "local", PendingTool: "none", MemoryStatus: "unsupported: missing"}
	if got := stripANSI(model.RenderStyledWithLayout(missing)); !strings.Contains(got, "Tip: /init-re creates project memory.") {
		t.Fatalf("missing-memory startup lacks actionable tip:\n%s", got)
	}
	if got := model.RenderWithLayout(missing); !strings.Contains(got, "Tip: /init-re creates project memory.") {
		t.Fatalf("plain missing-memory startup lacks actionable tip:\n%s", got)
	}
	working := Layout{Width: 80, Height: 24, Mode: "streaming", ActiveModel: "local", PendingTool: "agent", MemoryStatus: "fresh"}
	if got := startupStatus(working); got != "streaming  local  working" {
		t.Fatalf("working startup status = %q", got)
	}
	working.PendingTool = "none"
	if got := startupStatus(working); got != "streaming  local  working" {
		t.Fatalf("streaming startup status = %q", got)
	}
}

func startupGolden(view string) string {
	var lines []string
	for _, line := range strings.Split(view, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "│"))
		if line != "" {
			lines = append(lines, strings.Join(strings.Fields(line), " "))
		}
	}
	return strings.Join(lines, "\n")
}

func TestBubbleTeaRedesignedStyledChatStates(t *testing.T) {
	model := New(commands.Environment{})
	model.Transcript = []string{
		"user: inspect",
		"assistant: verified",
		"tool: read_file ok",
		"mcp ghidra connected",
		"blocked: denied",
		"unsupported: missing",
		"unverified: evidence",
	}
	model.Composer = "/"
	model.Status = "streaming"
	view := model.RenderStyledWithLayout(Layout{Width: 100, Height: 28, Mode: "run", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "wired", ContextStatus: "32k", PendingTool: "agent", MemoryStatus: "fresh"})
	for _, want := range []string{"COMMAND PALETTE", "assistant: verified", "tool: read_file ok", "blocked: denied", "unsupported: missing", "status: streaming"} {
		if !strings.Contains(view, want) {
			t.Fatalf("styled chat missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, "\x1b[") {
		t.Fatalf("styled chat missing ANSI:\n%q", view)
	}
	empty := transcriptBubble(New(commands.Environment{}), 60, 4)
	if !strings.Contains(empty, "No transcript yet") {
		t.Fatalf("empty transcript bubble = %q", empty)
	}
	if got := transcriptCard("user: "+strings.Repeat("x", 50), 20, true); lipgloss.Width(got) != 20 {
		t.Fatalf("compact transcriptCard width=%d text=%q", lipgloss.Width(got), got)
	}
	model.Composer = "/m"
	if small := paletteBubble(model, 36); !strings.Contains(small, "COMMAND PALETTE") {
		t.Fatalf("small palette = %q", small)
	}
	if styledDefault := model.RenderStyledWithLayout(Layout{}); !strings.Contains(styledDefault, "COMMAND PALETTE") {
		t.Fatalf("styled default layout missing palette:\n%s", styledDefault)
	}
	model.Composer = ""
	tiny := model.RenderStyledWithLayout(Layout{Width: 72, Height: 4, Mode: "run", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "wired", ContextStatus: "32k", PendingTool: "agent", MemoryStatus: "fresh"})
	if !strings.Contains(tiny, "needs a larger terminal") || !strings.Contains(tiny, "required 60x18") {
		t.Fatalf("tiny styled chat missing size diagnostic:\n%s", tiny)
	}
	if got := renderWidth(0); got != 110 {
		t.Fatalf("renderWidth zero = %d", got)
	}
}

func TestPhase48BubbleRuntimeInputAndTerminalFloor(t *testing.T) {
	if event, ok := bubbleEvent(keyText("x")); !ok || event.Text != "x" {
		t.Fatalf("bubbleEvent(text) = %#v, %v", event, ok)
	}
	bubble := NewBubble(commands.Environment{})
	updated, cmd := bubble.Update(tea.FocusMsg{})
	bubble = updated.(BubbleModel)
	if cmd == nil || !bubble.components.composer.Focused() {
		t.Fatalf("focus did not reach textarea: focused=%v cmd=%v", bubble.components.composer.Focused(), cmd)
	}
	updated, _ = bubble.Update(tea.BlurMsg{})
	bubble = updated.(BubbleModel)
	if bubble.components.composer.Focused() {
		t.Fatal("blur did not reach textarea")
	}
	updated, _ = bubble.Update(tea.FocusMsg{})
	bubble = updated.(BubbleModel)
	updated, _ = bubble.Update(keyText("ab"))
	bubble = updated.(BubbleModel)
	if bubble.State.Composer != "ab" || bubble.LastIntent.Kind != IntentNone {
		t.Fatalf("printable input failed: %+v", bubble)
	}
	updated, _ = bubble.Update(keyCode(tea.KeyBackspace))
	bubble = updated.(BubbleModel)
	if bubble.State.Composer != "a" {
		t.Fatalf("textarea backspace failed: %q", bubble.State.Composer)
	}
	updated, _ = bubble.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter, Mod: tea.ModShift}))
	bubble = updated.(BubbleModel)
	updated, _ = bubble.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Mod: tea.ModCtrl}))
	bubble = updated.(BubbleModel)
	if strings.Count(bubble.State.Composer, "\n") != 2 {
		t.Fatalf("multiline fallbacks failed: %q", bubble.State.Composer)
	}
	updated, _ = bubble.Update(keyCode(tea.KeyUp))
	bubble = updated.(BubbleModel)
	if bubble.State.Composer == "" {
		t.Fatal("multiline cursor key cleared textarea")
	}

	empty := NewBubble(commands.Environment{})
	updated, _ = empty.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelUp}))
	empty = updated.(BubbleModel)
	if empty.components.transcript.YOffset() != 0 {
		t.Fatal("empty transcript wheel changed viewport")
	}
	updated, _ = empty.Update(keyCode(tea.KeyPgUp))
	empty = updated.(BubbleModel)
	if empty.components.transcript.YOffset() != 0 {
		t.Fatal("empty transcript page changed viewport")
	}

	chat := NewBubble(commands.Environment{})
	for i := 0; i < 40; i++ {
		chat.State.Transcript = append(chat.State.Transcript, "assistant: line")
	}
	chat.components.syncFromState(chat.State)
	chat.components.transcript.GotoBottom()
	before := chat.components.transcript.YOffset()
	updated, _ = chat.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelUp}))
	chat = updated.(BubbleModel)
	if chat.components.transcript.YOffset() >= before {
		t.Fatalf("wheel did not scroll transcript: before=%d after=%d", before, chat.components.transcript.YOffset())
	}
	updated, _ = chat.Update(keyCode(tea.KeyPgDown))
	chat = updated.(BubbleModel)
	if chat.View().MouseMode != tea.MouseModeCellMotion {
		t.Fatal("chat with wheel support did not request mouse capture")
	}

	tooSmall := NewBubble(commands.Environment{})
	tooSmall.State.Layout = Layout{Width: 59, Height: 17}
	view := tooSmall.View()
	if view.Cursor != nil || !strings.Contains(view.Content, "current 59x17") || !terminalTooSmall(tooSmall.State.Layout) {
		t.Fatalf("terminal floor view invalid: %+v", view)
	}
	if terminalTooSmall(Layout{Width: MinimumWidth, Height: MinimumHeight}) {
		t.Fatal("minimum supported size was rejected")
	}
}

func TestPhase47ComponentsAndTypedIntents(t *testing.T) {
	components := newBubbleComponents()
	if len(components.keys.ShortHelp()) != 3 || len(components.keys.FullHelp()) != 2 {
		t.Fatalf("key help is incomplete: short=%v full=%v", components.keys.ShortHelp(), components.keys.FullHelp())
	}
	state := New(commands.Environment{})
	state.Composer = "unicode 界"
	components.syncFromState(state)
	if components.composer.Value() != state.Composer || components.composer.Width() <= 0 || components.help.Width() != state.Layout.Width {
		t.Fatalf("startup component sync failed: composer=%q width=%d help=%d", components.composer.Value(), components.composer.Width(), components.help.Width())
	}
	state.Transcript = []string{"assistant: verified"}
	state.Status = "ready"
	state.Layout.Height = 7
	components.syncFromState(state)
	if !strings.Contains(components.transcript.GetContent(), "assistant") || !strings.Contains(components.transcript.GetContent(), "status") || components.transcript.Height() != 3 {
		t.Fatalf("chat component sync failed: %q", components.transcript.GetContent())
	}
	var zero bubbleComponents
	zero.syncFromState(state)
	if !zero.initialized {
		t.Fatal("zero component set was not initialized lazily")
	}
	cases := []struct {
		action Action
		kind   IntentKind
	}{
		{ActionNone, IntentNone},
		{ActionSubmit, IntentSubmit},
		{ActionCancel, IntentCancel},
		{ActionQuit, IntentQuit},
	}
	for _, tc := range cases {
		intent := intentFromAction(tc.action, "payload")
		if intent.Kind != tc.kind || (tc.action == ActionSubmit && intent.Value != "payload") {
			t.Fatalf("intentFromAction(%q) = %+v", tc.action, intent)
		}
	}
}

func TestPhase42OverlayModalsAndSelection(t *testing.T) {
	customDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(customDir, "custom.md"), []byte("# Custom\nUse this skill when custom work is needed.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	model := New(commands.Environment{CustomSkillsDir: customDir, ActiveSkills: []string{"custom"}})
	for _, tc := range []struct {
		composer string
		title    string
		want     string
	}{
		{"/models", "MODEL PICKER", "config is not loaded"},
		{"/skills", "SKILL PICKER", "custom"},
		{"/skill custom", "SKILL PICKER", "custom"},
		{"/mcp", "MCP CONTROLS", "ghidra"},
		{"/help", "HELP", "Ctrl+D"},
	} {
		model.Composer = tc.composer
		plain := stripANSI(model.RenderStyledWithLayout(Layout{Width: 100, Height: 28, Mode: "ready", ActiveModel: "local", ActiveSkill: "custom", MCPStatus: "gated", ContextStatus: "32k", MemoryStatus: "fresh", PendingTool: "none"}))
		if !strings.Contains(plain, tc.title) || !strings.Contains(plain, tc.want) {
			t.Fatalf("overlay %s missing title/want:\n%s", tc.composer, plain)
		}
		if title := overlayTitle(model.overlayKind()); title != tc.title {
			t.Fatalf("overlayTitle(%s) = %q", tc.composer, title)
		}
	}
	if rows := model.overlayRows(); len(rows) == 0 {
		t.Fatal("help overlay rows missing")
	}
	model.Composer = "/"
	model, _ = model.Update(Event{Key: KeyDown})
	model, _ = model.Update(Event{Key: KeyTab})
	if model.Composer != "/models " || !strings.Contains(model.Status, "/models") {
		t.Fatalf("selected tab completion composer=%q status=%q", model.Composer, model.Status)
	}
	model.PaletteIndex = 99
	model.Composer = "/m"
	model = model.completeComposer()
	if model.Composer != "/models " {
		t.Fatalf("out-of-range tab completion composer=%q", model.Composer)
	}
	model.Composer = "/skills"
	model.PaletteIndex = 0
	model, _ = model.Update(Event{Key: KeyDown})
	if model.PaletteIndex != 1 {
		t.Fatalf("modal down index=%d", model.PaletteIndex)
	}
	model.PaletteIndex = len(model.overlayRows()) - 1
	model, _ = model.Update(Event{Key: KeyDown})
	if model.PaletteIndex != len(model.overlayRows())-1 {
		t.Fatalf("modal down clamp index=%d", model.PaletteIndex)
	}
	model.PaletteIndex = 1
	model, _ = model.Update(Event{Key: KeyUp})
	model, _ = model.Update(Event{Key: KeyUp})
	if model.PaletteIndex != 0 {
		t.Fatalf("modal up clamp index=%d", model.PaletteIndex)
	}
	model.Composer = "/zz"
	if rows := model.PaletteRows(); len(rows) != 0 {
		t.Fatalf("unknown command palette rows=%v", rows)
	}
	model.Composer = "/models"
	if rows := model.PaletteRows(); rows != nil {
		t.Fatalf("model modal should not return command palette rows=%v", rows)
	}
	if empty := overlayBubble(Model{}, "commands", nil, 0, 60); !strings.Contains(empty, "unverified: no matches") {
		t.Fatalf("empty overlay = %q", empty)
	}
	paletteModel := New(commands.Environment{})
	paletteModel.Composer = "/"
	paletteModel.PaletteIndex = 10
	if got := stripANSI(overlayBubble(paletteModel, "commands", paletteModel.PaletteRows(), 10, 120)); !strings.Contains(got, "usage:") {
		t.Fatalf("scrolled command overlay lacks selected detail: %q", got)
	}
	paletteModel.PaletteIndex = 99
	if got := paletteModel.paletteDetail(); !strings.Contains(got, "usage:") {
		t.Fatalf("out-of-range palette detail = %q", got)
	}
	paletteModel.Composer = "/zz"
	if got := paletteModel.paletteDetail(); got != "" {
		t.Fatalf("no-match palette detail = %q", got)
	}
	_ = overlayBubble(Model{}, "models", []string{" row"}, -1, 70)
	_ = overlayBubble(Model{}, "models", []string{" row"}, 0, 100)
	_ = overlayBubble(Model{}, "models", []string{"! blocked unavailable"}, -1, 80)
	bubble := NewBubble(commands.Environment{})
	bubble.State.Composer = "/"
	updated, _ := bubble.Update(keyText("j"))
	bubble = updated.(BubbleModel)
	if bubble.State.PaletteIndex != 1 {
		t.Fatalf("j palette index = %d", bubble.State.PaletteIndex)
	}
	updated, _ = bubble.Update(keyText("k"))
	if updated.(BubbleModel).State.PaletteIndex != 0 {
		t.Fatalf("k palette index = %d", updated.(BubbleModel).State.PaletteIndex)
	}
	for _, tc := range []struct {
		composer string
		kind     IntentKind
	}{
		{"/skills", IntentSkill},
		{"/mcp", IntentMCP},
	} {
		picker := NewBubble(commands.Environment{})
		picker.State.Composer = tc.composer
		updated, _ := picker.Update(keyCode(tea.KeyEnter))
		got := updated.(BubbleModel)
		if got.LastIntent.Kind != tc.kind || got.LastIntent.Value == "" || got.State.Composer != "" {
			t.Fatalf("%s picker intent = %+v composer=%q", tc.composer, got.LastIntent, got.State.Composer)
		}
	}
	blocked := New(commands.Environment{})
	blocked.Composer = "/models"
	if _, ok := blocked.selectedOverlayIntent(); ok {
		t.Fatal("blocked model row emitted intent")
	}
	help := New(commands.Environment{})
	help.Composer = "/help"
	if _, ok := help.selectedOverlayIntent(); ok {
		t.Fatal("help overlay emitted selection intent")
	}
	if _, ok := selectedRowName(nil, 0); ok {
		t.Fatal("empty rows selected a value")
	}
	if name, ok := selectedRowName([]string{" model ready"}, 99); !ok || name != "model" {
		t.Fatalf("out-of-range selected row = %q, %v", name, ok)
	}
	if _, ok := selectedRowName([]string{""}, 0); ok {
		t.Fatal("blank row selected a value")
	}
	if got := intentKindForOverlay("models"); got != IntentModel {
		t.Fatalf("model intent kind = %q", got)
	}
	plain := overlayPlain("mcp", []string{"* ghidra connected tools 1"}, Layout{Width: 80})
	if !strings.Contains(plain, "Mcp Controls") || !strings.Contains(plain, "ghidra") {
		t.Fatalf("plain overlay = %q", plain)
	}
}

func TestPhase42BlockedOverlayRows(t *testing.T) {
	model := New(commands.Environment{CustomSkillsDir: string(rune(0))})
	for _, composer := range []string{"/skills", "/models"} {
		model.Composer = composer
		view := stripANSI(model.RenderStyledWithLayout(Layout{Width: 96, Height: 24, Mode: "ready", ActiveModel: "local", ActiveSkill: "none", MCPStatus: "error", ContextStatus: "32k", MemoryStatus: "fresh", PendingTool: "none"}))
		if !strings.Contains(view, "blocked") && !strings.Contains(view, "config is not loaded") {
			t.Fatalf("blocked overlay %s missing blocked evidence:\n%s", composer, view)
		}
	}
}

func TestPhase43TranscriptRuntimeStatesAndRedaction(t *testing.T) {
	model := New(commands.Environment{})
	model.DebugSecrets = []string{"secret-token"}
	model.Transcript = []string{
		"user: map function with secret-token",
		"assistant: working from evidence",
		"tool: powershell completed",
		"mcp: ghidra.decompile returned",
		"verification: local checks passed",
		"blocked: permission denied",
		"unsupported: remote metrics unavailable",
		"unverified: context header missing",
	}
	model.Status = "streaming verification without metrics secret-token"
	plain := model.RenderWithLayout(Layout{Width: 100, Height: 28, Mode: "streaming", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "wired", ContextStatus: "32k", MemoryStatus: "fresh", PendingTool: "powershell"})
	for _, want := range []string{"user        user: map function", "assistant   assistant: working", "tool        tool: powershell", "mcp         mcp: ghidra", "verification verification:", "blocked     blocked:", "unsupported unsupported:", "unverified  unverified:", "status      status: streaming verification"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("runtime transcript missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "secret-token") || !strings.Contains(plain, "[REDACTED]") {
		t.Fatalf("transcript redaction failed:\n%s", plain)
	}
	for _, fake := range []string{"$0.", " tokens", "ms", "Thought:"} {
		if strings.Contains(plain, fake) {
			t.Fatalf("transcript contains fake metric marker %q:\n%s", fake, plain)
		}
	}
	styled := stripANSI(model.RenderStyledWithLayout(Layout{Width: 100, Height: 28, Mode: "verifying", ActiveModel: "local", ActiveSkill: "core-re", MCPStatus: "wired", ContextStatus: "32k", MemoryStatus: "fresh", PendingTool: "doctor"}))
	if strings.Contains(styled, "secret-token") || !strings.Contains(styled, "verification: local checks passed") || !strings.Contains(styled, "status: streaming verification") {
		t.Fatalf("styled runtime state render failed:\n%s", styled)
	}
}

func TestPhase51TranscriptFollowBoundingAndRuntimeFeedback(t *testing.T) {
	model := New(commands.Environment{})
	model.DebugSecrets = []string{"phase51-secret"}
	for i := 0; i < 20; i++ {
		model = model.AppendRuntimeTranscript(fmt.Sprintf("assistant: evidence %02d", i))
	}
	model = model.scrollTranscript(6)
	before := model.TranscriptOffset
	model = model.AppendRuntimeTranscript("tool: powershell\n" + strings.Repeat("long phase51-secret line\n", 14))
	if !model.NewOutput || model.TranscriptOffset != before+1 {
		t.Fatalf("paused append offset=%d new=%v", model.TranscriptOffset, model.NewOutput)
	}
	view := stripANSI(model.RenderStyledWithLayout(Layout{Width: 80, Height: 24, Mode: "run", ActiveModel: "local", PendingTool: "powershell"}))
	for _, want := range []string{"new output", "PgDn to follow"} {
		if !strings.Contains(view, want) {
			t.Fatalf("scrolled transcript missing %q:\n%s", want, view)
		}
	}
	model = model.scrollTranscript(-999)
	if model.TranscriptOffset != 0 || model.NewOutput {
		t.Fatalf("follow restore offset=%d new=%v", model.TranscriptOffset, model.NewOutput)
	}
	bounded := boundedTranscriptBlock("tool: powershell\n"+strings.Repeat("output line\n", 14), 30)
	if !strings.Contains(bounded, "output truncated") || lipgloss.Width(strings.Split(bounded, "\n")[0]) > 22 {
		t.Fatalf("bounded tool output = %q", bounded)
	}
	if got := boundedTranscriptBlock("note", 10); got != "note        note" {
		t.Fatalf("narrow bounded note = %q", got)
	}
	if got := transcriptBlock("warning: verify symbols"); !strings.HasPrefix(got, "warning") {
		t.Fatalf("warning block = %q", got)
	}
	if got := stripANSI(styleTranscriptLine("warning: verify symbols", 60)); !strings.Contains(got, "warning") {
		t.Fatalf("warning style = %q", got)
	}
	if got := visibleTranscriptWindow([]string{"a", "b"}, 5, 99); len(got) != 0 {
		t.Fatalf("overscrolled window = %v", got)
	}
	if got := visibleTranscriptWindow([]string{"a", "b", "c"}, 2, 0); strings.Join(got, "") != "bc" {
		t.Fatalf("follow window = %v", got)
	}
	empty := New(commands.Environment{})
	empty = empty.AppendRuntimeTranscript()
	empty = empty.scrollTranscript(5)
	if empty.TranscriptOffset != 0 {
		t.Fatalf("empty scroll offset = %d", empty.TranscriptOffset)
	}
	bubble := NewBubble(commands.Environment{})
	bubble.State = model
	bubble.State.TranscriptOffset = 4
	updated, _ := bubble.Update(keyCode(tea.KeyPgUp))
	bubble = updated.(BubbleModel)
	if bubble.State.TranscriptOffset <= 4 {
		t.Fatalf("page up did not move from follow: %d", bubble.State.TranscriptOffset)
	}
	bubble.State.TranscriptOffset = 4
	updated, _ = bubble.Update(keyCode(tea.KeyPgDown))
	bubble = updated.(BubbleModel)
	if bubble.State.TranscriptOffset >= 4 {
		t.Fatalf("page down did not approach follow: %d", bubble.State.TranscriptOffset)
	}
	updated, _ = bubble.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown}))
	if updated.(BubbleModel).State.TranscriptOffset != 0 {
		t.Fatalf("wheel down did not restore follow: %d", updated.(BubbleModel).State.TranscriptOffset)
	}
	bubble.State.Layout.Height = 2
	bubble.State.TranscriptOffset = 1
	updated, _ = bubble.Update(keyCode(tea.KeyPgDown))
	if updated.(BubbleModel).State.TranscriptOffset != 0 {
		t.Fatalf("short page delta did not clamp: %d", updated.(BubbleModel).State.TranscriptOffset)
	}
	bubble.State.TranscriptOffset = 1
	updated, _ = bubble.Update(keyCode(tea.KeyPgUp))
	if updated.(BubbleModel).State.TranscriptOffset != 2 {
		t.Fatalf("short page-up delta = %d", updated.(BubbleModel).State.TranscriptOffset)
	}
	if got := transcriptBubble(model, 60, 2); got == "" {
		t.Fatal("short transcript bubble is empty")
	}
	goldenModel := New(commands.Environment{})
	goldenModel.Transcript = []string{
		"user: map function", "assistant: evidence found", "tool: powershell verified",
		"mcp: ghidra connected", "verification: hashes match", "warning: inspect relocation",
		"blocked: permission denied", "unsupported: provider metric", "paste: paste-1 (20 bytes)",
	}
	golden := runtimeGolden(stripANSI(goldenModel.RenderStyledWithLayout(Layout{Width: 80, Height: 32, Mode: "ready", ActiveModel: "local", PendingTool: "none"})))
	want, err := os.ReadFile(filepath.Join("testdata", "runtime_states.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if golden != strings.TrimSpace(string(want)) {
		t.Fatalf("runtime golden mismatch:\n--- got ---\n%s\n--- want ---\n%s", golden, want)
	}
}

func TestPhase52ResponsiveThemeAccessibilityAndUnicode(t *testing.T) {
	profiles := []struct {
		name    string
		profile colorprofile.Profile
	}{
		{"no-color", colorprofile.ASCII},
		{"ansi16", colorprofile.ANSI},
		{"ansi256", colorprofile.ANSI256},
		{"truecolor", colorprofile.TrueColor},
	}
	var profileEvidence []string
	model := New(commands.Environment{})
	model.Composer = "解析 e\u0301vidence"
	model.Transcript = []string{"user: 解析 e\u0301vidence", "assistant: verified symbols", "warning: inspect relocation"}
	for _, size := range []Layout{{Width: 120, Height: 32}, {Width: 96, Height: 28}, {Width: 80, Height: 24}, {Width: 60, Height: 20}, {Width: 60, Height: 18}} {
		for _, profile := range profiles {
			size.ColorProfile = profile.profile
			size.Mode, size.ActiveModel, size.PendingTool = "ready", strings.Repeat("local-model-", 8), "none"
			view := model.RenderStyledWithLayout(size)
			plain := stripANSI(view)
			for lineNumber, line := range strings.Split(plain, "\n") {
				if width := lipgloss.Width(line); width > size.Width {
					t.Fatalf("%s %dx%d line %d width=%d: %q", profile.name, size.Width, size.Height, lineNumber, width, line)
				}
			}
			if !strings.Contains(plain, "解析") || !strings.Contains(plain, "warning") {
				t.Fatalf("%s %dx%d lost Unicode/state label:\n%s", profile.name, size.Width, size.Height, plain)
			}
			switch profile.profile {
			case colorprofile.ASCII:
				if strings.Contains(view, "[38;") || strings.Contains(view, "[48;") {
					t.Fatalf("no-color render contains color sequence: %q", view)
				}
				if size.Width == 80 {
					profileEvidence = append(profileEvidence, "no-color: labels+rail+reverse; color=none")
				}
			case colorprofile.ANSI:
				if strings.Contains(view, "38;5") || strings.Contains(view, "38;2") || strings.Contains(view, "48;5") || strings.Contains(view, "48;2") {
					t.Fatalf("ANSI16 render contains extended color: %q", view)
				}
				if size.Width == 80 {
					profileEvidence = append(profileEvidence, "ansi16: basic-color only")
				}
			case colorprofile.ANSI256:
				if !strings.Contains(view, "38;5") {
					t.Fatalf("ANSI256 render lacks indexed color: %q", view)
				}
				if size.Width == 80 {
					profileEvidence = append(profileEvidence, "ansi256: indexed-color")
				}
			case colorprofile.TrueColor:
				if !strings.Contains(view, "38;2") {
					t.Fatalf("truecolor render lacks RGB color: %q", view)
				}
				if size.Width == 80 {
					profileEvidence = append(profileEvidence, "truecolor: rgb-color")
				}
			}
		}
	}
	wantProfiles, err := os.ReadFile(filepath.Join("testdata", "color_profiles.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(profileEvidence, "\n"); got != strings.TrimSpace(string(wantProfiles)) {
		t.Fatalf("color profile golden mismatch:\n%s", got)
	}
	t.Setenv("NO_COLOR", "1")
	if got := DefaultLayout().ColorProfile; got != colorprofile.ASCII {
		t.Fatalf("NO_COLOR profile = %s", got)
	}
	bubble := NewBubble(commands.Environment{})
	updated, _ := bubble.Update(tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	bubble = updated.(BubbleModel)
	if bubble.State.Layout.ColorProfile != colorprofile.TrueColor {
		t.Fatalf("color profile message = %s", bubble.State.Layout.ColorProfile)
	}
	bubble.State.Layout.Width = 60
	bubble.State.Layout.Height = 20
	bubble.State.Composer = strings.Repeat("界", 80)
	bubble.State.Transcript = []string{"user: long composer"}
	view := bubble.View()
	if view.Cursor == nil || view.Cursor.X < 0 || view.Cursor.X >= 60 || view.Cursor.Y < 0 || view.Cursor.Y >= 20 {
		t.Fatalf("bounded Unicode cursor = %+v", view.Cursor)
	}
	tooSmall := model.RenderStyledWithLayout(Layout{Width: 59, Height: 17, ColorProfile: colorprofile.ASCII})
	if strings.Contains(tooSmall, "[38;") || !strings.Contains(tooSmall, "required 60x18") {
		t.Fatalf("no-color too-small state = %q", tooSmall)
	}
	startup := New(commands.Environment{})
	startup.Composer = strings.Repeat("界", 80)
	startupScreen := startup.startupScreen(Layout{Width: 60, Height: 20, ColorProfile: colorprofile.ASCII})
	if startupScreen.cursorX < 0 || startupScreen.cursorX >= 60 {
		t.Fatalf("bounded startup cursor = %d", startupScreen.cursorX)
	}
}

func runtimeGolden(view string) string {
	var out []string
	for _, line := range strings.Split(view, "\n") {
		line = strings.Join(strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "│"))), " ")
		for _, label := range []string{"user ", "assistant ", "tool ", "mcp ", "verification ", "warning ", "blocked ", "unsupported ", "attachment "} {
			if strings.HasPrefix(line, label) {
				out = append(out, line)
			}
		}
	}
	return strings.Join(out, "\n")
}

func keyText(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}

func keyCode(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func stripANSI(text string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		if ch == 0x1b {
			inEscape = true
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

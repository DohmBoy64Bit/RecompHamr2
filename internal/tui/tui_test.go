package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

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
	for _, want := range []string{"RECOMP HAMR", "RE . decomp . recomp", "Ask RecompHamr", "/ commands", "Tip:"} {
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
	for _, want := range []string{"RecompHamr", "Ask RecompHamr", "Build * local", "/ commands", "Tip:"} {
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
		"blocked: denied",
		"unsupported: later",
		"unverified: missing evidence",
		"status: cancelled",
		"paste: paste-1 (12 bytes)",
		"plain note",
	}
	wantLabels := []string{"user", "assistant", "tool", "mcp", "mcp", "blocked", "unsupported", "unverified", "status", "attachment", "note"}
	for i, line := range lines {
		got := transcriptBlock(line)
		if !strings.HasPrefix(got, wantLabels[i]) || !strings.Contains(got, line) {
			t.Fatalf("transcriptBlock(%q) = %q", line, got)
		}
	}
	model := New(commands.Environment{})
	model.Transcript = lines
	view := model.RenderWithLayout(Layout{Width: 80, Mode: "ready", ActiveModel: "local", ActiveSkill: "none", MCPStatus: "gated", ContextStatus: "ok", PendingTool: "none", MemoryStatus: "fresh"})
	for _, want := range []string{"blocked     blocked: denied", "attachment  paste: paste-1", "note        plain note"} {
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
	if len(rows) == 0 || !strings.Contains(rows[0], "> /models") || !strings.Contains(rows[0], "usage: /models [name]") {
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
	if !view.AltScreen || view.MouseMode != tea.MouseModeCellMotion || !view.ReportFocus || view.WindowTitle != "RecompHamr" {
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
	for _, want := range []string{"RECOMP HAMR", "Ask RecompHamr", "Build * lmstudio-amd * ready"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("startup layout missing %q:\n%s", want, plain)
		}
	}
	bubble := BubbleModel{State: model}
	teaView := bubble.View()
	if teaView.Cursor == nil || teaView.Cursor.Y != launcherTopPadding(model.Layout.Height)+4 {
		t.Fatalf("startup cursor = %+v", teaView.Cursor)
	}
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
	if got := transcriptCard("user: "+strings.Repeat("x", 50), 20, true); len(got) != 20 {
		t.Fatalf("compact transcriptCard length=%d text=%q", len(got), got)
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
	if !strings.Contains(tiny, "assistant: verified") || !strings.Contains(tiny, "Build * local") {
		t.Fatalf("tiny styled chat missing content:\n%s", tiny)
	}
	if got := renderWidth(0); got != 110 {
		t.Fatalf("renderWidth zero = %d", got)
	}
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

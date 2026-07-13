package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"recomphamr2/internal/commands"
)

func TestNewAndStartupRendering(t *testing.T) {
	snapshot := testSnapshot()
	model := New(snapshot)
	if model.Snapshot().ActiveModel != "local" || model.ComposerValue() != "" || len(model.Entries()) != 0 {
		t.Fatalf("new model = %#v value=%q entries=%v", model.Snapshot(), model.ComposerValue(), model.Entries())
	}
	if model.Init() == nil {
		t.Fatal("Init did not return cursor blink")
	}
	view := model.View()
	plain := ansi.Strip(view.Content)
	for _, want := range []string{"RECOMP", "HAMR", "evidence-backed reconstruction", "Ask RecompHamr", "local  ready"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("startup missing %q:\n%s", want, plain)
		}
	}
	if view.Cursor == nil || !view.AltScreen || !view.ReportFocus || view.WindowTitle != "RecompHamr" {
		t.Fatalf("startup view fields = %#v", view)
	}
}

func TestBlurredStartupAndChatHideCursorWithoutPanicking(t *testing.T) {
	model := New(testSnapshot())
	model = updateModel(t, model, tea.BlurMsg{})
	startup := model.View()
	if startup.Cursor != nil || !strings.Contains(ansi.Strip(startup.Content), "RECOMP") {
		t.Fatalf("blurred startup cursor=%#v content=%q", startup.Cursor, ansi.Strip(startup.Content))
	}

	model = updateModel(t, model, TranscriptMsg{Entries: []TranscriptEntry{{Kind: TranscriptAssistant, Text: "ready"}}})
	chat := model.View()
	if chat.Cursor != nil || !strings.Contains(ansi.Strip(chat.Content), "assistant") {
		t.Fatalf("blurred chat cursor=%#v content=%q", chat.Cursor, ansi.Strip(chat.Content))
	}
}

func TestUpdateMessagesAndAuthoritativeComposer(t *testing.T) {
	model := New(testSnapshot())
	model = updateModel(t, model, tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updateModel(t, model, tea.ColorProfileMsg{Profile: colorprofile.TrueColor})
	model = updateModel(t, model, ColorProfileMsg{Profile: colorprofile.ASCII})
	snapshot := testSnapshot()
	snapshot.Status = "verified"
	model = updateModel(t, model, SnapshotMsg{Snapshot: snapshot})
	model = updateModel(t, model, tea.FocusMsg{})
	model = updateModel(t, model, tea.BlurMsg{})
	model = updateModel(t, model, tea.FocusMsg{})
	model = updateModel(t, model, tea.PasteMsg{Content: "paste"})
	if model.ComposerValue() != "paste" {
		t.Fatalf("paste value = %q", model.ComposerValue())
	}
	model = updateModel(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	if model.ComposerValue() != "past" {
		t.Fatalf("backspace value = %q", model.ComposerValue())
	}
	model = updateModel(t, model, struct{}{})
	model = updateModel(t, model, tea.MouseWheelMsg{})
}

func TestSubmitCommandPromptHistoryAndSlashSafety(t *testing.T) {
	model := New(testSnapshot())
	model, cmd := updateKey(model, keyCode(tea.KeyEnter))
	if cmd != nil {
		t.Fatal("empty submit emitted intent")
	}
	model, _ = updateKey(model, keyText("/"))
	if model.overlay != overlayCommands || model.ComposerValue() != "/" {
		t.Fatalf("slash state overlay=%q value=%q", model.overlay, model.ComposerValue())
	}
	model, _ = updateKey(model, keyCode(tea.KeyBackspace))
	if model.overlay != overlayNone || model.ComposerValue() != "" {
		t.Fatalf("bare slash backspace overlay=%q value=%q", model.overlay, model.ComposerValue())
	}
	model, _ = updateKey(model, keyText("/"))
	model.closeOverlay()
	model, cmd = updateKey(model, keyCode(tea.KeyEnter))
	if cmd != nil {
		t.Fatal("bare slash emitted command")
	}
	model.composer.SetValue("/doctor")
	model, cmd = updateKey(model, keyCode(tea.KeyEnter))
	assertIntent(t, cmd, IntentCommand, "/doctor")
	model.composer.SetValue("hello")
	model, cmd = updateKey(model, keyCode(tea.KeyEnter))
	assertIntent(t, cmd, IntentSubmit, "hello")
	model.recall(-1)
	if model.ComposerValue() != "hello" {
		t.Fatalf("history previous = %q", model.ComposerValue())
	}
	model.recall(-10)
	model.recall(1)
	model.recall(10)
	if model.ComposerValue() != "" {
		t.Fatalf("history end = %q", model.ComposerValue())
	}
}

func TestTextareaOwnsMultilinePasteAndCursor(t *testing.T) {
	model := New(testSnapshot())
	if model.ComposerValue() != "" || !strings.Contains(ansi.Strip(model.View().Content), "Ask RecompHamr") {
		t.Fatal("placeholder became textarea value")
	}
	model = updateModel(t, model, tea.PasteMsg{Content: "alpha βeta"})
	model, _ = updateKey(model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter, Mod: tea.ModShift}))
	model, _ = updateKey(model, keyText("second"))
	if !strings.Contains(model.ComposerValue(), "\n") || !strings.Contains(model.ComposerValue(), "second") {
		t.Fatalf("multiline value=%q", model.ComposerValue())
	}
	before := model.composer.Cursor().Position
	model, _ = updateKey(model, keyCode(tea.KeyLeft))
	after := model.composer.Cursor().Position
	if before == after {
		t.Fatalf("cursor did not move: before=%v after=%v", before, after)
	}
	model, _ = updateKey(model, keyCtrl('j'))
	if strings.Count(model.ComposerValue(), "\n") < 2 {
		t.Fatalf("ctrl+j did not insert newline: %q", model.ComposerValue())
	}
}

func TestOverlaySelectionCompletionAndDismissal(t *testing.T) {
	model := New(testSnapshot())
	model.openOverlay(overlayCommands)
	model.picker.Select(0)
	updated, cmd := model.acceptSelection()
	model = updated.(Model)
	assertIntent(t, cmd, IntentCommand, "/clear")

	for _, tc := range []struct {
		name string
		kind overlayKind
	}{
		{"/models", overlayModels},
		{"/skills", overlaySkills},
		{"/skill", overlaySkills},
		{"/mcp", overlayMCP},
		{"/help", overlayHelp},
	} {
		model.openOverlay(overlayCommands)
		selectItem(t, &model, tc.name)
		updated, cmd = model.acceptSelection()
		model = updated.(Model)
		if cmd != nil || model.overlay != tc.kind {
			t.Fatalf("%s overlay=%q cmd=%v", tc.name, model.overlay, cmd)
		}
	}

	model.openOverlay(overlayModels)
	_ = model.picker.SetItems([]list.Item{pickerItem{name: "local", kind: IntentModel}})
	model.picker.Select(0)
	updated, cmd = model.acceptSelection()
	model = updated.(Model)
	assertIntent(t, cmd, IntentModel, "local")

	model.openOverlay(overlayCommands)
	model.picker.Select(1)
	updated, cmd = model.completeSelection()
	model = updated.(Model)
	if cmd != nil || !strings.HasPrefix(model.ComposerValue(), "/models") || model.overlay != overlayNone {
		t.Fatalf("completion value=%q overlay=%q", model.ComposerValue(), model.overlay)
	}

	blocked := pickerItem{name: "blocked", description: "no", blocked: true}
	_ = model.picker.SetItems([]list.Item{blocked})
	model.overlay = overlayModels
	if _, cmd = model.acceptSelection(); cmd != nil {
		t.Fatal("blocked selection emitted intent")
	}
	if _, cmd = model.completeSelection(); cmd != nil {
		t.Fatal("blocked completion emitted intent")
	}
	_ = model.picker.SetItems(nil)
	if _, cmd = model.acceptSelection(); cmd != nil {
		t.Fatal("empty selection emitted intent")
	}
	if _, cmd = model.completeSelection(); cmd != nil {
		t.Fatal("empty completion emitted intent")
	}

	model.openOverlay(overlayHelp)
	model, _ = updateKey(model, keyCode(tea.KeyEsc))
	if model.overlay != overlayNone {
		t.Fatalf("escape overlay=%q", model.overlay)
	}
}

func TestPickerKindsAndKeyboardRouting(t *testing.T) {
	model := New(testSnapshot())
	item := pickerItem{name: "name", description: "description"}
	if item.Title() != "name" || item.Description() != "description" || item.FilterValue() == "" {
		t.Fatalf("picker item=%#v", item)
	}
	if len(model.keys.ShortHelp()) != 3 || len(model.keys.FullHelp()) != 2 {
		t.Fatal("help bindings incomplete")
	}
	for _, kind := range []overlayKind{overlayModels, overlaySkills, overlayMCP, overlayHelp, overlayNone} {
		items := pickerItems(kind, model.snapshot)
		if kind != overlayNone && len(items) == 0 {
			t.Fatalf("picker %q empty", kind)
		}
		_ = overlayTitle(kind)
	}
	model.openOverlay(overlayCommands)
	model, _ = updateKey(model, keyText("m"))
	model, _ = updateKey(model, keyCode(tea.KeyTab))
	model.openOverlay(overlayCommands)
	model, _ = updateKey(model, keyCode(tea.KeyDown))
	model, _ = updateKey(model, keyCode(tea.KeyEnter))
	model = New(testSnapshot())
	model, _ = updateKey(model, keyCode(tea.KeyPgUp))
	model, _ = updateKey(model, keyCode(tea.KeyUp))
	model, _ = updateKey(model, keyCode(tea.KeyDown))
	model.composer.SetValue("x")
	model, _ = updateKey(model, keyText("?"))
	model.composer.SetValue("a\nb")
	model, _ = updateKey(model, keyCode(tea.KeyUp))
	model, _ = updateKey(model, keyCode(tea.KeyDown))
	model.recall(-1)
}

type nonPickerItem string

// FilterValue satisfies list.Item for the delegate rejection test.
func (item nonPickerItem) FilterValue() string { return string(item) }

func TestPickerDelegateSemanticRows(t *testing.T) {
	model := New(testSnapshot())
	delegate := pickerDelegate{profile: colorprofile.ASCII}
	var output strings.Builder
	delegate.Render(&output, model.picker, 0, pickerItem{name: "server", description: "offline", blocked: true})
	if !strings.Contains(output.String(), "> server") || !strings.Contains(output.String(), "[blocked]") {
		t.Fatalf("blocked delegate row=%q", output.String())
	}
	before := output.Len()
	delegate.Render(&output, model.picker, 0, nonPickerItem("ignored"))
	if output.Len() != before {
		t.Fatalf("non-picker item rendered: %q", output.String())
	}
}

func TestListOwnsFilteringNavigationAndEmptyState(t *testing.T) {
	model := New(testSnapshot())
	model.openOverlay(overlayCommands)
	for _, character := range "doctor" {
		updated, cmd := updateKey(model, keyText(string(character)))
		model = updated
		if cmd == nil {
			t.Fatalf("filter character %q returned no command", character)
		}
	}
	model.picker.SetFilterText(model.picker.FilterValue())
	filterCmd := model.picker.SetItems(commandItems())
	if filterCmd == nil {
		t.Fatal("filtered item replacement returned no filter command")
	}
	model = updateModel(t, model, filterCmd())
	visible := model.picker.VisibleItems()
	if len(visible) != 1 || visible[0].(pickerItem).name != "/doctor" {
		t.Fatalf("filtered items=%#v filter=%q", visible, model.picker.FilterValue())
	}

	model.overlay = overlayModels
	model.picker.SetFilterState(list.Unfiltered)
	_ = model.picker.SetItems([]list.Item{
		pickerItem{name: "first", kind: IntentModel},
		pickerItem{name: "second", kind: IntentModel},
	})
	model.picker.Select(0)
	model, _ = updateKey(model, keyText("j"))
	if model.picker.SelectedItem().(pickerItem).name != "second" {
		t.Fatalf("j navigation selected %#v", model.picker.SelectedItem())
	}
	model, _ = updateKey(model, keyText("k"))
	if model.picker.SelectedItem().(pickerItem).name != "first" {
		t.Fatalf("k navigation selected %#v", model.picker.SelectedItem())
	}

	_ = model.picker.SetItems(nil)
	if !strings.Contains(strings.ToLower(ansi.Strip(model.picker.View())), "no items") {
		t.Fatalf("empty picker view:\n%s", ansi.Strip(model.picker.View()))
	}
}

func TestArgumentAndHelpSelectionsPopulateComposer(t *testing.T) {
	model := New(testSnapshot())
	for _, command := range []string{"/skill-audit", "/skill-new"} {
		model.composer.Reset()
		model.openOverlay(overlayCommands)
		selectItem(t, &model, command)
		updated, cmd := model.acceptSelection()
		model = updated.(Model)
		if cmd != nil || model.ComposerValue() != command+" " || model.overlay != overlayNone {
			t.Fatalf("argument command %s value=%q overlay=%q cmd=%v", command, model.ComposerValue(), model.overlay, cmd)
		}
	}
	if !commandNeedsInput("/skill-audit") || commandNeedsInput("/doctor") {
		t.Fatal("command input classification failed")
	}

	model.composer.Reset()
	model.openOverlay(overlayHelp)
	selectItem(t, &model, "/doctor")
	updated, cmd := model.acceptSelection()
	model = updated.(Model)
	if cmd != nil || model.ComposerValue() != "/doctor " || model.overlay != overlayNone {
		t.Fatalf("help selection value=%q overlay=%q cmd=%v", model.ComposerValue(), model.overlay, cmd)
	}
}

func TestCancellationQuitHelpAndMinimumSize(t *testing.T) {
	model := New(testSnapshot())
	model.snapshot.PendingTool = "agent"
	_, cmd := updateKey(model, keyCtrl('c'))
	assertIntent(t, cmd, IntentCancel, "")
	model.snapshot.PendingTool = "none"
	model.snapshot.Mode = "ready"
	model, cmd = updateKey(model, keyCtrl('c'))
	if cmd != nil || !model.quitArmed {
		t.Fatal("first idle ctrl+c did not arm")
	}
	_, cmd = updateKey(model, keyCtrl('c'))
	assertIntent(t, cmd, IntentQuit, "")
	_, cmd = updateKey(model, keyCtrl('d'))
	assertIntent(t, cmd, IntentQuit, "")

	model = New(testSnapshot())
	model, _ = updateKey(model, keyText("?"))
	if model.overlay != overlayHelp {
		t.Fatalf("help overlay=%q", model.overlay)
	}
	model.closeOverlay()
	model.snapshot.Status = "notice"
	model.quitArmed = true
	model, _ = updateKey(model, keyCode(tea.KeyEsc))
	if model.snapshot.Status != "" || model.quitArmed {
		t.Fatalf("escape status=%q armed=%v", model.snapshot.Status, model.quitArmed)
	}

	model.width, model.height = 59, 17
	model.resize()
	if !strings.Contains(model.View().Content, "needs at least") || model.View().Cursor != nil {
		t.Fatalf("minimum view=%q", model.View().Content)
	}
	model, cmd = updateKey(model, keyText("x"))
	if cmd != nil {
		t.Fatal("small terminal accepted text")
	}
	_, cmd = updateKey(model, keyCtrl('d'))
	assertIntent(t, cmd, IntentQuit, "")
}

func TestTranscriptViewportClassificationAndRendering(t *testing.T) {
	model := New(testSnapshot())
	model.snapshot.Secrets = []string{"secret"}
	entries := []TranscriptEntry{
		{Kind: TranscriptUser, Text: "user: hello secret"},
		{Kind: TranscriptAssistant, Text: "assistant: answer"},
		{Kind: TranscriptTool, Text: "tool output"},
		{Kind: TranscriptMCP, Text: "mcp output"},
		{Kind: TranscriptVerified, Text: "verified fact"},
		{Kind: TranscriptWarning, Text: "warning"},
		{Kind: TranscriptBlocked, Text: "blocked"},
		{Kind: TranscriptUnsupported, Text: "unsupported"},
		{Kind: TranscriptAttachment, Text: "file.bin"},
		{Kind: TranscriptNote, Text: strings.Repeat("long ", 40)},
	}
	model = updateModel(t, model, TranscriptMsg{Entries: entries})
	if len(model.Entries()) != len(entries) || strings.Contains(model.Entries()[0].Text, "secret") || strings.HasPrefix(model.Entries()[0].Text, "user:") {
		t.Fatalf("entries=%#v", model.Entries())
	}
	if model.View().MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("chat mouse mode=%v", model.View().MouseMode)
	}
	model = updateModel(t, model, tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	model = updateModel(t, model, keyCode(tea.KeyPgUp))
	for index := 0; index < 30; index++ {
		model = updateModel(t, model, TranscriptMsg{Entries: []TranscriptEntry{{Kind: TranscriptNote, Text: "scroll history"}}})
	}
	model.transcript.GotoTop()
	model = updateModel(t, model, TranscriptMsg{Entries: []TranscriptEntry{{Kind: TranscriptAssistant, Text: "new while paused"}}})
	if !model.newOutput || !strings.Contains(ansi.Strip(model.View().Content), "new output  PgDn to follow") {
		t.Fatalf("paused follow state newOutput=%v\n%s", model.newOutput, ansi.Strip(model.View().Content))
	}
	model.transcript.GotoBottom()
	model = updateModel(t, model, keyCode(tea.KeyPgDown))
	if model.newOutput {
		t.Fatal("new-output notice remained at bottom")
	}
	model = updateModel(t, model, TranscriptMsg{})
	model = updateModel(t, model, ClearTranscriptMsg{})
	if len(model.Entries()) != 0 || model.newOutput {
		t.Fatalf("clear entries=%v newOutput=%v", model.Entries(), model.newOutput)
	}
}

func TestTranscriptBoundsLabelsAndProfile(t *testing.T) {
	lines := make([]string, 20)
	for index := range lines {
		lines[index] = "line"
	}
	entries := []TranscriptEntry{
		{Kind: TranscriptUser, Text: "hello"},
		{Kind: TranscriptTool, Text: strings.Join(lines, "\n")},
		{Kind: TranscriptMCP, Text: strings.Join(lines, "\n")},
	}
	plain := ansi.Strip(renderTranscript(entries, 80, colorprofile.ASCII))
	if !strings.Contains(plain, "user        hello") || strings.Count(plain, "output truncated") != 2 {
		t.Fatalf("bounded transcript:\n%s", plain)
	}
	styled := renderTranscript(entries[:1], 80, colorprofile.TrueColor)
	if !strings.Contains(styled, "\x1b[") {
		t.Fatalf("truecolor transcript has no styling: %q", styled)
	}
}

func TestResponsiveBranchesAndOverlayChat(t *testing.T) {
	model := New(testSnapshot())
	model.width, model.height = 0, 0
	model.resize()
	if model.width != DefaultWidth || model.height != DefaultHeight {
		t.Fatalf("default size=%dx%d", model.width, model.height)
	}
	model.width = 140
	if model.laneWidth() != 112 {
		t.Fatalf("wide lane=%d", model.laneWidth())
	}
	model.width, model.height = 60, 18
	model.openOverlay(overlayCommands)
	model.resize()
	model.snapshot.Status = "working"
	model.appendTranscript([]TranscriptEntry{{Kind: TranscriptUser, Text: "hello"}})
	view := model.View()
	if view.Cursor != nil || !strings.Contains(ansi.Strip(view.Content), "Filter:") || !strings.Contains(ansi.Strip(view.Content), "working") {
		t.Fatalf("overlay chat view:\n%s", ansi.Strip(view.Content))
	}
	model.closeOverlay()
	model.profile = colorprofile.ASCII
	if strings.Contains(model.View().Content, "\x1b[") {
		t.Fatalf("ASCII view contains ANSI: %q", model.View().Content)
	}
	model = New(testSnapshot())
	model.snapshot.MemoryStatus = "missing"
	if !strings.Contains(ansi.Strip(model.View().Content), "/init-re") {
		t.Fatal("missing-memory tip absent")
	}
	canvas := []string{""}
	putBlock(canvas, -2, 0, "a\nb\nc")
	putBlock(canvas, 2, 0, "outside")
	_ = wrapText("a", 0)
}

func TestResponsiveFramesColorProfilesAndUnicode(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{
		{140, 40}, {120, 32}, {80, 24}, {60, 20}, {59, 17},
	}
	for _, size := range sizes {
		model := New(testSnapshot())
		model.width, model.height = size.width, size.height
		model.resize()
		view := model.View()
		plain := ansi.Strip(view.Content)
		if got := strings.Count(plain, "\n") + 1; got != size.height {
			t.Fatalf("%dx%d frame height=%d", size.width, size.height, got)
		}
		for row, line := range strings.Split(plain, "\n") {
			if lipgloss.Width(line) > size.width {
				t.Fatalf("%dx%d row %d width=%d: %q", size.width, size.height, row, lipgloss.Width(line), line)
			}
		}
		if view.Cursor != nil && (view.Cursor.X < 0 || view.Cursor.X >= size.width || view.Cursor.Y < 0 || view.Cursor.Y >= size.height) {
			t.Fatalf("%dx%d cursor=%#v", size.width, size.height, view.Cursor)
		}
	}

	profiles := []struct {
		profile colorprofile.Profile
		marker  string
	}{
		{colorprofile.ASCII, ""},
		{colorprofile.ANSI, "\x1b["},
		{colorprofile.ANSI256, "38;5;"},
		{colorprofile.TrueColor, "38;2;"},
	}
	for _, tc := range profiles {
		model := New(testSnapshot())
		model = updateModel(t, model, ColorProfileMsg{Profile: tc.profile})
		content := model.View().Content
		if tc.profile == colorprofile.ASCII && strings.Contains(content, "\x1b[") {
			t.Fatalf("ASCII content contains ANSI: %q", content)
		}
		if tc.marker != "" && !strings.Contains(content, tc.marker) {
			t.Fatalf("profile %v missing %q", tc.profile, tc.marker)
		}
	}

	t.Setenv("NO_COLOR", "1")
	noColor := New(testSnapshot())
	if noColor.profile != colorprofile.ASCII || strings.Contains(noColor.View().Content, "\x1b[") {
		t.Fatalf("NO_COLOR profile=%v", noColor.profile)
	}

	unicodeModel := New(testSnapshot())
	unicodeModel.width, unicodeModel.height = 60, 20
	unicodeModel.resize()
	unicodeModel.composer.SetValue("解析 e\u0301 " + strings.Repeat("long-model-value ", 8))
	unicodeModel.composer.MoveToEnd()
	unicodeView := unicodeModel.View()
	if unicodeView.Cursor == nil || unicodeView.Cursor.X >= unicodeModel.width || unicodeView.Cursor.Y >= unicodeModel.height {
		t.Fatalf("unicode cursor=%#v", unicodeView.Cursor)
	}
	if got := padRight("解析", 6); lipgloss.Width(got) != 6 {
		t.Fatalf("display-width padding=%q width=%d", got, lipgloss.Width(got))
	}
}

func TestComponentChromeRespectsColorProfile(t *testing.T) {
	tests := []struct {
		name       string
		profile    colorprofile.Profile
		forbidden  []string
		wantEscape bool
	}{
		{name: "ansi", profile: colorprofile.ANSI, forbidden: []string{"[38;5;", "[38;2;"}, wantEscape: true},
		{name: "ansi256", profile: colorprofile.ANSI256, forbidden: []string{"[38;2;"}, wantEscape: true},
		{name: "ascii", profile: colorprofile.ASCII, forbidden: []string{"\x1b["}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := New(testSnapshot())
			model = updateModel(t, model, tea.WindowSizeMsg{Width: 120, Height: 32})
			model = updateModel(t, model, ColorProfileMsg{Profile: test.profile})
			model = updateModel(t, model, keyText("/"))
			frame := model.View().Content
			for _, forbidden := range test.forbidden {
				if strings.Contains(frame, forbidden) {
					t.Fatalf("frame contains forbidden profile sequence %q", forbidden)
				}
			}
			if test.wantEscape && !strings.Contains(frame, "\x1b[") {
				t.Fatal("styled profile rendered without ANSI sequences")
			}
		})
	}
}

func TestParseRenderAndHelpers(t *testing.T) {
	cases := map[string]TranscriptKind{
		"user: x": TranscriptUser, "assistant: x": TranscriptAssistant,
		"tool: x": TranscriptTool, "mcp: x": TranscriptMCP,
		"verified: x": TranscriptVerified, "warning: x": TranscriptWarning,
		"blocked: x": TranscriptBlocked, "unsupported: x": TranscriptUnsupported,
		"attachment: x": TranscriptAttachment, "plain": TranscriptNote,
	}
	for input, kind := range cases {
		if got := ParseEntry(input); got.Kind != kind || got.Text == "" {
			t.Fatalf("ParseEntry(%q)=%#v", input, got)
		}
	}
	snapshot := testSnapshot()
	snapshot.MemoryStatus = "unsupported: missing"
	content := Render(snapshot, []TranscriptEntry{{Kind: TranscriptAssistant, Text: "hello"}}, 80, 24)
	for _, want := range []string{"assistant", "hello", "Ask RecompHamr"} {
		if !strings.Contains(ansi.Strip(content), want) {
			t.Fatalf("render missing %q:\n%s", want, ansi.Strip(content))
		}
	}
	if !memoryNeedsInit("missing") || !memoryNeedsInit("unsupported") || memoryNeedsInit("verified") {
		t.Fatal("memoryNeedsInit classification failed")
	}
	if readiness(Snapshot{PendingTool: "tool"}) != "working" || readiness(Snapshot{}) != "ready" || readiness(Snapshot{Mode: "idle"}) != "idle" {
		t.Fatal("readiness failed")
	}
	if truncate("abcdef", 3) == "abcdef" || truncate("abc", 3) != "abc" {
		t.Fatal("truncate failed")
	}
	if wrapText("abcdef", 3) == "abcdef" {
		t.Fatal("wrapText failed")
	}
}

func testSnapshot() Snapshot {
	return Snapshot{
		Env:           commands.Environment{},
		Mode:          "ready",
		ActiveModel:   "local",
		ActiveSkill:   "none",
		MCPStatus:     "manager wired",
		ContextStatus: "context=32768",
		PendingTool:   "none",
		MemoryStatus:  "verified",
	}
}

func updateModel(t *testing.T, model Model, msg tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(msg)
	return updated.(Model)
}

func updateKey(model Model, key tea.KeyPressMsg) (Model, tea.Cmd) {
	updated, cmd := model.Update(key)
	return updated.(Model), cmd
}

func assertIntent(t *testing.T, cmd tea.Cmd, kind IntentKind, value string) {
	t.Helper()
	if cmd == nil {
		t.Fatalf("missing %s intent", kind)
	}
	got, ok := cmd().(IntentMsg)
	if !ok || got.Kind != kind || got.Value != value {
		t.Fatalf("intent=%#v want kind=%s value=%q", got, kind, value)
	}
}

func selectItem(t *testing.T, model *Model, name string) {
	t.Helper()
	for index, item := range model.picker.Items() {
		if item.(pickerItem).name == name {
			model.picker.Select(index)
			return
		}
	}
	t.Fatalf("picker item %q not found", name)
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

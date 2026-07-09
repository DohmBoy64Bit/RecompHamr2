package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"recomphamr2/internal/llm"
)

type observingModel struct {
	msgs     []llm.Message
	err      error
	i        int
	seenLens []int
	seen     [][]llm.Message
}

type cancelingModel struct {
	cancel context.CancelFunc
}

func (m cancelingModel) Next(context.Context, []llm.Message) (llm.Message, error) {
	m.cancel()
	return llm.Message{}, errors.New("backend after cancel")
}

func (m *observingModel) Next(_ context.Context, history []llm.Message) (llm.Message, error) {
	if m.err != nil {
		return llm.Message{}, m.err
	}
	if m.i >= len(m.msgs) {
		return llm.Message{}, errors.New("no scripted message")
	}
	m.seenLens = append(m.seenLens, len(m.msgs[:m.i]))
	m.seen = append(m.seen, cloneMessages(history))
	msg := m.msgs[m.i]
	m.i++
	return msg, nil
}

func TestLoopRunsToolsThenFinal(t *testing.T) {
	model := &observingModel{msgs: []llm.Message{
		{Role: "assistant", Tools: []llm.ToolCall{{ID: "1", Name: "powershell", Arguments: map[string]any{"cmd": "Write-Output hi"}}}},
		{Role: "assistant", Content: "done"},
	}}
	var saw llm.ToolCall
	loop := Loop{Model: model, RunTool: func(_ context.Context, call llm.ToolCall) (string, error) {
		saw = call
		return "ok", nil
	}}
	got, err := loop.Run(context.Background(), []llm.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if saw.ID != "1" || got[2].Role != "tool" || got[2].ToolCallID != "1" || got[2].ToolName != "powershell" {
		t.Fatalf("tool pairing mismatch: saw=%+v transcript=%+v", saw, got)
	}
	if got[len(got)-1].Content != "done" {
		t.Fatalf("last message = %+v", got[len(got)-1])
	}
	if len(model.seenLens) != 2 {
		t.Fatalf("model calls = %d, want 2", len(model.seenLens))
	}
}

func TestLoopInjectsProjectMemory(t *testing.T) {
	model := &observingModel{msgs: []llm.Message{{Role: "assistant", Content: "done"}}}
	loop := Loop{
		Model:                  model,
		RunTool:                func(context.Context, llm.ToolCall) (string, error) { return "ok", nil },
		ProjectMemory:          "# Project State\nverified target",
		ProjectMemorySource:    ".rehamr/REPHAMR_STATE.md",
		ProjectMemoryMaxTokens: 100,
	}
	got, err := loop.Run(context.Background(), []llm.Message{{Role: "user", Content: "task"}})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(model.seen) != 1 || len(model.seen[0]) != 2 {
		t.Fatalf("model saw history = %#v", model.seen)
	}
	if model.seen[0][0].Role != "system" || !strings.Contains(model.seen[0][0].Content, "verified target") {
		t.Fatalf("memory not injected into model context: %#v", model.seen[0])
	}
	if got[0].Role != "system" || got[len(got)-1].Content != "done" {
		t.Fatalf("memory transcript shape = %#v", got)
	}
}

func TestLoopFailureModes(t *testing.T) {
	if _, err := (Loop{}).Run(context.Background(), nil); err == nil {
		t.Fatal("unconfigured Loop succeeded")
	}
	model := &observingModel{msgs: []llm.Message{{Role: "assistant", Tools: []llm.ToolCall{{ID: "x", Name: "x"}}}}}
	loop := Loop{Model: model, MaxRounds: 1, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "", errors.New("nope") }}
	got, err := loop.Run(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "exceeded 1 model rounds") {
		t.Fatalf("Run() err = %v, want exceeded", err)
	}
	if !strings.Contains(got[len(got)-1].Content, "blocked: nope") {
		t.Fatalf("tool error not recorded: %+v", got)
	}
}

func TestLoopModelErrorAndCancellation(t *testing.T) {
	model := &observingModel{err: errors.New("backend")}
	loop := Loop{Model: model, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }}
	got, err := loop.Run(context.Background(), []llm.Message{{Role: "user"}})
	if err == nil || !strings.Contains(err.Error(), "backend") || len(got) != 1 {
		t.Fatalf("Run() got len=%d err=%v, want preserved history and backend error", len(got), err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err = loop.Run(ctx, []llm.Message{{Role: "user"}}); !errors.Is(err, context.Canceled) || len(got) != 1 {
		t.Fatalf("pre-cancel Run() got len=%d err=%v", len(got), err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	cancelModel := &observingModel{msgs: []llm.Message{{Role: "assistant", Tools: []llm.ToolCall{{ID: "c", Name: "powershell"}}}}}
	cancelLoop := Loop{Model: cancelModel, RunTool: func(context.Context, llm.ToolCall) (string, error) {
		cancel()
		return "", context.Canceled
	}}
	got, err = cancelLoop.Run(ctx, []llm.Message{{Role: "user"}})
	if !errors.Is(err, context.Canceled) || got[len(got)-1].Content != "(cancelled)" {
		t.Fatalf("tool cancel transcript=%+v err=%v", got, err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	got, err = (Loop{Model: cancelingModel{cancel: cancel}, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }}).Run(ctx, []llm.Message{{Role: "user"}})
	if !errors.Is(err, context.Canceled) || len(got) != 1 {
		t.Fatalf("model cancel error got len=%d err=%v", len(got), err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	batch := &observingModel{msgs: []llm.Message{{Role: "assistant", Tools: []llm.ToolCall{{ID: "1", Name: "read_file"}, {ID: "2", Name: "read_file"}}}}}
	calls := 0
	got, err = (Loop{Model: batch, RunTool: func(context.Context, llm.ToolCall) (string, error) {
		calls++
		cancel()
		return "ok", nil
	}}).Run(ctx, []llm.Message{{Role: "user"}})
	if !errors.Is(err, context.Canceled) || calls != 1 || got[len(got)-1].Content != "ok" {
		t.Fatalf("between-tool cancel calls=%d transcript=%+v err=%v", calls, got, err)
	}
}

func TestRepeatedFailureNudgeFiresAndResets(t *testing.T) {
	msgs := make([]llm.Message, 0, failureNudgeStreak+1)
	for i := 0; i < failureNudgeStreak; i++ {
		msgs = append(msgs, llm.Message{Role: "assistant", Tools: []llm.ToolCall{{
			ID:        string(rune('a' + i)),
			Name:      "powershell",
			Arguments: map[string]any{"cmd": "make build\nsecond line"},
		}}})
	}
	msgs = append(msgs, llm.Message{Role: "assistant", Content: "blocked: compiler still fails"})
	model := &observingModel{msgs: msgs}
	loop := Loop{Model: model, RunTool: func(context.Context, llm.ToolCall) (string, error) {
		return "boom\n(exit: exit status 1)", nil
	}}
	got, err := loop.Run(context.Background(), []llm.Message{{Role: "user", Content: "build"}})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if countContaining(got, "same target failed") != 1 {
		t.Fatalf("failure nudge count mismatch: %+v", got)
	}
	if !strings.Contains(flatten(got), "last 5 tool calls") {
		t.Fatalf("failure nudge did not name count: %+v", got)
	}
}

func TestFailureNudgeSkipsCancelledReadContentAndDistinctTargets(t *testing.T) {
	state := turnState{lastToolKey: "powershell|build"}
	state.recordToolOutcome("powershell", "partial\n(cancelled)")
	if state.failStreak != 0 || state.failureNudge() != "" {
		t.Fatalf("cancelled result counted as failure: %+v", state)
	}
	for i := 0; i < failureNudgeStreak+1; i++ {
		call := llm.ToolCall{Name: "powershell", Arguments: map[string]any{"cmd": "echo " + string(rune('a'+i))}}
		state.lastToolKey = toolTargetKey(call)
		state.recordToolOutcome(call.Name, "(exit: exit status 1)")
	}
	if state.failStreak != 1 || state.failureNudge() != "" {
		t.Fatalf("distinct targets should not nudge: %+v", state)
	}
	state.lastToolKey = "read_file|x"
	state.recordToolOutcome("read_file", "(ns foo)\n(defn bar [] 1)")
	if state.failStreak != 0 {
		t.Fatalf("read_file leading paren content counted as failure: %+v", state)
	}
}

func TestRunawayNudgeFiresOnceAfterBatch(t *testing.T) {
	model := &observingModel{msgs: []llm.Message{
		{Role: "assistant", Tools: []llm.ToolCall{
			{ID: "1", Name: "read_file", Arguments: map[string]any{"path": "a"}},
			{ID: "2", Name: "read_file", Arguments: map[string]any{"path": "b"}},
			{ID: "3", Name: "read_file", Arguments: map[string]any{"path": "c"}},
		}},
		{Role: "assistant", Content: "done"},
	}}
	loop := Loop{Model: model, MaxToolCalls: 2, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }}
	got, err := loop.Run(context.Background(), []llm.Message{{Role: "user", Content: "inspect"}})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if countContaining(got, "tool calls so far") != 1 {
		t.Fatalf("runaway nudge count mismatch: %+v", got)
	}
}

func TestVerifyNudgeRepromptsOnceAndHonorsLabels(t *testing.T) {
	msgs := make([]llm.Message, 0, verifyNudgeMinTools+3)
	for i := 0; i < verifyNudgeMinTools; i++ {
		msgs = append(msgs, llm.Message{Role: "assistant", Tools: []llm.ToolCall{{ID: string(rune('a' + i)), Name: "read_file", Arguments: map[string]any{"path": "x"}}}})
	}
	msgs = append(msgs, llm.Message{Role: "assistant", Content: "Done - all good."})
	msgs = append(msgs, llm.Message{Role: "assistant", Content: "Verified one check; unverified: runtime check - no runner."})
	model := &observingModel{msgs: msgs}
	loop := Loop{Model: model, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }}
	got, err := loop.Run(context.Background(), []llm.Message{{Role: "user", Content: "do it"}})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if countContaining(got, "acceptance criteria") != 1 {
		t.Fatalf("verify nudge count mismatch: %+v", got)
	}

	labeled := &observingModel{msgs: append(msgs[:verifyNudgeMinTools], llm.Message{Role: "assistant", Content: "blocked: no compiler"})}
	got, err = (Loop{Model: labeled, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }}).Run(context.Background(), []llm.Message{{Role: "user"}})
	if err != nil || countContaining(got, "acceptance criteria") != 0 {
		t.Fatalf("labeled finish should not nudge: got=%+v err=%v", got, err)
	}
}

func TestEmptyReplyNudgeAndStop(t *testing.T) {
	model := &observingModel{msgs: []llm.Message{
		{Role: "assistant"},
		{Role: "assistant"},
	}}
	got, err := (Loop{Model: model, RunTool: func(context.Context, llm.ToolCall) (string, error) { return "ok", nil }}).Run(context.Background(), []llm.Message{{Role: "user"}})
	if err == nil || !strings.Contains(err.Error(), "no reply and no tool call after retry") {
		t.Fatalf("empty retry err=%v transcript=%+v", err, got)
	}
	if countContaining(got, "no reply and no tool call") != 1 {
		t.Fatalf("empty nudge count mismatch: %+v", got)
	}
}

func TestHelpers(t *testing.T) {
	cases := []struct {
		name   string
		tool   string
		result string
		want   bool
	}{
		{"cancelled", "write_file", "(cancelled)", false},
		{"router blocked", "made_up", "blocked: nope", true},
		{"invalid args", "powershell", "(tool arguments were not valid JSON: x)", true},
		{"unknown", "made_up", "(unknown tool: made_up)", true},
		{"write", "write_file", "(write_file: denied)", true},
		{"edit", "edit_file", "(edit_file: missing)", true},
		{"read", "read_file", "(read_file: missing)", true},
		{"shell", "powershell", "(exit: exit status 1)", true},
		{"unknown paren", "mystery", "(mystery: bad)", true},
	}
	for _, tc := range cases {
		if got := toolResultFailed(tc.tool, tc.result); got != tc.want {
			t.Fatalf("%s: toolResultFailed() = %v, want %v", tc.name, got, tc.want)
		}
	}
	if got := argString(llm.ToolCall{}, "path"); got != "" {
		t.Fatalf("argString(nil) = %q", got)
	}
	if got := toolTargetKey(llm.ToolCall{Name: "write_file", Arguments: map[string]any{"path": "x"}}); got != "write_file|x" {
		t.Fatalf("write target = %q", got)
	}
	if !newestAssistantEmpty([]llm.Message{{Role: "assistant"}}) || newestAssistantEmpty([]llm.Message{{Role: "assistant", Content: "x"}}) {
		t.Fatal("newestAssistantEmpty mismatch")
	}
	if newestAssistantEmpty([]llm.Message{{Role: "user", Content: "x"}}) {
		t.Fatal("user-only history should not be empty assistant")
	}
	if !newestAssistantEvidenceLabeled([]llm.Message{{Role: "assistant", Content: "unsupported: x"}}) {
		t.Fatal("unsupported label not detected")
	}
	if newestAssistantEvidenceLabeled([]llm.Message{{Role: "user", Content: "unverified"}}) {
		t.Fatal("user label should not count")
	}
	if err := contextError(context.Background(), context.DeadlineExceeded); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("contextError(deadline) = %v", err)
	}
	if err := contextError(context.Background(), errors.New("tool")); err != nil {
		t.Fatalf("contextError(tool) = %v, want nil", err)
	}
}

func countContaining(messages []llm.Message, needle string) int {
	n := 0
	for _, msg := range messages {
		if strings.Contains(msg.Content, needle) {
			n++
		}
	}
	return n
}

func flatten(messages []llm.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		b.WriteString(msg.Content)
		b.WriteByte('\n')
	}
	return b.String()
}

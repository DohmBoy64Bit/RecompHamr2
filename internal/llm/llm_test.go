package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read") }

func collect(ch <-chan Event) []Event {
	var events []Event
	for event := range ch {
		events = append(events, event)
	}
	return events
}

func sse(w http.ResponseWriter, chunks ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set(headerBudgetRemaining, "0.75")
	w.Header().Set(headerContextWindow, "65536")
	for _, chunk := range chunks {
		fmt.Fprintf(w, "data: %s\n\n", chunk)
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
}

func TestParseSSEContentReasoningToolArgsAndDone(t *testing.T) {
	stream := strings.Join([]string{
		`: keepalive`,
		`data: {"choices":[{"delta":{"reasoning":"think"}}]}`,
		`data: {"choices":[{"delta":{"content":"hi"}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"c1","function":{"name":"bash"}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"cmd"}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\":\"echo hi\"}"}}]}}]}`,
		`data: [DONE]`,
	}, "\n")
	events, err := ParseSSE(strings.NewReader(stream))
	if err != nil {
		t.Fatalf("ParseSSE() error = %v", err)
	}
	var kinds []EventKind
	for _, event := range events {
		kinds = append(kinds, event.Kind)
	}
	want := []EventKind{EventReasoning, EventContent, EventToolArgs, EventToolArgs, EventToolCall, EventDone}
	if fmt.Sprint(kinds) != fmt.Sprint(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	if events[4].Tool.Name != "bash" || events[4].Tool.Arguments["cmd"] != "echo hi" {
		t.Fatalf("tool event = %+v", events[4])
	}
	if events[5].Final.Content != "hi" || len(events[5].Final.Tools) != 1 {
		t.Fatalf("final event = %+v", events[5])
	}
}

func TestParseSSEErrorsAndMalformedArgs(t *testing.T) {
	if _, err := ParseSSE(strings.NewReader("data: nope\n")); err == nil {
		t.Fatal("ParseSSE() accepted invalid JSON")
	}
	if _, err := ParseSSE(errReader{}); err == nil {
		t.Fatal("ParseSSE() accepted reader error")
	}
	events, err := ParseSSE(strings.NewReader(`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"c1","function":{"name":"bash","arguments":"{"}}]}}]}`))
	if err != nil {
		t.Fatalf("ParseSSE() malformed args error = %v", err)
	}
	if _, ok := events[1].Tool.Arguments["_parse_error"]; !ok {
		t.Fatalf("malformed args should be marked: %+v", events)
	}
}

func TestHeaderParsing(t *testing.T) {
	h := http.Header{}
	if BudgetFromHeaders(h).Set || ContextWindowFromHeaders(h) != 0 {
		t.Fatal("missing headers should return zero values")
	}
	h.Set(headerBudgetRemaining, "2")
	if got := BudgetFromHeaders(h); !got.Set || got.Remaining != 1 {
		t.Fatalf("budget clamp high = %+v", got)
	}
	h.Set(headerBudgetRemaining, "-1")
	if got := BudgetFromHeaders(h); !got.Set || got.Remaining != 0 {
		t.Fatalf("budget clamp low = %+v", got)
	}
	for _, value := range []string{"bad", "NaN", "+Inf"} {
		h.Set(headerBudgetRemaining, value)
		if BudgetFromHeaders(h).Set {
			t.Fatalf("budget %q should be ignored", value)
		}
	}
	for _, value := range []string{"bad", "10", "999999999"} {
		h.Set(headerContextWindow, value)
		if ContextWindowFromHeaders(h) != 0 {
			t.Fatalf("context window %q should be ignored", value)
		}
	}
	h.Set(headerContextWindow, "4096")
	if ContextWindowFromHeaders(h) != 4096 {
		t.Fatal("valid context window was not parsed")
	}
}

func TestIdleTimeoutFromEnv(t *testing.T) {
	osUnset := func() { _ = osUnsetenv("RECOMPHAMR_IDLE_TIMEOUT") }
	osUnset()
	if IdleTimeoutFromEnv() != defaultIdleTimeout {
		t.Fatal("unset idle timeout should use default")
	}
	cases := map[string]time.Duration{
		"90m":         90 * time.Minute,
		"120":         120 * time.Second,
		"bad":         defaultIdleTimeout,
		"0":           defaultIdleTimeout,
		"-1s":         defaultIdleTimeout,
		"18446744074": defaultIdleTimeout,
	}
	for value, want := range cases {
		t.Setenv("RECOMPHAMR_IDLE_TIMEOUT", value)
		if got := IdleTimeoutFromEnv(); got != want {
			t.Fatalf("IdleTimeoutFromEnv(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestBudgetTokensAndTruncate(t *testing.T) {
	if Tokens("abcde") != 2 {
		t.Fatal("Tokens should round up")
	}
	if Budget(1000) != 0 || Budget(32768) <= 0 {
		t.Fatal("Budget should floor small contexts and leave room for large ones")
	}
	short := "hello"
	if TruncateToolOutput(short, 10) != short {
		t.Fatal("short output should not truncate")
	}
	if TruncateToolOutput(short, 0) != short {
		t.Fatal("non-positive limit should not truncate")
	}
	long := strings.Repeat("a", 100) + "é-tail"
	got := TruncateToolOutput(long, 10)
	if !strings.Contains(got, "truncated middle") || !utf8.ValidString(got) {
		t.Fatalf("truncate output invalid: %q", got)
	}
	if got := TruncateToolOutput(strings.Repeat("z", 100), 2); !strings.Contains(got, "truncated middle") {
		t.Fatalf("small max should still truncate with minimum kept bytes: %q", got)
	}
	if runeBoundaryDown("abc", 99) != 3 || runeBoundaryUp("abc", 0) != 0 || runeBoundaryDown("aé", 2) != 1 || runeBoundaryUp("aé", 2) != 3 {
		t.Fatal("rune boundary helpers did not preserve UTF-8 cuts")
	}
}

func TestWithProjectMemory(t *testing.T) {
	history := []Message{{Role: "system", Content: "base"}, {Role: "user", Content: "task"}}
	got := WithProjectMemory(history, ".rehamr/REPHAMR_STATE.md", "# State\nverified", 100)
	if len(got) != 2 || got[0].Role != "system" {
		t.Fatalf("WithProjectMemory() shape = %#v", got)
	}
	for _, want := range []string{"base", "Project memory from .rehamr/REPHAMR_STATE.md", "verified"} {
		if !strings.Contains(got[0].Content, want) {
			t.Fatalf("WithProjectMemory() missing %q:\n%s", want, got[0].Content)
		}
	}
	history[0].Content = "mutated"
	if strings.Contains(got[0].Content, "mutated") {
		t.Fatalf("WithProjectMemory() did not clone messages: %#v", got)
	}
}

func TestWithProjectMemoryPrependsDefaultsAndTruncates(t *testing.T) {
	got := WithProjectMemory([]Message{{Role: "user", Content: "task"}}, "", strings.Repeat("abcd", 30)+"é-tail", 4)
	if len(got) != 2 || got[0].Role != "system" || !strings.Contains(got[0].Content, "unknown source") {
		t.Fatalf("WithProjectMemory() default source shape = %#v", got)
	}
	if !strings.Contains(got[0].Content, "truncated middle") || !utf8.ValidString(got[0].Content) {
		t.Fatalf("WithProjectMemory() did not truncate safely:\n%s", got[0].Content)
	}
	empty := WithProjectMemory([]Message{{Role: "user", Content: "task"}}, "state", " \n\t", 1)
	if len(empty) != 1 || empty[0].Role != "user" {
		t.Fatalf("WithProjectMemory() empty memory = %#v", empty)
	}
}

func TestPackPreservesWireValidity(t *testing.T) {
	history := []Message{
		{Role: "system", Content: "nudge"},
		{Role: "user", Content: "task"},
		{Role: "assistant", Tools: []ToolCall{{ID: "old", Name: "bash"}}},
		{Role: "tool", ToolCallID: "old", ToolName: "bash", Content: strings.Repeat("x", 200)},
		{Role: "assistant", Tools: []ToolCall{{ID: "dangling", Name: "bash"}}},
		{Role: "tool", ToolCallID: "orphan", Content: "bad"},
		{Role: "assistant", Content: "final"},
	}
	packed := Pack(history, 20)
	if packed[0].Role != "user" || packed[0].Content != "task" {
		t.Fatalf("original user task should anchor packed history: %+v", packed)
	}
	for _, msg := range packed {
		if msg.Role == "system" {
			t.Fatalf("system history message should be demoted: %+v", packed)
		}
		if msg.Role == "tool" && msg.ToolCallID == "orphan" {
			t.Fatalf("orphan tool survived: %+v", packed)
		}
		for _, call := range msg.Tools {
			if call.ID == "dangling" {
				t.Fatalf("dangling assistant call survived: %+v", packed)
			}
		}
	}
	history[0].Role = "mutated"
	if packed[0].Role == "mutated" {
		t.Fatal("Pack should return a copy")
	}
}

func TestPackNewestToolGroupRecovery(t *testing.T) {
	history := []Message{
		{Role: "user", Content: "task"},
		{Role: "assistant", Content: strings.Repeat("x", 1000)},
		{Role: "assistant", Tools: []ToolCall{{ID: "c1", Name: "bash"}, {ID: "c2", Name: "bash"}}},
		{Role: "tool", ToolCallID: "c1", Content: "one"},
		{Role: "tool", ToolCallID: "c2", Content: "two"},
	}
	packed := Pack(history, 1)
	if len(packed) != 4 || packed[1].Role != "assistant" || len(packed[1].Tools) != 2 {
		t.Fatalf("newest tool group not recovered: %+v", packed)
	}
	partial := history[:4]
	if got := Pack(partial, 1); len(got) != 1 || got[0].Role != "user" {
		t.Fatalf("partial parallel tool group should be dropped except anchor: %+v", got)
	}
	if got := Pack(nil, 1); len(got) != 0 {
		t.Fatalf("empty pack = %+v", got)
	}
	noUser := Pack([]Message{{Role: "assistant", Content: "ok"}}, 1)
	if len(noUser) != 1 || noUser[0].Role != "assistant" {
		t.Fatalf("no-user history should keep assistant only: %+v", noUser)
	}
	noOwner := newestToolGroup([]Message{{Role: "tool", ToolCallID: "missing"}})
	if noOwner != nil {
		t.Fatalf("unowned tool should not recover: %+v", noOwner)
	}
	if empty := newestToolGroup(nil); empty != nil {
		t.Fatalf("empty history should not recover a tool group: %+v", empty)
	}
	noMatchingOwner := newestToolGroup([]Message{{Role: "assistant", Tools: []ToolCall{{ID: "other", Name: "bash"}}}, {Role: "tool", ToolCallID: "missing"}})
	if noMatchingOwner != nil {
		t.Fatalf("tool without matching assistant owner should not recover: %+v", noMatchingOwner)
	}
	noMatchingOwner = newestToolGroup([]Message{{Role: "user", Content: "task"}, {Role: "assistant", Tools: []ToolCall{{ID: "other", Name: "bash"}}}, {Role: "tool", ToolCallID: "missing"}})
	if noMatchingOwner != nil {
		t.Fatalf("tool without matching owner after non-assistant scan should not recover: %+v", noMatchingOwner)
	}
	notTool := newestToolGroup([]Message{{Role: "assistant", Content: "x"}})
	if notTool != nil {
		t.Fatalf("non-tool tail should not recover: %+v", notTool)
	}
	copied := Pack([]Message{{Role: "assistant", Content: "x", Tools: []ToolCall{{ID: "id", Name: "n", Arguments: map[string]any{"k": "v"}}}, ToolCallID: "id", ToolName: "n"}}, 0)
	copied[0].Tools[0].Name = "changed"
	again := Pack([]Message{{Role: "assistant", Content: "x", Tools: []ToolCall{{ID: "id", Name: "n", Arguments: map[string]any{"k": "v"}}}, ToolCallID: "id", ToolName: "n"}}, 0)
	if again[0].Tools[0].Name != "n" {
		t.Fatal("Pack should copy tool slices")
	}
	if got := messageTokens(Message{Role: "tool", Content: "x", ToolCallID: "id", ToolName: "bash", Tools: []ToolCall{{ID: "id", Name: "bash", Arguments: map[string]any{"cmd": "ls"}}}}); got <= 0 {
		t.Fatal("messageTokens should count tool metadata")
	}
	if got := newestToolGroup([]Message{{Role: "user", Content: "task"}, {Role: "assistant", Tools: []ToolCall{{ID: "c", Name: "bash"}}}, {Role: "tool", ToolCallID: "c", Content: "ok"}}); len(got) != 2 {
		t.Fatalf("newestToolGroup should skip non-assistant history and recover owner: %+v", got)
	}
	if got := newestToolGroup([]Message{{Role: "assistant", Tools: []ToolCall{{ID: "c", Name: "bash"}, {Name: "empty"}}}, {Role: "tool", ToolCallID: "c", Content: "ok"}}); len(got) != 2 {
		t.Fatalf("newestToolGroup should ignore empty issued IDs: %+v", got)
	}
	if got := anchorUser([]Message{{Role: "user", Content: "kept"}}, history); len(got) != 1 || got[0].Content != "kept" {
		t.Fatalf("anchorUser should not duplicate surviving user: %+v", got)
	}
	if got := demoteSystem([]Message{{Role: "assistant", Content: "ok"}}); got[0].Role != "assistant" {
		t.Fatalf("demoteSystem should leave non-system roles alone: %+v", got)
	}
	if got := demoteSystem([]Message{{Role: "system", Content: "note"}}); got[0].Role != "user" {
		t.Fatalf("demoteSystem should rewrite system roles: %+v", got)
	}
}

func TestChatSuccessRequestAndEvents(t *testing.T) {
	var body, auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		data, _ := io.ReadAll(r.Body)
		body = string(data)
		sse(w,
			`{"choices":[{"delta":{"reasoning":"hm"}}]}`,
			`{"choices":[{"delta":{"content":"Hel"}}]}`,
			`{"choices":[{"delta":{"content":"lo"}}],"usage":{"completion_tokens":7,"prompt_tokens":11}}`,
		)
	}))
	defer srv.Close()
	client := NewClient(srv.URL, "model-a", "sk-test")
	events := collect(client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, nil))
	if auth != "Bearer sk-test" {
		t.Fatalf("auth = %q", auth)
	}
	for _, want := range []string{`"model":"model-a"`, `"stream_options":{"include_usage":true}`, `"reasoning_effort":"high"`, `"content":"hi"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("request missing %s: %s", want, body)
		}
	}
	var content string
	var done Event
	for _, event := range events {
		if event.Kind == EventContent {
			content += event.Content
		}
		if event.Kind == EventDone {
			done = event
		}
	}
	if content != "Hello" || done.Final.Content != "Hello" || done.Tokens != 7 || done.PromptTokens != 11 || done.ContextWindow != 65536 || done.Budget.Remaining != 0.75 {
		t.Fatalf("events wrong: %+v", events)
	}
}

func BenchmarkPackLargeHistory(b *testing.B) {
	history := make([]Message, 0, 401)
	history = append(history, Message{Role: "system", Content: "base prompt"})
	for i := 0; i < 200; i++ {
		history = append(history,
			Message{Role: "user", Content: fmt.Sprintf("task %03d %s", i, strings.Repeat("a", 256))},
			Message{Role: "assistant", Content: fmt.Sprintf("answer %03d %s", i, strings.Repeat("b", 256))},
		)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		packed := Pack(history, 4096)
		if len(packed) == 0 {
			b.Fatal("Pack returned empty history")
		}
	}
}

func TestChatUsesDefaultIdleWhenConfiguredNonPositive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sse(w, `{"choices":[{"delta":{"content":"ok"}}]}`)
	}))
	defer srv.Close()
	client := NewClient(srv.URL, "m", "")
	client.IdleTimeout = 0
	for _, event := range collect(client.Chat(context.Background(), nil, nil)) {
		if event.Kind == EventError {
			t.Fatalf("chat should succeed with default idle fallback: %v", event.Err)
		}
	}
}

func TestChatToolCallsFragmentedByIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sse(w,
			`{"choices":[{"delta":{"tool_calls":[{"index":1,"id":"c2","function":{"name":"python"}}]}}]}`,
			`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"c1","function":{"name":"bash"}}]}}]}`,
			`{"choices":[{"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{\"cmd\":\"print()\"}"}}]}}]}`,
			`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"cmd\":\"ls\"}"}}]}}]}`,
			`{"choices":[{"delta":{},"finish_reason":"stop"}]}`,
		)
	}))
	defer srv.Close()
	var calls []ToolCall
	var argFragments string
	for _, event := range collect(NewClient(srv.URL, "m", "").Chat(context.Background(), nil, nil)) {
		if event.Kind == EventToolArgs {
			argFragments += event.Content
		}
		if event.Kind == EventToolCall {
			calls = append(calls, event.Tool)
		}
	}
	if len(calls) != 2 || calls[0].Name != "python" || calls[1].Name != "bash" {
		t.Fatalf("calls = %+v", calls)
	}
	if calls[0].Arguments["cmd"] != "print()" || calls[1].Arguments["cmd"] != "ls" || !strings.Contains(argFragments, "print") {
		t.Fatalf("args = %+v fragments=%q", calls, argFragments)
	}
}

func TestChatProviderErrorsAndFallback(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(data))
		if len(bodies) == 1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":{"message":"reasoning_effort is not supported"}}`)
			return
		}
		sse(w, `{"choices":[{"delta":{"content":"ok"}}]}`)
	}))
	defer srv.Close()
	client := NewClient(srv.URL, "m", "")
	for _, event := range collect(client.Chat(context.Background(), nil, nil)) {
		if event.Kind == EventError {
			t.Fatalf("fallback should recover: %v", event.Err)
		}
	}
	collect(client.Chat(context.Background(), nil, nil))
	if len(bodies) != 3 || !strings.Contains(bodies[0], "reasoning_effort") || strings.Contains(bodies[1], "reasoning_effort") || strings.Contains(bodies[2], "reasoning_effort") {
		t.Fatalf("fallback stickiness wrong: %#v", bodies)
	}
}

func TestChatHTTPErrorMapping(t *testing.T) {
	statuses := []struct {
		code int
		body string
		want error
		text string
	}{
		{http.StatusUnauthorized, "", ErrUnauthorized, ""},
		{http.StatusPaymentRequired, "", ErrBudgetExhausted, ""},
		{http.StatusServiceUnavailable, `{"error":{"message":"wrapped","provider_hint":"provider down"}}`, nil, "provider down"},
		{http.StatusBadGateway, "raw line\nextra", nil, "raw line"},
		{http.StatusInternalServerError, "plain", nil, "plain"},
	}
	for _, tc := range statuses {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.code)
			fmt.Fprint(w, tc.body)
		}))
		events := collect(NewClient(srv.URL, "m", "").Chat(context.Background(), nil, nil))
		srv.Close()
		if len(events) != 1 || events[0].Kind != EventError {
			t.Fatalf("events = %+v", events)
		}
		if tc.want != nil && !errors.Is(events[0].Err, tc.want) {
			t.Fatalf("err = %v, want %v", events[0].Err, tc.want)
		}
		if tc.text != "" && !strings.Contains(events[0].Err.Error(), tc.text) {
			t.Fatalf("err = %v, want text %q", events[0].Err, tc.text)
		}
	}
}

func TestTransportProbeSendAndIdleErrors(t *testing.T) {
	wrapped := ErrUnreachable{Err: errors.New("socket")}
	if wrapped.Error() != "backend unreachable: socket" || wrapped.Unwrap().Error() != "socket" {
		t.Fatalf("ErrUnreachable methods broken: %v", wrapped)
	}
	client := NewClient("http://example.invalid", "m", "")
	client.HTTP = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("dial")
	})}
	events := collect(client.Chat(context.Background(), nil, nil))
	var unreachable ErrUnreachable
	if !errors.As(events[0].Err, &unreachable) {
		t.Fatalf("expected unreachable, got %+v", events)
	}
	badURL := NewClient("://bad", "m", "")
	events = collect(badURL.Chat(context.Background(), nil, nil))
	if len(events) != 1 || events[0].Kind != EventError {
		t.Fatalf("bad URL should emit one error: %+v", events)
	}
	badJSON := NewClient("http://example.invalid", "m", "")
	events = collect(badJSON.Chat(context.Background(), []Message{{Role: "assistant", Tools: []ToolCall{{Arguments: map[string]any{"bad": func() {}}}}}}, nil))
	if len(events) != 1 || events[0].Kind != EventError {
		t.Fatalf("bad JSON should emit one error: %+v", events)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(headerBudgetRemaining, "0.5")
		w.Header().Set(headerContextWindow, "8192")
		sse(w, `{"choices":[{"delta":{"content":"ok"}}]}`)
	}))
	defer srv.Close()
	budget, window, err := NewClient(srv.URL, "m", "").Probe(context.Background())
	if err != nil || !budget.Set || budget.Remaining != 0.75 || window != 65536 {
		t.Fatalf("Probe() = %+v %d %v", budget, window, err)
	}
	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer errSrv.Close()
	_, _, err = NewClient(errSrv.URL, "m", "").Probe(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Probe unauthorized err = %v", err)
	}
	invalidStream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: nope\n\n")
	}))
	defer invalidStream.Close()
	events = collect(NewClient(invalidStream.URL, "m", "").Chat(context.Background(), nil, nil))
	if len(events) != 1 || events[0].Kind != EventError || !strings.Contains(events[0].Err.Error(), "sse payload") {
		t.Fatalf("invalid stream events = %+v", events)
	}

	blocking := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer blocking.Close()
	stalled := NewClient(blocking.URL, "m", "")
	stalled.IdleTimeout = 20 * time.Millisecond
	events = collect(stalled.Chat(context.Background(), nil, nil))
	if len(events) != 1 || events[0].Kind != EventError || !strings.Contains(events[0].Err.Error(), "stopped sending") {
		t.Fatalf("idle events = %+v", events)
	}
}

func TestSendEventAndCancelledRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan Event)
	done := make(chan bool, 1)
	go func() { done <- sendEvent(ctx, out, Event{Kind: EventContent}) }()
	cancel()
	if <-done {
		t.Fatal("sendEvent should report cancelled send")
	}

	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	_, err := readSSE(ctx, strings.NewReader(`data: {"choices":[{"delta":{"content":"x"}}]}`), BudgetStatus{}, 0, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("readSSE cancelled err = %v", err)
	}
}

func osUnsetenv(key string) error {
	return os.Unsetenv(key)
}

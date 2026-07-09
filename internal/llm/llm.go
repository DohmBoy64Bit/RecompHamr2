package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

var timeAfterFunc = time.AfterFunc

const (
	headerBudgetRemaining = "X-Budget-Remaining"
	headerContextWindow   = "X-Context-Window"
	contextWindowMin      = 1024
	contextWindowMax      = 8 * 1024 * 1024
	defaultIdleTimeout    = 5 * time.Minute
	projectMemoryPrefix   = "Project memory"
)

// Message is one OpenAI-compatible chat transcript item.
type Message struct {
	// Role is the OpenAI chat role such as system, user, assistant, or tool.
	Role string `json:"role"`
	// Content is always sent on the wire, including empty tool results.
	Content string `json:"content"`
	// Tools contains assistant tool calls.
	Tools []ToolCall `json:"tool_calls,omitempty"`
	// ToolCallID links a tool-role message to its assistant tool call.
	ToolCallID string `json:"tool_call_id,omitempty"`
	// ToolName is the optional tool name on tool-role messages.
	ToolName string `json:"name,omitempty"`
}

// ToolCall is an assistant-requested tool invocation.
type ToolCall struct {
	// ID is the provider tool-call identifier.
	ID string `json:"id"`
	// Name is the function/tool name.
	Name string `json:"name"`
	// Arguments is the parsed JSON argument map.
	Arguments map[string]any `json:"arguments,omitempty"`
}

// Tool describes one callable function exposed to the model.
type Tool struct {
	// Type is normally "function".
	Type string `json:"type"`
	// Function describes the callable function schema.
	Function FunctionDef `json:"function"`
}

// FunctionDef is the OpenAI-compatible function schema.
type FunctionDef struct {
	// Name is the function name.
	Name string `json:"name"`
	// Description explains the function to the model.
	Description string `json:"description"`
	// Parameters is a JSON Schema object for arguments.
	Parameters map[string]any `json:"parameters"`
}

// EventKind identifies a streaming event category.
type EventKind string

const (
	// EventContent contains assistant text.
	EventContent EventKind = "content"
	// EventToolCall contains an assembled tool call.
	EventToolCall EventKind = "tool_call"
	// EventDone marks stream completion.
	EventDone EventKind = "done"
	// EventError carries a provider, transport, parse, or idle-timeout error.
	EventError EventKind = "error"
	// EventReasoning carries reasoning deltas that must not enter history.
	EventReasoning EventKind = "reasoning"
	// EventToolArgs carries live tool argument fragments before assembly.
	EventToolArgs EventKind = "tool_args"
)

// Event is one parsed streaming event.
type Event struct {
	// Kind identifies the event category.
	Kind EventKind
	// Content is text, reasoning text, or a tool-argument fragment.
	Content string
	// Tool is populated on EventToolCall.
	Tool ToolCall
	// Final is populated on EventDone with assistant content and tool calls.
	Final *Message
	// Err is populated on EventError.
	Err error
	// Budget is the latest provider budget header snapshot.
	Budget BudgetStatus
	// ContextWindow is the server-provided context window, or zero when absent.
	ContextWindow int
	// Tokens is usage.completion_tokens from the stream tail.
	Tokens int
	// PromptTokens is usage.prompt_tokens from the stream tail.
	PromptTokens int
}

// BudgetStatus records the X-Budget-Remaining header when a provider sends it.
type BudgetStatus struct {
	// Set reports whether a usable header was present.
	Set bool
	// Remaining is clamped to the inclusive range [0,1].
	Remaining float64
}

// ErrBudgetExhausted maps provider HTTP 402.
var ErrBudgetExhausted = errors.New("budget depleted")

// ErrUnauthorized maps provider HTTP 401.
var ErrUnauthorized = errors.New("invalid or expired token")

// ErrUnreachable wraps transport failures.
type ErrUnreachable struct {
	// Err is the underlying transport error.
	Err error
}

// Error returns the user-facing transport message.
func (e ErrUnreachable) Error() string { return "backend unreachable: " + e.Err.Error() }

// Unwrap returns the underlying transport error.
func (e ErrUnreachable) Unwrap() error { return e.Err }

// Client sends OpenAI-compatible chat-completion requests.
type Client struct {
	// BaseURL is the endpoint root without the /v1/chat/completions suffix.
	BaseURL string
	// Model is the model identifier sent in each request.
	Model string
	// Token is the optional bearer token.
	Token string
	// HTTP is the HTTP client used for requests.
	HTTP *http.Client
	// IdleTimeout bounds silence between SSE frames.
	IdleTimeout time.Duration
	noReasoning atomic.Bool
}

// NewClient creates an OpenAI-compatible streaming client.
func NewClient(baseURL string, model string, token string) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Model:       model,
		Token:       token,
		HTTP:        &http.Client{},
		IdleTimeout: IdleTimeoutFromEnv(),
	}
}

// Chat streams one assistant response and closes the returned channel.
func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool) <-chan Event {
	out := make(chan Event, 32)
	go c.run(ctx, messages, tools, out)
	return out
}

// Probe sends a tiny request and returns header-derived provider metadata.
func (c *Client) Probe(ctx context.Context) (BudgetStatus, int, error) {
	resp, budget, _, err := c.post(ctx, chatRequest{
		Model:    c.Model,
		Messages: []Message{{Role: "user", Content: "hi"}},
		Stream:   true,
	}, false)
	if err != nil {
		return budget, 0, err
	}
	defer resp.Body.Close()
	return BudgetFromHeaders(resp.Header), ContextWindowFromHeaders(resp.Header), nil
}

// ParseSSE parses an OpenAI-compatible SSE stream into events.
func ParseSSE(r io.Reader) ([]Event, error) {
	return readSSE(context.Background(), r, BudgetStatus{}, 0, nil)
}

// BudgetFromHeaders reads and clamps X-Budget-Remaining.
func BudgetFromHeaders(h http.Header) BudgetStatus {
	raw := h.Get(headerBudgetRemaining)
	if raw == "" {
		return BudgetStatus{}
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return BudgetStatus{}
	}
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	return BudgetStatus{Set: true, Remaining: value}
}

// ContextWindowFromHeaders reads X-Context-Window and rejects unsafe values.
func ContextWindowFromHeaders(h http.Header) int {
	raw := h.Get(headerContextWindow)
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < contextWindowMin || value > contextWindowMax {
		return 0
	}
	return value
}

// IdleTimeoutFromEnv resolves RECOMPHAMR_IDLE_TIMEOUT.
func IdleTimeoutFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("RECOMPHAMR_IDLE_TIMEOUT"))
	if raw == "" {
		return defaultIdleTimeout
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return defaultIdleTimeout
	}
	d := time.Duration(seconds) * time.Second
	if d/time.Second != time.Duration(seconds) {
		return defaultIdleTimeout
	}
	return d
}

// Tokens estimates token count as bytes divided by four, rounded up.
func Tokens(text string) int {
	return (len(text) + 3) / 4
}

// Budget computes the history budget left after fixed reserves.
func Budget(contextSize int) int {
	reserve := contextSize/8 + 5500
	if contextSize <= reserve {
		return 0
	}
	remaining := contextSize - reserve
	return remaining - remaining/10
}

// TruncateToolOutput keeps head and tail text while preserving UTF-8 validity.
func TruncateToolOutput(text string, maxTokens int) string {
	if maxTokens <= 0 || Tokens(text) <= maxTokens {
		return text
	}
	keepBytes := maxTokens * 2
	if keepBytes < 8 {
		keepBytes = 8
	}
	head := runeBoundaryDown(text, keepBytes)
	tail := runeBoundaryUp(text, len(text)-keepBytes)
	return text[:head] + "\n----- truncated middle -----\n" + text[tail:]
}

// WithProjectMemory injects persistent workspace memory into model context.
func WithProjectMemory(messages []Message, source string, memory string, maxTokens int) []Message {
	content := strings.TrimSpace(memory)
	if content == "" {
		return cloneMessages(messages)
	}
	if maxTokens > 0 && Tokens(content) > maxTokens {
		content = TruncateToolOutput(content, maxTokens)
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "unknown source"
	}
	note := projectMemoryPrefix + " from " + source + ". Use as workspace context; verify facts against current files before changing behavior.\n\n" + content
	out := cloneMessages(messages)
	if len(out) > 0 && out[0].Role == "system" {
		out[0].Content = strings.TrimSpace(out[0].Content) + "\n\n" + note
		return out
	}
	return append([]Message{{Role: "system", Content: note}}, out...)
}

// Pack trims history to budget while preserving OpenAI tool-call integrity.
func Pack(messages []Message, maxTokens int) []Message {
	if maxTokens <= 0 || estimate(messages) <= maxTokens {
		return cloneMessages(messages)
	}
	kept := make([]Message, 0, len(messages))
	used := 0
	for i := len(messages) - 1; i >= 0; i-- {
		cost := messageTokens(messages[i])
		if len(kept) > 0 && used+cost > maxTokens {
			break
		}
		kept = append(kept, messages[i])
		used += cost
	}
	slices.Reverse(kept)
	kept = dropDanglingToolCalls(kept)
	kept = dropOrphanTools(kept)
	if len(kept) == 0 {
		kept = newestToolGroup(messages)
		kept = dropDanglingToolCalls(kept)
		kept = dropOrphanTools(kept)
	}
	kept = anchorUser(kept, messages)
	kept = demoteSystem(kept)
	return cloneMessages(kept)
}

func (c *Client) run(ctx context.Context, messages []Message, tools []Tool, out chan<- Event) {
	defer close(out)
	resp, budget, _, err := c.post(ctx, chatRequest{
		Model:           c.Model,
		Messages:        messages,
		Tools:           tools,
		Stream:          true,
		StreamOptions:   &streamOptions{IncludeUsage: true},
		ReasoningEffort: "high",
	}, true)
	if err != nil {
		sendEvent(ctx, out, Event{Kind: EventError, Err: err, Budget: budget})
		return
	}
	defer resp.Body.Close()
	budget = BudgetFromHeaders(resp.Header)
	window := ContextWindowFromHeaders(resp.Header)
	idle := c.IdleTimeout
	if idle <= 0 {
		idle = defaultIdleTimeout
	}
	var stalled atomic.Bool
	timer := timeAfterFunc(idle, func() {
		stalled.Store(true)
		_ = resp.Body.Close()
	})
	events, err := readSSE(ctx, resp.Body, budget, window, func() { timer.Reset(idle) })
	timer.Stop()
	if err != nil {
		if stalled.Load() {
			err = fmt.Errorf("the server stopped sending data for %s", idle)
		}
		sendEvent(ctx, out, Event{Kind: EventError, Err: err, Budget: budget})
		return
	}
	for _, event := range events {
		sendEvent(ctx, out, event)
	}
}

func (c *Client) post(ctx context.Context, body chatRequest, allowReasoningFallback bool) (*http.Response, BudgetStatus, []byte, error) {
	if c.noReasoning.Load() {
		body.ReasoningEffort = ""
	}
	resp, budget, data, err := c.doPost(ctx, body)
	if err != nil && allowReasoningFallback && body.ReasoningEffort != "" && rejectsReasoning(data) {
		c.noReasoning.Store(true)
		body.ReasoningEffort = ""
		return c.doPost(ctx, body)
	}
	return resp, budget, data, err
}

func (c *Client) doPost(ctx context.Context, body chatRequest) (*http.Response, BudgetStatus, []byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, BudgetStatus{}, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, BudgetStatus{}, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, BudgetStatus{}, nil, ErrUnreachable{Err: err}
	}
	if resp.StatusCode == http.StatusOK {
		return resp, BudgetStatus{}, nil, nil
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, BudgetStatus{}, nil, ErrUnauthorized
	case http.StatusPaymentRequired:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, BudgetStatus{Set: true, Remaining: 0}, nil, ErrBudgetExhausted
	default:
		data, _ := io.ReadAll(resp.Body)
		return nil, BudgetStatus{}, data, fmt.Errorf("%d: %s", resp.StatusCode, errorMessage(data))
	}
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatRequest struct {
	Model           string         `json:"model"`
	Messages        []Message      `json:"messages"`
	Tools           []Tool         `json:"tools,omitempty"`
	Stream          bool           `json:"stream"`
	StreamOptions   *streamOptions `json:"stream_options,omitempty"`
	ReasoningEffort string         `json:"reasoning_effort,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta streamDelta `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
	} `json:"usage,omitempty"`
}

type streamDelta struct {
	Content   string          `json:"content,omitempty"`
	Reasoning string          `json:"reasoning,omitempty"`
	ToolCalls []toolCallDelta `json:"tool_calls,omitempty"`
}

type toolCallDelta struct {
	Index    int    `json:"index,omitempty"`
	ID       string `json:"id,omitempty"`
	Function struct {
		Name      string `json:"name,omitempty"`
		Arguments string `json:"arguments,omitempty"`
	} `json:"function"`
}

type toolSlot struct {
	id, name string
	args     strings.Builder
}

func readSSE(ctx context.Context, r io.Reader, budget BudgetStatus, contextWindow int, onFrame func()) ([]Event, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 4*1024*1024)
	var events []Event
	var content strings.Builder
	slots := map[int]*toolSlot{}
	var order []int
	tokens, promptTokens := 0, 0
	for scanner.Scan() {
		if onFrame != nil {
			onFrame()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return nil, fmt.Errorf("sse payload: %w", err)
		}
		for _, choice := range chunk.Choices {
			delta := choice.Delta
			if delta.Reasoning != "" {
				events = append(events, Event{Kind: EventReasoning, Content: delta.Reasoning, Budget: budget})
			}
			if delta.Content != "" {
				content.WriteString(delta.Content)
				events = append(events, Event{Kind: EventContent, Content: delta.Content, Budget: budget})
			}
			for _, call := range delta.ToolCalls {
				slot, ok := slots[call.Index]
				if !ok {
					slot = &toolSlot{}
					slots[call.Index] = slot
					order = append(order, call.Index)
				}
				if call.ID != "" {
					slot.id = call.ID
				}
				if call.Function.Name != "" {
					slot.name = call.Function.Name
				}
				if call.Function.Arguments != "" {
					slot.args.WriteString(call.Function.Arguments)
					events = append(events, Event{Kind: EventToolArgs, Content: call.Function.Arguments, Budget: budget})
				}
			}
		}
		if chunk.Usage != nil {
			tokens = chunk.Usage.CompletionTokens
			promptTokens = chunk.Usage.PromptTokens
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	var calls []ToolCall
	for _, idx := range order {
		calls = append(calls, slots[idx].resolve())
		events = append(events, Event{Kind: EventToolCall, Tool: calls[len(calls)-1], Budget: budget})
	}
	events = append(events, Event{
		Kind:          EventDone,
		Final:         &Message{Role: "assistant", Content: content.String(), Tools: calls},
		Budget:        budget,
		ContextWindow: contextWindow,
		Tokens:        tokens,
		PromptTokens:  promptTokens,
	})
	return events, nil
}

func (t *toolSlot) resolve() ToolCall {
	args := map[string]any{}
	raw := strings.TrimSpace(t.args.String())
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			args["_parse_error"] = err.Error()
		}
	}
	return ToolCall{ID: t.id, Name: t.name, Arguments: args}
}

func sendEvent(ctx context.Context, out chan<- Event, event Event) bool {
	select {
	case out <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

func rejectsReasoning(data []byte) bool {
	text := string(data)
	return (strings.Contains(text, "not support") && strings.Contains(text, "reasoning_effort")) ||
		strings.Contains(text, "does not support thinking")
}

func errorMessage(data []byte) string {
	var envelope struct {
		Error struct {
			Message      string `json:"message"`
			ProviderHint string `json:"provider_hint"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &envelope) == nil {
		if envelope.Error.ProviderHint != "" {
			return envelope.Error.ProviderHint
		}
		if envelope.Error.Message != "" {
			return envelope.Error.Message
		}
	}
	return firstLine(string(data))
}

func firstLine(text string) string {
	if idx := strings.IndexAny(text, "\r\n"); idx >= 0 {
		return strings.TrimSpace(text[:idx])
	}
	return strings.TrimSpace(text)
}

func estimate(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += messageTokens(msg)
	}
	return total
}

func messageTokens(msg Message) int {
	total := Tokens(msg.Role) + Tokens(msg.Content) + 8
	for _, tool := range msg.Tools {
		total += Tokens(tool.ID) + Tokens(tool.Name)
		for k, v := range tool.Arguments {
			total += Tokens(k) + Tokens(fmt.Sprint(v))
		}
	}
	total += Tokens(msg.ToolCallID) + Tokens(msg.ToolName)
	return total
}

func dropOrphanTools(messages []Message) []Message {
	seen := map[string]bool{}
	out := messages[:0]
	for _, msg := range messages {
		if msg.Role == "assistant" {
			for _, call := range msg.Tools {
				if call.ID != "" {
					seen[call.ID] = true
				}
			}
		}
		if msg.Role == "tool" && (msg.ToolCallID == "" || !seen[msg.ToolCallID]) {
			continue
		}
		out = append(out, msg)
	}
	return out
}

func dropDanglingToolCalls(messages []Message) []Message {
	answered := map[string]bool{}
	for _, msg := range messages {
		if msg.Role == "tool" && msg.ToolCallID != "" {
			answered[msg.ToolCallID] = true
		}
	}
	out := messages[:0]
	for _, msg := range messages {
		if msg.Role == "assistant" && len(msg.Tools) > 0 {
			dangling := false
			for _, call := range msg.Tools {
				if call.ID == "" || !answered[call.ID] {
					dangling = true
				}
			}
			if dangling {
				continue
			}
		}
		out = append(out, msg)
	}
	return out
}

func newestToolGroup(history []Message) []Message {
	if len(history) == 0 {
		return nil
	}
	last := history[len(history)-1]
	if last.Role != "tool" || last.ToolCallID == "" {
		return nil
	}
	owner := -1
	for i := len(history) - 2; i >= 0 && owner < 0; i-- {
		if history[i].Role != "assistant" {
			continue
		}
		for _, call := range history[i].Tools {
			if call.ID == last.ToolCallID {
				owner = i
				break
			}
		}
	}
	if owner < 0 {
		return nil
	}
	ids := map[string]bool{}
	for _, call := range history[owner].Tools {
		if call.ID != "" {
			ids[call.ID] = true
		}
	}
	group := []Message{history[owner]}
	for i := owner + 1; i < len(history); i++ {
		if history[i].Role == "tool" && ids[history[i].ToolCallID] {
			group = append(group, history[i])
		}
	}
	return group
}

func anchorUser(kept []Message, history []Message) []Message {
	for _, msg := range kept {
		if msg.Role == "user" {
			return kept
		}
	}
	for _, msg := range history {
		if msg.Role == "user" {
			return append([]Message{msg}, kept...)
		}
	}
	return kept
}

func demoteSystem(messages []Message) []Message {
	for i := range messages {
		if messages[i].Role == "system" {
			messages[i].Role = "user"
		}
	}
	return messages
}

func cloneMessages(messages []Message) []Message {
	out := append([]Message(nil), messages...)
	for i := range out {
		out[i].Tools = append([]ToolCall(nil), out[i].Tools...)
	}
	return out
}

func runeBoundaryDown(text string, idx int) int {
	if idx >= len(text) {
		return len(text)
	}
	for idx > 0 && !utf8.RuneStart(text[idx]) {
		idx--
	}
	return idx
}

func runeBoundaryUp(text string, idx int) int {
	if idx <= 0 {
		return 0
	}
	for idx < len(text) && !utf8.RuneStart(text[idx]) {
		idx++
	}
	return idx
}

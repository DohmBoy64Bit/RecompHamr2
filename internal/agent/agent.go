package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"recomphamr2/internal/llm"
)

const (
	defaultMaxModelRounds = 32
	defaultMaxToolCalls   = 75
	failureNudgeStreak    = 5
	verifyNudgeMinTools   = 8
	nudgeOrigin           = "[Automated recomphamr check - not a message from your user.] "
)

// Model produces one assistant message for a transcript.
type Model interface {
	Next(context.Context, []llm.Message) (llm.Message, error)
}

// ToolRunner executes a tool call.
type ToolRunner func(context.Context, llm.ToolCall) (string, error)

// Loop owns model-tool turn execution.
type Loop struct {
	// Model is the chat backend used for each model round.
	Model Model
	// RunTool dispatches one model-requested tool call.
	RunTool ToolRunner
	// ProjectMemory is optional persistent workspace context for this turn.
	ProjectMemory string
	// ProjectMemorySource identifies where ProjectMemory was loaded from.
	ProjectMemorySource string
	// ProjectMemoryMaxTokens caps ProjectMemory before model calls.
	ProjectMemoryMaxTokens int
	// MaxRounds caps model requests in one user turn.
	MaxRounds int
	// MaxToolCalls controls when the runaway-tool nudge fires.
	MaxToolCalls int
}

// Run executes a turn until the model emits no tool calls.
func (l Loop) Run(ctx context.Context, history []llm.Message) ([]llm.Message, error) {
	if l.Model == nil || l.RunTool == nil {
		return nil, fmt.Errorf("agent loop is not configured")
	}
	max := l.MaxRounds
	if max <= 0 {
		max = defaultMaxModelRounds
	}
	maxTools := l.MaxToolCalls
	if maxTools <= 0 {
		maxTools = defaultMaxToolCalls
	}
	out := cloneMessages(history)
	if strings.TrimSpace(l.ProjectMemory) != "" {
		out = llm.WithProjectMemory(out, l.ProjectMemorySource, l.ProjectMemory, l.ProjectMemoryMaxTokens)
	}
	state := turnState{}
	for round := 0; round < max; round++ {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		msg, err := l.Model.Next(ctx, out)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return out, ctxErr
			}
			return out, err
		}
		out = append(out, msg)
		if len(msg.Tools) == 0 {
			if newestAssistantEmpty(out) {
				if state.emptyNudged {
					return out, fmt.Errorf("blocked: model ended with no reply and no tool call after retry")
				}
				state.emptyNudged = true
				out = append(out, systemNudge("Your last turn ended with no reply and no tool call. If you meant to call a tool and it did not run, issue it again now as a proper tool call. If you are still working, continue. If the task is done, check it against the original request and reply with a one-line summary."))
				continue
			}
			if state.shouldVerifyNudge(out) {
				state.verifyNudged = true
				out = append(out, systemNudge("Before you finish: re-read the original request and walk its acceptance criteria one at a time. For each, name the check you actually ran and what it showed. Anything runnable you built or changed is proven only by running it. If a check cannot run here, mark it `unverified`, `unsupported`, or `blocked` with the exact reason. Then reply with your one-line summary."))
				continue
			}
			return out, nil
		}
		state.emptyNudged = false
		for _, call := range msg.Tools {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			state.toolCalls++
			state.lastToolKey = toolTargetKey(call)
			result, err := l.RunTool(ctx, call)
			if err != nil {
				if ctxErr := contextError(ctx, err); ctxErr != nil {
					out = append(out, toolMessage(call, "(cancelled)"))
					return out, ctxErr
				}
				result = "blocked: " + err.Error()
			}
			out = append(out, toolMessage(call, result))
			state.recordToolOutcome(call.Name, result)
		}
		if note := state.failureNudge(); note != "" {
			out = append(out, systemNudge(note))
		}
		if note := state.runawayNudge(maxTools); note != "" {
			out = append(out, systemNudge(note))
		}
	}
	return out, fmt.Errorf("blocked: exceeded %d model rounds", max)
}

type turnState struct {
	toolCalls     int
	failKey       string
	failStreak    int
	lastToolKey   string
	emptyNudged   bool
	runawayNudged bool
	verifyNudged  bool
}

func (s *turnState) recordToolOutcome(name string, result string) {
	if !toolResultFailed(name, result) {
		s.failKey, s.failStreak = "", 0
		return
	}
	if s.lastToolKey != "" && s.lastToolKey == s.failKey {
		s.failStreak++
		return
	}
	s.failKey = s.lastToolKey
	s.failStreak = 1
}

func (s *turnState) failureNudge() string {
	if s.failStreak < failureNudgeStreak {
		return ""
	}
	note := fmt.Sprintf("The last %d tool calls to the same target failed the same way. Stop repeating it; read the error, change approach, or tell the user what is blocking you.", s.failStreak)
	s.failKey, s.failStreak = "", 0
	return note
}

func (s *turnState) runawayNudge(maxTools int) string {
	if s.runawayNudged || s.toolCalls < maxTools {
		return ""
	}
	s.runawayNudged = true
	return fmt.Sprintf("%d tool calls so far this turn without finishing. If you are still making real progress, keep going. If you are repeating a step that cannot work here, verify another way or tell the user where things stand.", s.toolCalls)
}

func (s *turnState) shouldVerifyNudge(history []llm.Message) bool {
	return !s.verifyNudged && s.toolCalls >= verifyNudgeMinTools && !newestAssistantEvidenceLabeled(history)
}

func systemNudge(content string) llm.Message {
	return llm.Message{Role: "system", Content: nudgeOrigin + content}
}

func toolMessage(call llm.ToolCall, result string) llm.Message {
	return llm.Message{Role: "tool", Content: result, ToolCallID: call.ID, ToolName: call.Name}
}

func contextError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

func toolTargetKey(call llm.ToolCall) string {
	switch call.Name {
	case "write_file", "edit_file", "read_file":
		return call.Name + "|" + argString(call, "path")
	case "bash", "powershell":
		cmd := argString(call, "cmd")
		if i := strings.IndexByte(cmd, '\n'); i >= 0 {
			cmd = cmd[:i]
		}
		return call.Name + "|" + strings.TrimSpace(cmd)
	default:
		return call.Name
	}
}

func argString(call llm.ToolCall, key string) string {
	if call.Arguments == nil {
		return ""
	}
	value, _ := call.Arguments[key].(string)
	return value
}

func toolResultFailed(name string, result string) bool {
	if strings.Contains(result, "(cancelled)") {
		return false
	}
	trimmed := strings.TrimSpace(result)
	if strings.HasPrefix(trimmed, "blocked:") || strings.HasPrefix(trimmed, "(tool arguments were not valid JSON") || strings.HasPrefix(trimmed, "(unknown tool:") {
		return true
	}
	switch name {
	case "write_file":
		return strings.HasPrefix(trimmed, "(write_file:")
	case "edit_file":
		return strings.HasPrefix(trimmed, "(edit_file:")
	case "read_file":
		return strings.HasPrefix(trimmed, "(read_file:")
	case "bash", "powershell":
		return strings.Contains(trimmed, "(exit:") || strings.Contains(trimmed, "(timeout after ")
	default:
		return strings.HasPrefix(trimmed, "(")
	}
}

func newestAssistantEmpty(history []llm.Message) bool {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "assistant" {
			return strings.TrimSpace(history[i].Content) == "" && len(history[i].Tools) == 0
		}
	}
	return false
}

func newestAssistantEvidenceLabeled(history []llm.Message) bool {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role != "assistant" {
			continue
		}
		text := strings.ToLower(history[i].Content)
		return strings.Contains(text, "unverified") || strings.Contains(text, "unsupported") || strings.Contains(text, "blocked")
	}
	return false
}

func cloneMessages(messages []llm.Message) []llm.Message {
	out := append([]llm.Message(nil), messages...)
	for i := range out {
		out[i].Tools = append([]llm.ToolCall(nil), out[i].Tools...)
	}
	return out
}

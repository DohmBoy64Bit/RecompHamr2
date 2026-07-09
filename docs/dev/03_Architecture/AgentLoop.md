# Agent Loop

`internal/agent` owns the model-tool turn loop independently from the TUI. UI
code may render transcript state and cancellation controls, but it must not
duplicate tool dispatch, retry, or nudge policy.

## Turn Policy

- `Loop.Run` copies the input transcript and returns the updated transcript plus
  an explicit error when the turn is blocked or cancelled.
- `MaxRounds` caps model requests in one user turn. The default is 32 model
  rounds.
- `MaxToolCalls` controls the runaway-tool nudge. The default threshold is 75
  tool calls.
- Each tool result is appended as a `tool` message with the original
  `tool_call_id` and tool name so OpenAI-compatible transcripts remain paired.

## Nudges

- Repeated failure: five consecutive failures to the same target append one
  system nudge, then reset the streak.
- Runaway tools: crossing the tool-call threshold appends one self-check nudge
  per turn.
- Empty reply: one empty assistant reply with no tool calls gets one retry
  nudge; a second empty reply returns `blocked`.
- Verification: after at least eight tool calls, a clean non-empty finish gets
  one acceptance-criteria nudge unless the assistant already labels a limit as
  `unverified`, `unsupported`, or `blocked`.

## Cancellation

The loop checks the parent context before model calls, before tool dispatch, and
after tool runner errors. If cancellation wins, the transcript records
`(cancelled)` for an in-flight tool and returns the context error instead of
claiming success.

# Context Packing

Context packing lives in `internal/llm`.

Implemented Phase 4 behavior:

- `Tokens(text)` estimates prompt cost as bytes divided by four, rounded up.
- `Budget(context_size)` reserves fixed system/tool/response room and leaves a 10% headroom margin.
- `Pack(messages, maxTokens)` trims oldest history, keeps the newest valid content, anchors a user task when history has one, demotes secondary system notes to user notes, drops orphan tool messages, and drops dangling assistant tool calls.
- `TruncateToolOutput(text, maxTokens)` keeps head and tail text with a truncation marker and preserves UTF-8 boundaries.
- `WithProjectMemory(messages, source, memory, maxTokens)` clones transcript history,
  injects `.rehamr/REPHAMR_STATE.md` into the primary system context, labels the
  source path, and trims the memory body by token budget.

Invalid OpenAI wire shapes are stripped before sending:

- tool messages without a preceding assistant `tool_calls` entry;
- assistant tool calls without matching tool responses;
- empty tool-call IDs;
- secondary `system` messages after the embedded system prompt.

Reference evidence: RecompHamr 1.x `internal/ctx/ctx.go` and `internal/ctx/ctx_test.go`.

Memory file I/O stays in `internal/project`; context packing never reads the
filesystem. Runtime callers load memory first, then pass text into
`WithProjectMemory` or `agent.Loop`.

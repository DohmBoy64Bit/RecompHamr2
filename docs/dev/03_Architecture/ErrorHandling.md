# Error Handling

Errors must be explicit, actionable, and evidence-preserving. Use `unsupported`, `unverified`, or `blocked` where appropriate. Do not hide unsupported behavior behind success output.

Phase 4 LLM error classes:

- `ErrUnauthorized`: HTTP 401, invalid or expired provider token.
- `ErrBudgetExhausted`: HTTP 402, provider budget/pass depleted.
- `ErrUnreachable`: transport, DNS, connection, or timeout failure before a response.
- non-2xx provider errors: include status code and the provider message or hint when the body has an OpenAI-style error envelope.
- SSE parse errors: returned as stream errors with `sse payload` context.
- idle stalls: returned when no SSE frame arrives within `RECOMPHAMR_IDLE_TIMEOUT`.

Reasoning fallback is not silent success. It is limited to provider errors that
explicitly reject `reasoning_effort` or report that the model does not support
thinking; unrelated 400 responses remain errors.

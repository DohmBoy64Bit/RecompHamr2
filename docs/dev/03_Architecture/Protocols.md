# Protocols

JSON is used for protocol payloads. YAML is reserved for user configuration.

## OpenAI-Compatible Chat

`internal/llm.Client` posts to:

```text
<base_url>/v1/chat/completions
```

Requests include `model`, `messages`, optional `tools`, `stream: true`,
`stream_options.include_usage: true`, and `reasoning_effort: high` unless the
provider rejects that field. If the provider reports a reasoning incompatibility,
the client retries once without `reasoning_effort` and keeps that fallback for
the client lifetime.

## SSE

The SSE parser accepts `data:` frames, ignores blank lines and keepalive
comments, emits content deltas immediately, emits reasoning deltas as
UI-only events, streams tool argument fragments as UI-only events, and emits
assembled tool calls at stream end. Tool calls are assembled by provider
`index`, so interleaved calls remain distinct.

Headers:

- `X-Budget-Remaining`: optional provider budget fraction, clamped to `[0,1]`.
- `X-Context-Window`: optional provider context window, accepted only from `1024` through `8388608`.

`RECOMPHAMR_IDLE_TIMEOUT` configures the inter-frame idle timeout. It accepts a
Go duration such as `90m` or a bare second count such as `300`.

## MCP JSON-RPC

`internal/mcp` implements MCP JSON-RPC 2.0 protocol foundations.

Supported method constants:

- `initialize`
- `notifications/initialized`
- `tools/list`
- `tools/call`

Stdio transport uses newline-delimited JSON over an injected stream. Streamable
HTTP transport posts JSON-RPC to:

```text
<base_url>/mcp
```

HTTP requests set `Content-Type: application/json` and
`Accept: application/json, text/event-stream`. Phase 10 supports JSON
request/response handling and notification posts; long-lived HTTP event stream
session management remains reserved for the MCP manager/runtime phase.

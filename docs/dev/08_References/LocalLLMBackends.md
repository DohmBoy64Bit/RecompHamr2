# Local LLM Backends

Phase 4 implements a single OpenAI-compatible chat-completions path for local
and hosted endpoints:

```text
POST <base_url>/v1/chat/completions
Accept: text/event-stream
```

Supported local backends must expose the OpenAI-compatible `/v1` chat shape.
Native provider-specific APIs are `unsupported`; adapters should emit standard
OpenAI request and SSE response shapes instead.

Backend notes:

- LM Studio: default profiles use `http://localhost:1234`.
- Ollama OpenAI shim: default profile uses `http://localhost:11434`.
- llama.cpp server mode: default profile uses `http://localhost:8080`.

No silent local-to-cloud fallback is allowed. A failed local endpoint reports
`ErrUnreachable` through the LLM event stream; changing to a cloud endpoint
requires an explicit profile or `RECOMPHAMR_URL` override.

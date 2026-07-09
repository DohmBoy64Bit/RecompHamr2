# Integration Tests

Integration tests are required for LLM streaming, tool calls, MCP lifecycle, config bootstrap, memory, TUI command dispatch, and install scripts when those phases exist.

Next 20 Phase 15 adds deterministic fake-runtime smoke tests in
`internal/app`. These tests cover slash-command dispatch, first prompt
execution, fake model-tool loop, cancellation, memory injection, and transcript
rendering without real network, model backend, tool execution, MCP autoconnect,
or terminal process dependencies.

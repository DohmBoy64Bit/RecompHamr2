# Decision Log

- Use Go for the RecompHamr 2.0 agent shell.
- Use `Makefile` and `make verify` as the canonical local verification entrypoint.
- Keep product runtime diagnostic-only until parity evidence exists.
- Superseded on Next 20 Phase 14: bare startup may compose local runtime state
  after config, memory, command, MCP, and TUI package parity evidence exists.
  Startup still must not call a model backend, execute tools, or autoconnect MCP
  servers.
- Model TUI interaction quality after terminal-first agents such as OpenCode, but use a unique RecompHamr initiative layout with an evidence rail and context deck to support reverse-engineering work.

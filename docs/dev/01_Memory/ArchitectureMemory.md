# Architecture Memory

The TUI must render state and dispatch intent only. Agent loop, tools, LLM, MCP, config, security, and docs checks must be independently testable without a terminal.

The TUI layout is RecompHamr-specific: header, initiative rail, transcript, evidence deck, and composer. It is opencode-inspired in terminal-first workflow only, not copied 1:1.

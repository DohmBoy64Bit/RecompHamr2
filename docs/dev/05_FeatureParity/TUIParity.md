# TUI Parity

TUI parity requires transcript rendering, multiline prompt, slash palette, completions, footer, skill and MCP indicators, streaming tool status, paste handling, cancellation, quit behavior, and redacted debug mode.

Current foundation:

- Implemented a testable initiative layout renderer in `internal/tui`.
- Implemented wide and compact rendering contracts.
- Implemented a Bubble Tea adapter around the pure TUI model.
- Implemented multiline composer state, key handling, resize handling, prompt
  history, paste chips, cancellation, quit behavior, and redacted debug output.
- Implemented slash command completion.
- Kept core agent/tool/config/MCP behavior outside the TUI.

Known limit:

- Product runtime wiring composes pure TUI state and reports `ready` at bare
  startup.
- Deterministic fake-runtime smoke renders slash-command and prompt transcripts
  through the pure TUI model.
- The live Bubble Tea process loop remains unsupported.

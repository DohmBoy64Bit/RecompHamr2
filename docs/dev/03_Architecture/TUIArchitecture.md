# TUI Architecture

The TUI renders application state and dispatches user intent. It must not contain the core agent loop, tool execution, config persistence, or MCP lifecycle logic.

## Initiative Layout

The RecompHamr TUI is inspired by the terminal-first workflow quality of OpenCode, but it must not copy OpenCode 1:1. RecompHamr uses its own evidence-first "initiative" layout:

- Header: current mode and active model.
- Initiative rail: memory freshness, active skill, MCP gate state, and pending tool.
- Transcript: user, assistant, command, and tool-visible conversation lines.
- Evidence deck: context budget and verified working state separated from chat prose.
- Composer: prompt entry area.

Compact terminals collapse the rail and evidence deck into a status band. This improves narrow-terminal usability while preserving the same state.

## Phase 7 Shell Contract

`internal/tui` now provides two layers:

- A pure `Model` state with deterministic update events for prompt text,
  multiline composer rendering, slash completion, paste chips, prompt history,
  resize handling, cancellation, quit, status text, and redacted debug lines.
- A thin Bubble Tea `BubbleModel` adapter that translates Bubble Tea key and
  window-size messages into the pure model.

The TUI may dispatch slash commands through `internal/commands` and redact debug
text through `internal/security`. It must not execute tools, own the agent loop,
persist config, or manage MCP lifecycles. Product executable wiring composes the
pure TUI state in `internal/app`; Phase 15 fake-runtime smoke renders the pure
model through injected dependencies, while the interactive Bubble Tea process
loop remains outside the current runtime.

## Key Behavior

- `Enter` submits the composer.
- `Up` and `Down` navigate prompt history.
- `Ctrl+C` cancels active thinking/streaming/tool status; while idle it arms
  quit, and a second `Ctrl+C` quits.
- `Ctrl+D` quits immediately.
- `Esc` clears transient quit/status state.
- Large or multiline paste text becomes a named paste chip instead of flooding
  the composer.

## Improvement Rationale

The evidence rail is a deliberate improvement for reverse-engineering work because it keeps the facts that prevent drift visible at all times. The evidence deck separates verified context from conversational text, which makes unsupported claims easier to spot. Compact mode keeps the interface usable in terminals below the wide layout threshold without adding a second UI product.

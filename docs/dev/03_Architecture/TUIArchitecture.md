# TUI Architecture

The TUI renders application state and dispatches user intent. It must not contain the core agent loop, tool execution, config persistence, or MCP lifecycle logic.

## Initiative Layout

The RecompHamr TUI is inspired by the terminal-first workflow quality of
OpenCode, but it must not copy OpenCode 1:1. RecompHamr uses its own
evidence-first terminal layout:

- Header: brand, domain line, current mode, active model, and safety line.
- Signals band: memory freshness, active skill, MCP gate state, pending tool,
  and context state.
- Transcript: user, assistant, command, and tool-visible conversation lines.
- Evidence column: context budget and verified working state separated from
  chat prose.
- Composer: prompt entry area.

Compact terminals collapse signals and evidence into status chips. This
improves narrow-terminal usability while preserving the same state.

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
model through injected dependencies, and Phase 28 live runtime wiring launches
the Bubble Tea process from `internal/app` while keeping side effects outside
`internal/tui`.

## Live Runtime Boundary

`internal/app` owns the live Bubble Tea wrapper. It observes submit, cancel, and
quit actions from `tui.BubbleModel`, then starts a cancellable `internal/agent`
turn for plain prompts. Slash commands remain handled by `internal/tui` through
`internal/commands`. The wrapper updates status, appends assistant/tool lines,
and returns `tea.Quit` for clean exits; it does not change the pure TUI contract.

## Key Behavior

- `Enter` submits the composer.
- `/` opens or filters the slash command palette.
- `Tab` completes the first matching slash command from the registry.
- `Up` and `Down` navigate prompt history.
- `Ctrl+C` cancels active thinking/streaming/tool status; while idle it arms
  quit, and a second `Ctrl+C` quits.
- `Ctrl+D` quits immediately.
- `Esc` clears transient quit/status state.
- Large or multiline paste text becomes a named paste chip instead of flooding
  the composer.
- Palette rows are generated from `internal/commands.Registry` and include the
  command summary and usage string.

## Transcript Blocks

Transcript lines are classified at render time without changing agent or command
semantics:

- `user` for submitted prompts;
- `assistant` for model replies;
- `tool` for built-in tool results;
- `mcp` for MCP status or tool output;
- `blocked`, `unsupported`, and `unverified` for evidence labels;
- `status` for runtime status lines;
- `attachment` for paste chips;
- `note` for remaining command or informational output.

The renderer preserves the original text after the block label. It must not add
fake timing, token, cost, or reasoning metrics.

## Improvement Rationale

The signals band is a deliberate improvement for reverse-engineering work
because it keeps the facts that prevent drift visible at all times. The evidence
column separates verified context from conversational text, which makes
unsupported claims easier to spot. Compact mode keeps the interface usable in
terminals below the wide layout threshold without adding a second UI product.

## Corrective TUI Hardening Direction

Phases 30-33 flesh out the end-user TUI before post-parity feature intake. The
reference direction is terminal-first polish similar in quality to OpenCode, but
the implementation must be RecompHamr-specific and must not copy OpenCode or
RecompHamr 1.x 1:1.

Required RecompHamr-specific outcomes:

- branded startup state with `RECOMP HAMR` wide branding and compact
  `RecompHamr` branding;
- dark terminal visual system using RecompHamr-owned text tokens before ANSI
  color is layered in later;
- persistent model, memory, skill, MCP, context, tool, and permission status;
- registry-driven slash command palette and completion;
- professional transcript blocks for assistant, user, command, tool, MCP,
  blocked, and unsupported messages;
- no fake token, cost, timing, or reasoning data;
- golden renders for every major state and responsive breakpoint.

OpenCode public docs and user screenshots may inform broad interaction patterns
such as terminal-based workflow, footer hints, command palettes, and agent mode
visibility. They are not source truth for implementation details, styling, or
copy. RecompHamr source truth remains local code, local docs, parity rows, and
verified runtime output.

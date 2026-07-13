# TUI Architecture

## Phase 47 Component Architecture

The corrective architecture uses Bubble Tea v2 as the runtime and Bubbles
v2.1.1 for interaction mechanics. `bubbleComponents` owns a Bubbles textarea,
transcript viewport, help model, and declarative key bindings. RecompHamr still
owns the layout, semantic theme, palette rows, transcript classification, and
runtime state.

`BubbleModel` emits a typed `Intent` alongside the compatibility `Action`.
`internal/app` consumes submit, cancel, and quit intent kinds and remains the
only owner of model, command, tool, MCP, config, filesystem, and network side
effects. Future model, skill, MCP, and command picker intents are defined for
Phase 50 but are not emitted speculatively.

Component state initializes both through `NewBubble` and lazily for zero-value
test/app adapters. State synchronization configures textarea width, viewport
size/content, and help width from the current layout. Display measurement uses
Lip Gloss width and ANSI grapheme-aware truncation rather than byte length.

The user-facing interface is intentionally unchanged in Phase 47; phases
48-52 replace behavior and visual composition against `TUIUXSpec.md`.

## Phase 48 Runtime Input Contract

`BubbleModel.Update` routes supported terminal messages through the Bubbles
textarea and viewport while keeping app effects behind typed intents. Printable
keys, deletion, cursor movement, focus, blur, and paste update the textarea;
`Shift+Enter` or `Ctrl+J` inserts a newline and plain `Enter` submits. Page keys
and mouse-wheel messages scroll the transcript viewport when transcript content
exists. Startup does not request mouse reporting; chat requests cell-motion
reporting so wheel input is available without making mouse input mandatory.

Window-size messages synchronize the composer and transcript component sizes.
Terminals smaller than 60 columns or 18 rows render a centered size requirement,
hide the cursor, and retain `Ctrl+D` exit. Declarative `tea.View` fields continue
to own alt-screen, focus, mouse, title, and cursor behavior.

## Phase 49 Startup Contract

The startup renderer implements five groups: two-tone identity, responsive
domain line, shallow three-row composer, three binding hints, and an optional
actionable tip. Its footer contains only mode, active model, and `ready` or
`working`; secondary memory, skill, MCP, context, permission, and tool details
are not duplicated on the empty screen.

Wide and 80-column layouts use `RECOMP HAMR`; 60-column layout uses compact
`RecompHamr` and `RE / decomp / recomp`. `/init-re` guidance appears only when
memory state reports `missing` or `unsupported`. Version and working-directory
corners remain omitted until app-owned state supplies verified values. Styled
and plain startup renderers share these hierarchy and evidence rules.

## Phase 50 Palette And Picker Contract

Slash rows contain command name and summary; selected usage and side effects
render once in the footer. Results are bounded to seven rows, or five below 80
columns, and scroll around selection in the command lane. Blocked picker rows
are promoted into view. Arrows and `j`/`k` navigate. Model, skill, and MCP Enter
acceptance emits typed intents consumed by `internal/app`; blocked and empty
rows emit none.

## Phase 51 Transcript And Runtime Feedback

Transcript rendering uses one offset-from-bottom shared by pure renders and
Bubble Tea page/wheel input. Offset zero follows output. Appending while
scrolled preserves the visible position and shows `new output  PgDn to follow`;
returning to zero clears it. All transcript additions use one append path.

Tool and MCP multiline output is limited to 12 rendered lines, display-width
truncated, and explicitly ends with `output truncated`. User, assistant, tool,
MCP, verification, warning, blocked, unsupported, unverified, attachment,
status, and note classes remain text-labeled and display-time redacted. The
fixed composer reports only mode, model, and ready/working state; no cost,
token, timing, or reasoning data is invented.

## Phase 52 Responsive Theme And Accessibility

Each render owns an immutable `tuiTheme` selected from Bubble Tea's
`ColorProfileMsg`. Lip Gloss `Complete` maps semantic roles to ANSI16,
ANSI256, or truecolor values; ASCII and `NO_COLOR` use `NoColor`. No global
renderer profile is mutated. Selection retains reverse video, focused input
retains a left rail, and warning/blocked states retain explicit labels, so
meaning never depends on color.

The acceptance matrix covers 120x32, 96x28, 80x24, 60x20, and 60x18 across all
four profiles. Every stripped line must fit the viewport. Tests cover CJK,
combining marks, long model names, long composers, bounded startup/chat cursor
positions, and the 59x17 too-small state. `color_profiles.golden` records the
profile degradation contract.

The TUI renders application state and dispatches user intent. It must not contain the core agent loop, tool execution, config persistence, or MCP lifecycle logic.

## Initiative Layout

The RecompHamr TUI is inspired by the terminal-first workflow quality of
OpenCode, but it must not copy OpenCode 1:1. RecompHamr uses its own
evidence-first terminal layout:

- Centered launcher: brand, domain line, prompt panel, model/mode/status line,
  key hints, and setup tip.
- Transcript-first chat: recent user, assistant, tool, MCP, blocked,
  unsupported, unverified, status, attachment, and note lines occupy the main
  surface without a permanent debug board.
- Floating command palette: registry-backed slash command rows render as an
  overlay above the launcher or chat surface.
- Bottom composer: multiline prompt entry and model/skill/MCP/context status
  stay in the bottom command area.

Compact terminals keep the same launcher, transcript, palette, composer, and
status concepts with reduced width and truncated transcript cards where needed.

## Phase 7 Shell Contract

`internal/tui` now provides focused layers:

- A pure `Model` state with deterministic update events for prompt text,
  multiline composer rendering, slash completion, paste chips, prompt history,
  resize handling, cancellation, quit, status text, and redacted debug lines.
- A thin Bubble Tea `BubbleModel` adapter that translates Bubble Tea key and
  window-size messages into the pure model.
- Component-oriented files for constants, types, model updates, Bubble Tea
  messages, composer helpers, layout helpers, plain render, Bubble Tea render,
  palette, status, transcript classification, and Lip Gloss styles.

Bubble Tea work must start by reading the documentation for the active module
version. The current implementation uses `charm.land/bubbletea/v2` and
`charm.land/lipgloss/v2`. Bubble Tea v2 requires `View() tea.View`,
`tea.KeyPressMsg` for key presses, `tea.PasteMsg` for bracketed paste, and
declarative `tea.View` fields for terminal behavior. The live TUI sets
`AltScreen`, `MouseModeCellMotion`, focus reporting, window title, and cursor
shape in `View()` rather than through imperative startup commands.

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

User screenshot review later rejected the Phase 30-34 layout as unacceptable.
Phases 37-44 supersede that implementation with a full Bubble Tea v2 rewrite
track while preserving the old roadmap for later use. The rewrite must read the
current Bubble Tea v2 docs before every TUI task, use Bubble Tea's model/update/
view contract directly, keep `internal/tui` side-effect free, and close only
after screenshot verification.

Required RecompHamr-specific outcomes:

- branded startup state with `RECOMP HAMR` wide branding and compact
  `RecompHamr` branding;
- dark terminal visual system using RecompHamr-owned hammer orange, cyan,
  soft green, warning yellow, blocked red, and neutral gray roles;
- centered startup launcher, transcript-first chat surface, floating command
  palette, bottom composer/status panel, and compact responsive rendering;
- Bubble Tea v2 styled view content with Lip Gloss color roles for logo,
  launcher, composer panel, palette overlay, selected command, assistant,
  tool/MCP, blocked, warning, muted, hint, and tip states;
- persistent model, memory, skill, MCP, context, tool, and permission status;
- registry-driven slash command palette and completion;
- professional transcript blocks for assistant, user, command, tool, MCP,
  blocked, and unsupported messages;
- no fake token, cost, timing, or reasoning data;
- golden renders for every major state and responsive breakpoint.

Phase 38 design audit adds a stricter two-lane contract:

- the transcript lane is append-only visible work history;
- the command lane is the mutable composer, palette, modal, footer, and status
  surface;
- startup is a no-transcript launcher state;
- chat mode begins only after a user-visible transcript entry;
- model, skill, MCP, help, and blocked states must have explicit modal or card
  contracts backed by local state.

Phase 39 resets the code shape before visual rebuilding. The previous single
large `internal/tui/tui.go` renderer is replaced by focused files inside the
same package. This is intentionally behavior-preserving: app-facing exported
types and methods remain stable while Phase 40-44 rebuild the live shell and
screen contracts on top of clearer ownership.

Phase 40 rebuilds the Bubble Tea shell boundary around a single styled screen
pass. `BubbleModel.View()` asks the pure model for a `styledScreen` containing
the rendered content and cursor coordinates, then returns a declarative
`tea.View` with alt screen, mouse mode, focus reporting, window title, and
cursor fields set from that pass. This removes the prior split where visible
content came from one renderer and cursor position came from plain-render line
counting. `tea.KeyPressMsg`, `tea.PasteMsg`, and Bubble Tea window-size messages
remain the only terminal input types translated by the adapter.

Phase 41 polishes the startup and composer surface without adding runtime side
effects. The styled startup now selects wide or compact branding from the active
terminal width, shows the domain and local-permission safety line above the
composer, includes memory in the status row, repeats the permission boundary
inside the composer panel, and places the Bubble Tea cursor from the same styled
startup layout. This keeps bare launch as a no-transcript state until a
user-visible prompt or command is submitted.

Phase 42 adds command and modal overlays while keeping ownership boundaries.
`internal/commands` exposes read-only picker rows for model profiles, skill
state, and MCP state because that package already owns config, skills, and MCP
coordination. `internal/tui` renders those snapshots through the same overlay
component used by the slash command palette. Up/Down changes only
`Model.PaletteIndex`; Tab completes the selected slash command; model, skill,
MCP, and help overlays do not mutate config, activate skills, connect MCP
servers, or spawn processes during rendering.

Phase 43 completes transcript and runtime-state rendering for the corrective
track. Transcript lines for user, assistant, tool, MCP, verification, blocked,
unsupported, unverified, attachment, status, and note states are classified at
render time, and display-time redaction is applied to transcript, status, and
debug output without mutating stored transcript text. The renderer does not
invent timing, token, cost, or private reasoning data; those fields may appear
only when verified upstream state provides them in a documented future change.

OpenCode public docs and user screenshots may inform broad interaction patterns
such as terminal-based workflow, footer hints, command palettes, and agent mode
visibility. They are not source truth for implementation details, styling, or
copy. RecompHamr source truth remains local code, local docs, parity rows, and
verified runtime output.

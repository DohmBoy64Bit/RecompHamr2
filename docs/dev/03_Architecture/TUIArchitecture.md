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

## Phase 53 Layout Correction

The styled renderer owns one measured frame. Chat reserves fixed rows for the
composer, actionable feedback, and key help, then assigns remaining rows to the
transcript and any open overlay. A palette or picker is inserted immediately
above the composer and reduces transcript height; it is never appended beyond
the terminal frame.

Cursor coordinates are derived from the final ANSI-stripped rendered frame so
responsive spacing, wrapped input, and overlays cannot drift from a separate
cursor calculation. Content is bounded using terminal display width before the
frame is padded to the active window dimensions. Routine progress remains
command-lane feedback rather than append-only transcript content.

## Phase 55 Replacement Message Flow

The replacement top-level Bubble Tea model owns official Bubbles models rather
than synchronizing them with a second domain model. UI gestures return commands
that emit typed intent messages. `internal/app` receives those messages, invokes
the existing backend owner, and sends immutable snapshot or transcript messages
back through Bubble Tea Update. View reads component state and snapshots only.

This contract removes mutable intent polling, parallel editor and viewport
state, composer-derived modal state, and rendered-output cursor reconstruction.
Custom TUI code may classify semantic transcript records and compose styled
regions, but it may not reproduce editing, scrolling, filtering, selection,
focus, or cursor mechanics supplied by Bubbles.

## Phase 56 Implemented Architecture

The previous TUI source files and golden fixtures were removed. The replacement
`tui.Model` directly owns one Bubbles textarea, viewport, list, help model, and
key map. App-owned state is carried by `tui.Snapshot`; semantic output arrives
through `tui.TranscriptMsg`; user actions leave through `tui.IntentMsg` commands.
There is no `BubbleModel`, pure composer string, `LastAction`, `LastIntent`,
custom transcript offset, or second renderer.

`internal/app.liveModel` receives intent messages in its Bubble Tea Update loop,
executes existing backend owners exactly there, and feeds results back through
snapshot or transcript messages. `Runtime.TUI` now stores an immutable startup
snapshot rather than mutable widget state. The official Bubbles list dependency
adds `github.com/sahilm/fuzzy` transitively for component-owned filtering.

## Phase 57 Authoritative Composer

The textarea value remains empty while its placeholder is visible. Normal
typing, Unicode text, Backspace/Delete, movement, paste, wrapping, focus, cursor,
and reset behavior stays in Bubbles textarea. Shift+Enter and Ctrl+J use the
textarea insertion API for portable multiline input. History restores values
only through textarea methods.

Typing bare `/` inserts one real slash and opens the command list. Backspace on
an empty command filter closes that list and resets the textarea, so the slash is
immediately removable. Enter emits no intent for empty or bare-slash input.

## Phase 58 Authoritative Transcript

The Bubbles viewport is the sole transcript scroll and follow owner. Appends
inspect `viewport.AtBottom`: output follows while the reader is at the bottom,
but preserves the viewport offset and raises a `new output  PgDn to follow`
notice while the reader is reviewing earlier content. Returning to the bottom
clears that notice. Clearing the transcript resets both content and feedback.

App messages enter as semantic `TranscriptEntry` blocks. Their bodies are
normalized once, secrets are redacted before storage and rendering, and visible
labels are padded before styling so ANSI sequences cannot corrupt alignment.
Tool and MCP blocks are bounded to eleven content lines plus an explicit
`output truncated` line. The viewport receives the final wrapped content and no
parallel transcript offset or follow flag exists outside this model.

## Phase 59 Authoritative Palettes And Pickers

One Bubbles list owns filtering, selection, pagination, navigation, and empty
states for commands, models, skills, MCP servers, and help. The shell forwards
both key messages and asynchronous `list.FilterMatchesMsg` results to the active
list; it does not maintain a parallel query or selected-row index. Arrow keys
work while filtering, and arrows plus `j`/`k` work in browsing mode.

The command registry is the sole command-row source. Enter on a complete
no-argument command emits one typed command intent. `/skill-audit` and
`/skill-new` populate the authoritative textarea for their required argument.
Model, skill, and MCP rows emit their typed selection intent once. Help rows
populate the selected command without executing it. Empty and blocked rows emit
no intent, and Escape closes an overlay without modifying composer text.

## Phase 60 Responsive Layout And Theme

The replacement uses one measured content lane and one final canvas. Standard
widths render an original five-row `RECOMP HAMR` block wordmark; widths below 80
columns use a compact literal wordmark. Both include a literal product/domain
line for monochrome output and assistive terminal readers. The startup composer
and active-session command lane retain fixed horizontal ownership, while
overlays consume transcript space above the composer instead of moving it.

Lip Gloss `Complete` colors map every semantic token across ANSI16, ANSI256, and
truecolor. `NO_COLOR` selects the ASCII profile before component construction.
Selection never depends on color: picker rows retain a `>` marker after ANSI is
removed, blocked rows include `[blocked]`, and transcript states keep literal
labels. Lip Gloss width measurement controls padding, truncation, centering, CJK
text, combining marks, cursor bounds, and line-width assertions. Deterministic
tests cover 140x40, 120x32, 80x24, 60x20, and terminal-too-small layouts.

## Phase 61 Runtime Integration

The frontend emits exactly one `IntentMsg` for submit, command, model, skill,
MCP, cancel, or quit. `internal/app.liveModel` is the only adapter that consumes
those messages. It invokes command, agent, tool, MCP, cancellation, or process
owners and returns immutable `SnapshotMsg`, semantic `TranscriptMsg`, or
`ClearTranscriptMsg` values to the TUI. The adapter never reads or writes
textarea, viewport, list, help, layout, cursor, or theme internals.

Exact-once tests prove one prompt creates one user block and one model turn; one
selection creates one command result; repeated cancellation calls the stored
cancel function once; unknown intents are inert; and quit returns one Bubble Tea
quit message. Existing fake-agent, built-in-tool, MCP-tool, blocked-result, and
cancellation tests prove the same boundary through complete runtime workflows.

## Phase 62 Release Profile Correction

The RecompHamr theme and the Bubbles-owned textarea, list, filter, pagination,
and help chrome now share one terminal capability profile. Component styles are
reapplied when Bubble Tea reports a profile change. ANSI16 frames cannot contain
ANSI256 or truecolor foreground sequences, ANSI256 frames cannot contain
truecolor foreground sequences, and ASCII/`NO_COLOR` frames contain no escape
sequences. This closes the remaining difference between semantic RecompHamr
styles and Bubbles defaults without replacing component interaction behavior.

The release executable and deterministic smoke evidence are recorded in
`docs/dev/07_ProjectManagement/Phase62TUIAcceptanceEvidence.md`. Runtime and
geometry verification are complete; real-terminal visual acceptance remains a
manual user decision.

### Focus-Safe Cursor Contract

Bubbles textarea returns no real cursor while its model is blurred. Because the
top-level view requests focus reporting, a terminal may send `BlurMsg` before
the first startup frame. Startup and chat treat the component cursor as
optional: an absent cursor produces a complete frame with the terminal cursor
hidden, and a later `FocusMsg` restores normal cursor rendering. The renderer
never synthesizes an editing position outside the textarea.

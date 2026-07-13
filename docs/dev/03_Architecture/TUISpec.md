# TUI Reference And Parity Specification

> Phase 46: `TUIUXSpec.md` is the current visual and interaction acceptance
> source for corrective phases 47-53. This file remains historical parity and
> architecture context.

Phase 30 defined the first corrective TUI target before rendering code changes.
User screenshot review rejected the resulting layout, so phases 37-44 supersede
that first polish track with a full Bubble Tea v2 rewrite. This remains a
RecompHamr-owned design contract, not an OpenCode or RecompHamr 1.x copy.

## Evidence Sources

- Current RecompHamr TUI: `internal/tui`, `internal/app`, `TUIArchitecture.md`,
  and `TUIParity.md`.
- RecompHamr parity requirements: transcript rendering, multiline prompt,
  slash palette, completions, footer, skill and MCP indicators, streaming tool
  status, paste handling, cancellation, quit behavior, and redacted debug mode.
- User screenshots: OpenCode-like terminal polish and RecompHamr banner/mockup
  references supplied in this thread.
- OpenCode reference checkout: temporary read-only clone at
  `%TEMP%\recomphamr-ui-references\opencode`.
- Bubble Tea v2 docs: `https://pkg.go.dev/charm.land/bubbletea/v2`, reviewed
  for the `tea.Model` `Init`/`Update`/`View` contract, `tea.KeyPressMsg`,
  `tea.PasteMsg`, `tea.NewProgram`, and declarative `tea.View` behavior.

OpenCode observations are limited to broad product patterns: a dedicated TUI
package, prompt component, command palette, theme/logo areas, footer variants,
scrollback/runtime queue, and Windows terminal handling. RecompHamr must not
copy source, exact layout, command names, palette colors, or wording.

## Phase 38 Audit

The current local TUI is rejected for final acceptance because user screenshot
evidence shows unstable placement, sparse hierarchy, and debug-board remnants.
Source review found `internal/tui/tui.go` still concentrates state, Bubble Tea
translation, layout math, palette rendering, transcript rendering, styles, and
cursor calculations in one file. Existing tests mostly assert substring
presence, so they can pass while the terminal layout remains visually broken.

The rewrite must therefore treat the existing renderer as historical behavior
evidence, not as a visual base to polish. Keep the app-facing intent contract
unless Phase 39 documents a compatibility shim, but rebuild the visual
composition, component boundaries, and golden acceptance from scratch.

Allowed OpenCode-inspired ideas, observed from the temporary checkout and user
screenshots, are:

- a separated terminal UI package with explicit host/runtime boundaries;
- immutable transcript or scrollback separated from mutable footer/composer
  state;
- command and modal overlays that are backed by real command/config state;
- model, skill/agent, provider, MCP, and status visibility in the command
  surface;
- screenshot or terminal smoke evidence before claiming polish.

Disallowed OpenCode reuse:

- exact green-heavy palette, logo treatment, modal sizing, command wording,
  source code, component names, or keybinding catalog;
- TypeScript/OpenTUI/Solid architecture in RecompHamr's Go/Bubble Tea code;
- fake timing, cost, token, model, or provider values copied from screenshots.

## Design Identity

RecompHamr's TUI must communicate local-first reverse-engineering work:

- Brand line: `RECOMP HAMR` or compact `RecompHamr` depending on width.
- Domain line: `RE . decomp . recomp . evidence-backed reconstruction`.
- Safety line: local commands run with user permissions; no model/tool work
  starts until the user submits a prompt or command.
- Status identity: model, memory, skill, MCP, workspace, and permission state
  are always visible or one keypress away.

Color tokens must be RecompHamr-owned: neutral black/gray base, hammer orange
accent, cyan activity accent, green verified state, yellow warning state, and
red blocked state. The exact OpenCode green-heavy palette is not used.

Final color roles:

| Role | Hex | Use |
|---|---|---|
| `base` | `#050607` | Terminal background. |
| `panel` | `#14171A` | Composer and modal panels. |
| `panelSoft` | `#0E1113` | Transcript cards and inactive surfaces. |
| `text` | `#E8EAED` | Primary readable text. |
| `muted` | `#8A929A` | Secondary metadata and quiet labels. |
| `hammer` | `#FF8A00` | RecompHamr brand, active prompt rail, destructive focus. |
| `cyan` | `#20D7F5` | Active command, cursor rail, in-progress state. |
| `verified` | `#72E06A` | Verified status and successful command/tool results. |
| `warning` | `#F5D547` | Unverified, missing config, or caution. |
| `blocked` | `#FF5C57` | Blocked/error states. |

Typography is terminal-native monospace only. Use weight, case, spacing, and
color roles for hierarchy; do not depend on custom fonts or bitmap assets.

Signature element: the `evidence rail`. It is a slim left accent used by the
active composer and important transcript cards. It belongs to RecompHamr's
evidence-first identity and must not appear as a generic decorative border.

## Layout Contract

The TUI has two lanes:

- **Transcript lane:** append-only visible work history. It owns user,
  assistant, tool, MCP, blocked, unsupported, unverified, attachment, and note
  blocks. Runtime code may append state; the TUI renders it.
- **Command lane:** mutable bottom area. It owns composer text, slash palette,
  modal overlays, footer status, hints, and selection state.

Startup is a special case with no transcript entries. It centers a compact
launcher in the upper-middle of the terminal, leaving breathable space below
but never pushing the cursor above the brand. Chat mode switches to transcript
lane plus fixed bottom command lane after the first user-visible transcript
entry.

Breakpoints:

- `wide`: width >= 112, centered launcher max width 84, transcript cards max
  width 104, modal max width 78.
- `medium`: width 80-111, launcher max width `width - 10`, transcript uses full
  width minus margins, footer wraps chips across two rows if needed.
- `compact`: width < 80, compact brand, one-line status chips, no side-by-side
  columns, modal rows truncate with preserved command names and state labels.

No layout may rely on full-screen blank padding to appear centered. All
vertical placement must be derived from terminal height, content height, and
explicit top/bottom reserves that tests can assert.

## Screen Contracts

### Startup

Required visible parts:

- compact or wide brand;
- domain line;
- safety line;
- composer panel with placeholder;
- status row with model, mode, memory, skill, MCP, context;
- key hints limited to implemented bindings;
- memory setup tip when memory is missing or unsupported.

Acceptance: no transcript labels appear; cursor lands inside composer; brand is
above composer; bottom half is not mostly empty on 1112x640 Windows Terminal.

### Chat

Required visible parts:

- recent transcript blocks with role labels or clear visual role treatment;
- status if running, blocked, cancelled, or waiting;
- bottom composer panel;
- status row visible at all widths;
- hints visible without overlapping the composer.

Acceptance: user and assistant entries are distinguishable; tool/MCP output is
styled as output; no fake timing/cost/token data appears.

### Slash Palette

Required visible parts:

- overlay above the composer or launcher;
- selected row;
- command name, summary, usage, and side-effect/error class when available;
- empty state for no matches;
- Esc hint and Tab/Enter behavior hints.

Acceptance: rows come from `internal/commands.Registry`; no hard-coded command
list is permitted.

### Model Picker

Required visible parts:

- active model marker;
- configured model profiles from config;
- context size or `unverified` if not locally known;
- provider URL redacted or summarized without leaking secrets;
- blocked state if config is unavailable.

Acceptance: model picker does not invent providers and does not mutate config
until the user confirms a selection through existing command/app ownership.

### Skill Picker

Required visible parts:

- active skill marker;
- embedded and custom skill source labels;
- skill-gated MCP implication text when known;
- blocked/unsupported state for missing custom directories.

Acceptance: rows derive from implemented skill registry behavior and user docs.

### MCP Modal

Required visible parts:

- server name, configured transport, state, autostart, and visible tool count;
- enabled/disabled tool status;
- skill gate status;
- blocked/error reason when connection failed.

Acceptance: MCP modal reads app-provided manager state only; the TUI does not
connect, disconnect, or spawn processes directly.

### Blocked And Unsupported

Required visible parts:

- clear label: `blocked`, `unsupported`, or `unverified`;
- exact actionable reason;
- command/tool/state that produced the condition when available;
- no success styling.

Acceptance: blocked cards are screenshot-tested and golden-tested.

## Required Screens

1. Startup/welcome: centered brand, domain, prompt panel, model/mode/status
   line, key hints, and setup tip.
2. Wide chat: transcript-first surface with bottom composer/status panel.
3. Compact chat: same state collapsed without losing model, memory, skill, MCP,
   pending tool, or blocked status.
4. Slash palette: registry-driven floating command overlay with selected row,
   summary, usage, and side-effect metadata from command docs.
5. Active skill: visible skill name and unlocked MCP/tool implications.
6. MCP state: disconnected, connected, error, and disabled tool states.
7. Tool transcript: PowerShell/tool/MCP command block with result or blocked
   reason.
8. Streaming state: thinking/tool-running status without storing private
   reasoning.
9. Blocked/unsupported state: clear label and actionable reason.

10. Model picker modal: configured profiles with active marker and blocked
    config states.
11. Skill picker modal: embedded/custom skills and active marker.
12. Help overlay: implemented keybindings and slash-command usage only.

## Interaction Requirements

- `/` opens or filters the slash palette.
- Tab completes command names or arguments from the registry.
- Enter submits; multiline content remains visibly distinct.
- Up/Down navigate prompt history when the palette is closed.
- Ctrl+C cancels active work; double Ctrl+C quits while idle.
- Ctrl+D quits immediately.
- Large paste becomes a paste chip with size/source metadata.
- Footer hints must show only implemented bindings.

## Rendering Rules

- Every visible item maps to local state, command metadata, or runtime evidence.
- Token, cost, timing, and model metrics render only when verified. Otherwise
  omit them or label them `unverified`.
- Tool output is bounded and styled as output, not assistant prose.
- Secrets are redacted before debug or transcript rendering.
- The plain renderer stays deterministic for tests, while the Bubble Tea view
  uses Lip Gloss styling and layout composition for the live terminal.
- Bubble Tea rendering must use the current Bubble Tea docs before design or
  implementation. For Bubble Tea v2, `View()` returns `tea.View`; terminal
  behavior such as alt screen, mouse mode, focus reporting, window title, and
  cursor shape is declared on the returned view; key handling uses
  `tea.KeyPressMsg`; bracketed paste uses `tea.PasteMsg`; styling uses Lip
  Gloss in the TUI boundary.
- Compact mode must not overlap or truncate critical status labels.
- Golden render tests are required for every required screen.

## Implementation Constraints For Phase 39-44

- `internal/tui` should split into focused files for model/update, Bubble Tea
  adapter, layout, theme, composer, palette, modals, transcript, status/footer,
  fixtures, and tests.
- `internal/tui` may import `internal/commands` for registry metadata and
  `internal/security` for redaction. It must not import or execute agent, LLM,
  tool, MCP lifecycle, config persistence, project mutation, release, or update
  logic.
- `internal/app` remains the side-effect owner and may pass state snapshots into
  the TUI.
- Bubble Tea v2 `tea.View` is the live-rendering contract; do not implement a
  separate manual terminal renderer for the live UI.
- Plain/golden rendering must use the same component/layout decisions as the
  styled Bubble Tea render so tests cannot pass a different UI.

## Acceptance Criteria

Phase 37-44 implementation is complete only when:

- all required screens have golden tests;
- all keybindings have model tests and user docs;
- command palette output is generated from `internal/commands`;
- no visual state is hard-coded fake data;
- docs, help, parity, status, and traceability rows are updated;
- current Bubble Tea v2 docs are read before every TUI implementation task;
- screenshot evidence covers startup, palette, chat, compact layout, blocked
  state, and model/MCP modal;
- `make verify` passes with 100% statement coverage.

# TUI Reference And Parity Specification

Phase 30 defines the corrective TUI target before rendering code changes. This
is a RecompHamr-owned design contract, not an OpenCode or RecompHamr 1.x copy.

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

OpenCode observations are limited to broad product patterns: a dedicated TUI
package, prompt component, command palette, theme/logo areas, footer variants,
scrollback/runtime queue, and Windows terminal handling. RecompHamr must not
copy source, exact layout, command names, palette colors, or wording.

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

## Acceptance Criteria

Phase 31-33 implementation is complete only when:

- all required screens have golden tests;
- all keybindings have model tests and user docs;
- command palette output is generated from `internal/commands`;
- no visual state is hard-coded fake data;
- docs, help, parity, status, and traceability rows are updated;
- `make verify` passes with 100% statement coverage.

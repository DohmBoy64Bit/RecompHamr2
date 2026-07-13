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
- Phase 49 replaces the cluttered startup with the specified five-group
  launcher. Wide, 80x24, and 60-column goldens verify responsive identity,
  shallow composition, concise mode/model state, three hints, and conditional
  memory guidance without duplicated secondary state.
- Phase 51 unifies rendered transcript scrolling with page/wheel input,
  preserves position across appended output, exposes a new-output marker,
  bounds tool/MCP output, and verifies every semantic runtime class through a
  deterministic golden.
- Phase 52 implements immutable per-render color profiles, explicit
  `NO_COLOR`, responsive width/cursor assertions, and Unicode/CJK/combining
  coverage across wide, medium, 80x24, 60-column, and minimum layouts.

Current runtime status:

- Product runtime wiring composes pure TUI state and reports `ready` at bare
  startup.
- Deterministic fake-runtime smoke renders slash-command and prompt transcripts
  through the pure TUI model.
- The live Bubble Tea process loop is implemented through `internal/app.Launch`.

Corrective TUI hardening phases 30-34 produced an initial Bubble Tea v2
end-user shell, but user screenshot review rejected the result as visually
unacceptable. Those rows remain historical evidence, not final acceptance.
Phases 37-44 supersede that polish path with a full TUI rewrite track.

Phase 31 visual foundation:

- Wide render uses `RECOMP HAMR` branding.
- Compact render uses `RecompHamr` branding.
- Domain, safety, signals/status chips, evidence column, footer hints, startup
  idle copy, and responsive compact status are implemented and tested.

Phase 32 composer and palette foundation:

- Slash palette rows are generated from `internal/commands.Registry`.
- Palette rows include selected marker, summary, and usage.
- `Tab` completes the first matching slash command.
- Prompt history, paste chips, cancellation, quit, and footer key hints remain
  covered by tests.

Phase 33 transcript foundation:

- Transcript lines render with deterministic block labels for user, assistant,
  tool, MCP, blocked, unsupported, unverified, status, attachment, and note
  output.
- Original line text is preserved after the label.
- The TUI does not invent timing, token, cost, or reasoning metrics.

Corrective rewrite acceptance:

- Phase 46 defines the replacement acceptance contract in `TUIUXSpec.md`:
  five-group startup maximum, compact growing composer, fixed transcript and
  command lanes, composer-anchored slash palette, centered state-backed
  pickers, semantic theme fallbacks, explicit 80x24/60-column/minimum behavior,
  and measurable clutter/stability gates.
- Phase 47 introduces Bubbles v2 textarea, viewport, help, and key mechanics,
  typed TUI intents, lazy-safe component initialization, and grapheme-aware
  width/truncation while preserving app-owned side effects and existing visible
  behavior for later corrective phases.
- Phase 48 routes printable, edit, focus, paste, multiline, page, and mouse-wheel
  input through Bubbles components; requests mouse reporting only in chat; and
  renders a deterministic terminal-too-small state below 60x18. Pure runtime
  tests cover every transition at 100% statement coverage.

- Phase 37 pins the rewrite and preserves the old roadmap.
- Phase 38 produces the non-copying RecompHamr-owned design spec before code
  changes, including startup, chat, slash palette, model picker, skill picker,
  MCP modal, help overlay, blocked cards, two-lane layout, breakpoints, and
  color roles.
- Phase 39 must reset the TUI architecture into testable concerns while
  preserving separation of concerns. Phase 39 is complete for the structural
  split: `internal/tui` now separates constants, types, model updates, Bubble
  Tea adapter, composer, layout, plain render, Bubble Tea render, palette,
  status, transcript, and styles while preserving app-facing behavior.
- Phase 40 is complete for the Bubble Tea shell rebuild: `BubbleModel.View()`
  returns a declarative `tea.View`; key and paste handling use Bubble Tea v2
  messages; styled content and cursor placement come from one pure screen pass;
  dead plain-render cursor helpers were removed; `internal/tui` and
  `internal/app` targeted coverage are 100%.
- Phase 41 is complete for startup/composer polish: wide styled startup uses
  `RECOMP HAMR`, compact styled startup uses `RecompHamr`, the safety line and
  local permission boundary are visible, memory appears in the status row,
  cursor placement follows the styled composer, and targeted coverage remains
  100%.
- Phase 42 is complete for command palette and modal overlays: command rows are
  registry-backed, Up/Down navigation changes the selected overlay row, Tab
  completes the selected slash command, and `/models`, `/skills`, `/skill`,
  `/mcp`, and `/help` render read-only overlays from command-owned model, skill,
  MCP, and help snapshots. Rendering does not mutate config, activate skills,
  connect MCP servers, or spawn processes.
- Phase 43 is complete for transcript and runtime states: plain and styled
  renderers classify user, assistant, tool, MCP, verification, blocked,
  unsupported, unverified, attachment, status, and note output; display-time
  redaction covers transcript and status text; and tests assert no fake token,
  cost, timing, or private reasoning markers are introduced.
- Phase 44 is complete for local screenshot and release smoke evidence:
  deterministic startup, palette, chat, compact, blocked, model modal, and MCP
  modal render captures are stored under
  `docs/dev/07_ProjectManagement/phase44_screenshots/`, the Windows
  `dist/recomphamr.exe` was built locally, `--summary` and `--diagnostic`
  smoked successfully, and the executable SHA-256 is recorded in
  `Phase44TUIScreenshotEvidence.md`.
- Phase 53 corrective layout work removes duplicated semantic labels, keeps
  routine progress out of transcript history, anchors bounded overlays above a
  fixed composer, derives cursor coordinates from the final rendered frame,
  and uses one centered content lane. Automated visual acceptance is retired;
  user-captured startup, palette, and transcript screenshots remain pending.
- Phase 54 supersedes phases 37-53 for implementation acceptance. The verified
  backend behaviors remain parity requirements, while the parallel composer,
  viewport, key routing, cursor, overlay, render, and intent-polling internals
  are explicitly rejected and scheduled for atomic replacement in Phase 56.
- Phase 56 replaces those internals. Backend parity behavior remains in its
  owning packages, while the new frontend uses one component tree and typed
  Bubble Tea messages. Focused TUI/app tests prove 100% statement coverage;
  visual and interaction acceptance remains assigned to phases 57-62.
- Phase 58 makes the Bubbles viewport authoritative for transcript scrolling and
  follow behavior. Semantic labels are normalized without duplication, secrets
  are redacted before rendering, paused readers receive explicit new-output
  feedback, and long tool/MCP blocks end with `output truncated`.
- Phase 59 makes one Bubbles list authoritative for every palette and picker.
  Registry-backed filtering consumes returned filter messages, browsing supports
  arrows and `j`/`k`, blocked or empty rows are inert, and accepted selections
  either emit one typed intent or populate the textarea when input is required.
- Phase 60 implements one responsive Lip Gloss layout tree with an original
  RecompHamr wordmark, stable composer geometry, semantic picker rows, explicit
  blocked/selected labels, four color profiles, and display-width-safe Unicode.
  Required sizes have deterministic clipping, height, width, and cursor tests.
- Phase 61 proves the typed frontend/backend boundary. Submit, command, model,
  skill, MCP, cancel, quit, and unknown intents are tested for exact-once or
  inert behavior, while fake agent, tool, MCP, blocked, and cancellation flows
  return immutable snapshots and semantic transcript messages.
- Phase 62 automated release gates build and smoke the Windows executable and
  verify that Bubbles component chrome obeys ANSI16, ANSI256, truecolor, and
  `NO_COLOR` profiles. Exact binary evidence is recorded in
  `Phase62TUIAcceptanceEvidence.md`. Final TUI parity acceptance remains pending
  until the user approves six real WezTerm screenshots.
- Phase 62 live acceptance found and corrected a startup panic when focus
  reporting delivered `BlurMsg` before the first frame. Both startup and chat
  now render safely with no component cursor while blurred, and app-boundary
  coverage reproduces the production message order. Fresh manual screenshots
  remain required.

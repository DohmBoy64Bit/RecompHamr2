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

Current runtime status:

- Product runtime wiring composes pure TUI state and reports `ready` at bare
  startup.
- Deterministic fake-runtime smoke renders slash-command and prompt transcripts
  through the pure TUI model.
- The live Bubble Tea process loop is implemented through `internal/app.Launch`.

Corrective TUI hardening phases 30-33 now track the remaining end-user polish:
startup/welcome screen, visual system, responsive layout, registry-driven slash
palette, composer completion, transcript/tool blocks, runtime feedback, and
golden renders for all required states.

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

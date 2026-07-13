# Phase 45 TUI Baseline

## Purpose

This baseline records why the Phase 37-44 visual result is superseded and pins
the evidence used by corrective phases 45-53. It does not claim that an
unobserved behavior is broken.

## Evidence Sources

- Source: `internal/tui` and the `internal/app` live Bubble Tea wrapper.
- Tests: `internal/tui/tui_test.go` and `internal/app/app_test.go`.
- Runtime visuals: the user-supplied OpenCode reference, live RecompHamr
  screenshot, and generated RecompHamr design mock supplied on 2026-07-13.
- Framework: Bubble Tea v2.0.8 package documentation and the local `tui-design`
  skill with Go, visual, interaction, and exemplar references.

## Verified Baseline Findings

| Concern | Source or visual evidence | Corrective gate |
|---|---|---|
| Startup clutter | Live screenshot shows safety, memory, MCP, context, permission, hints, and tip competing inside one launcher. Source repeats safety outside and permission inside the composer. | One identity line, one concise status line, no duplicated state, at most three persistent hints, and one conditional tip. |
| Composer height | `renderBubbleStartup` renders blank rows and status/permission text inside a padded panel. | Empty composer is compact and grows only for multiline content. |
| Palette placement | `overlayPalette` vertically joins the palette above the entire screen; it is not anchored to the composer. | Slash palette shares composer width and opens directly above it without moving the global layout. |
| Palette density | `PaletteRows` renders command, summary, and usage on every row. | Rows contain command and summary; selected detail contains usage and side effects once. |
| Cursor measurement | `composerCursorX`, startup cursor, and `centerText` use byte length. | All cell measurements use ANSI- and grapheme-aware display width. |
| Mouse contract | `tea.View` enables cell-motion mouse capture, while the adapter handles no mouse messages. | Enable only implemented mouse behavior; keyboard remains complete. |
| Render divergence | Plain and styled paths independently compose screens and overlays. | One layout/component tree feeds styled output and deterministic fixtures. |
| Responsive floor | Existing tests cover selected widths but no explicit 60-column or terminal-too-small contract. | Wide, medium, 80x24, 60-column, and minimum-size states are specified and tested. |

## Non-Copying Boundary

OpenCode is limited to broad observations: command-first hierarchy, restrained
startup density, composer-adjacent palette, transcript-first chat, and modal
pickers. RecompHamr must not copy OpenCode source, exact composition, wording,
glyph construction, green palette, or behavior 1:1.

## Evidence Rule

Corrective claims require three checks when applicable: local source truth,
automated tests, and runtime or deterministic render evidence. Unsupported
claims remain labeled `unverified`, `unsupported`, or `blocked`.

## Phase 54 Replacement Baseline

User screenshots after Phase 53 verify that the prior corrective track is not
accepted. `internal/tui` stores composer text in both its pure model and Bubbles
textarea, routes keys through custom and component handlers, stores transcript
position both as a custom offset and viewport state, derives overlays from
composer strings, and reconstructs the cursor from rendered output. The app also
polls mutable `LastIntent` state after updates. These are replacement targets,
not compatibility requirements.

The retained boundary is backend ownership: `internal/app` executes effects;
commands, agent, tools, MCP, config, skills, memory, LLM, and security remain in
their existing packages. The replacement may retain typed intent concepts and
verified runtime data, but no old widget state, renderer, key router, cursor
calculator, scrolling offset, palette implementation, or golden fixture is
preserved solely for compatibility.

# RecompHamr TUI UX Specification

## Status And Evidence

This Phase 46 specification is the acceptance source for corrective phases
47-53. It supersedes the visual acceptance portions of phases 37-44 without
deleting their historical evidence. Facts come from current RecompHamr source,
tests, supplied screenshots, Bubble Tea v2.0.8 and Bubbles v2 documentation,
and the local `tui-design` guidance.

OpenCode supplies only broad reference patterns: command-first hierarchy,
restrained startup density, composer-adjacent filtering, transcript-first chat,
and focused modal selection. RecompHamr does not copy OpenCode source, wording,
glyphs, exact composition, behavior, or green palette.

## Product Character And Priority

RecompHamr is an evidence-first reverse-engineering workstation. Its interface
is calm, compact, keyboard-first, Windows Terminal-first, and usable without
color. A cyan left evidence rail marks the focused composer and user transcript
blocks. Orange identifies RecompHamr; cyan identifies focus and information;
green means verified; yellow means warning or unverified; red means blocked.

Information priority is input, visible work, active mode/model, actionable
state, then project details. Startup must not display a state twice. Empty or
default secondary state is omitted. Full memory, skill, MCP, context, and
project state remains available through commands and pickers.

## Semantic Theme

| Token | Truecolor | 256-color | No-color signal |
|---|---|---|---|
| `brand` | `#FF9D2E` | ANSI 214 | bold `HAMR` |
| `focus` | `#21D4E8` | ANSI 45 | left rail and reverse selection |
| `verified` | `#72E06A` | ANSI 120 | `verified` or `ready` label |
| `warning` | `#F5D547` | ANSI 220 | `warning` or `unverified` label |
| `blocked` | `#FF5C57` | ANSI 203 | `blocked` or `failed` label |
| `text` | `#E6E6E6` | ANSI 255 | normal text |
| `muted` | `#8A918A` | ANSI 244 | dim metadata |
| `surface` | `#171A18` | ANSI 235 | whitespace-delimited panel |
| `selection` | `#21D4E8` | ANSI 45 | reverse video plus `>` |

No state depends on color alone. `NO_COLOR` removes color decoration while
preserving labels, rails, selection markers, weight, and spacing. Unsupported
glyphs use ASCII fallbacks.

## Responsive Contract

| Class | Width | Layout |
|---|---:|---|
| Wide | 112 or more | launcher max 84; transcript max 104; modal max 78 |
| Medium | 80-111 | 5-cell side margins; full-width transcript lane |
| Compact | 60-79 | 2-cell margins; compact brand; one-line metadata |
| Too small | below 60 or height below 18 | diagnostic with current/required size |

At 80x24 every primary action remains visible. At 60 columns all content is
single-column and descriptions truncate by display width. The too-small screen
accepts resize, `Ctrl+C`, and `Ctrl+D` without rendering clipped controls.

## Startup

Startup contains exactly five groups: identity, domain, composer, hints, and an
optional actionable tip. The group is centered horizontally near the vertical
center without full-screen blank padding.

```text
                         RECOMP HAMR
              evidence-backed reconstruction

              | Ask RecompHamr... "map this function"
              | Build  lmstudio-amd  ready

                         tab complete   / commands   ? help

              Tip: /init-re creates project memory.       (conditional)

~                                                        2.0.0
```

- `RECOMP` uses primary text; `HAMR` uses orange. RecompHamr-owned wordmark
  glyphs fit within 44 cells.
- Compact domain text is `RE / decomp / recomp`.
- Empty composer height is three rows including metadata. It grows to seven
  rows before internal scrolling.
- Composer metadata contains mode, model, and running/blocked state only.
- Persistent hints are limited to three context-sensitive actions.
- The tip renders only for a locally verified actionable condition.
- Bottom corners show redacted working directory and verified version.

## Chat

```text
  user
  | map this function

  assistant
    Evidence indicates ...

  tool  powershell  verified
    > rg "symbol" internal/
    4 matches

  [scrollable transcript viewport]

  | Ask a follow-up...
  | Build  lmstudio-amd  ready
  / commands   PgUp/PgDn scroll   Ctrl+C cancel
```

Transcript consumes remaining height; the command dock is fixed at the bottom.
New output follows only while the viewport is at the bottom. A `new output`
marker appears when output arrives while scrolled up. Tool output is bounded
and is never styled as assistant prose.

## Slash Palette

The slash palette attaches directly above the composer, shares its width, and
does not move identity or transcript lanes.

```text
  5/11 commands
  > /models       Switch model profile
    /mcp          Manage MCP servers
    /status-re    Inspect project status
  usage: /models [name]   writes config after confirmation
  | /m
```

- Visible rows: seven wide/medium and five compact.
- Rows contain command and summary only. Selected detail contains usage and
  side effects once.
- Selection uses reverse video and `>`; filter count is always visible.
- `/` opens; typing filters; Up/Down and `j`/`k` navigate; Tab completes; Enter
  accepts; Esc closes.
- Enter executes a complete no-argument command. Commands requiring arguments
  are inserted with a trailing space and remain in the composer.
- No matches renders `No commands match <query>` as neutral information.
- Rows derive exclusively from `internal/commands.Registry`.

## Modal Pickers

Model, skill, MCP, and help use a centered modal over a dimmed screen. Width is
64 cells capped by terminal width minus four; height is content-limited and
capped by terminal height minus four.

- Header contains title and `esc`; optional filter follows.
- Active entries use `*`; unavailable entries state `blocked`, `unsupported`,
  or `unverified`.
- Enter emits a typed intent. The TUI never changes config, skills, MCP
  processes, files, or network state directly.
- Model rows show profile/provider and verified context only.
- Skill rows show active state and `embedded` or `custom` source.
- MCP rows show server, transport, connection state, and visible tool count.
- Help derives from implemented bindings and command registry metadata.

## Runtime States

| State | Required treatment |
|---|---|
| User | `user` label and cyan evidence rail |
| Assistant | `assistant` label and primary prose |
| Tool | tool name, state, bounded command/result |
| MCP | server/tool name, state, bounded result |
| Verification | `verified` label and evidence summary |
| Warning | `warning` or `unverified` plus actionable reason |
| Blocked | `blocked`, operation, exact reason, known recovery |
| Unsupported | `unsupported` and documented boundary |
| Attachment | stable name, byte size, and source class |
| Streaming | `working`, `tool running`, or `verifying`; no private reasoning |

Token, cost, timing, and model metrics are omitted unless runtime evidence
supplies verified values. Secrets are redacted before measurement and render.

## Interaction Contract

| Context | Key | Behavior |
|---|---|---|
| Composer | Enter | submit or accept palette selection |
| Composer | Shift+Enter when distinguishable | newline |
| Composer | Ctrl+J | legacy newline fallback |
| Composer | Up/Down | history only when input does not consume key |
| Global | `/` | open/filter command palette |
| Global | `?` | contextual help |
| Palette/modal | Up/Down, `j`/`k` | move selection |
| Palette/modal | Tab | complete selected value |
| Palette/modal | Enter | accept selected value |
| Palette/modal | Esc | dismiss without mutation |
| Transcript | PgUp/PgDn, mouse wheel | scroll viewport |
| Active work | Ctrl+C | cancel active work |
| Idle | Ctrl+C twice | quit with visible armed state |
| Global | Ctrl+D | quit cleanly |

Mouse is additive and never required. Capture is disabled unless an implemented
mouse interaction is active.

## Clutter And Stability Gates

- No full-screen outer border or nested decorative borders.
- At most one rail between terminal edge and content.
- No state uses more than two simultaneous signals.
- Startup chrome consumes at most 20% of visible nonblank cells.
- Startup has at most five information groups and three persistent hints.
- Palette rows never repeat usage or side-effect prose.
- The command dock does not move as palette/status/transcript content changes.
- Truncation uses display width, preserves identifiers, and ends in one
  ellipsis cell.
- Cursor remains within composer and terminal bounds after resize, paste,
  modal transition, and profile change.

## Acceptance Matrix

Every required screen receives source tests, deterministic fixtures, and
runtime evidence at 120x32, 80x24, and 60x20 where applicable. Profiles cover
no-color, ANSI, 256-color, and truecolor. Acceptance fails on clipping,
overlap, duplicated status, invented state, stale registry rows, unredacted
secrets, divergent render paths, or undocumented interaction.

## Phase 49 Implementation Evidence

The minimal startup hierarchy is implemented. Golden fixtures at 120x32,
80x24, and 60x20 verify identity, responsive domain text, composer,
mode/model/readiness metadata, and three hints. Tests prove that project-memory
guidance is absent by default and appears only for reported missing or
unsupported memory. Version and working-directory values are not present in
`tui.Layout`, so the renderer omits corner metadata instead of inventing it.

## Phase 51 Implementation Evidence

Chat uses offset-from-bottom follow state, an explicit new-output marker,
bounded multiline tool/MCP blocks, semantic warning/evidence labels, and a
fixed concise command lane. `runtime_states.golden` covers every required
runtime class; model and app tests cover scrolling, paused append, follow
restoration, cancellation, redaction, and app-owned output.

## Phase 52 Implementation Evidence

Per-render semantic themes now consume Bubble Tea terminal profiles and Lip
Gloss profile completion. Deterministic tests cover ASCII/`NO_COLOR`, ANSI16,
ANSI256, and truecolor at every supported breakpoint, enforce line-width and
cursor bounds, and retain non-color state signals. CJK, combining marks, long
models, and long composers are included in the matrix.

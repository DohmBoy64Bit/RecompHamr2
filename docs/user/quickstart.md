# Quickstart

RecompHamr 2.0 launches the local terminal app by default:

```sh
go run ./cmd/recomphamr
```

The Phase 47 component replacement does not change commands or visible user
behavior. Bubble Tea v2 remains the runtime while official Bubbles components
provide the internal composer, transcript viewport, and key-help mechanics.

The Phase 56 replacement removes the previous synchronized editor and render
paths. The visible composer is now the actual Bubbles textarea: its prompt hint
is a placeholder, Backspace/Delete edit normally, and bare `/` opens command
selection without executing an unknown slash command. App-owned model, tool,
MCP, command, configuration, memory, and security behavior is unchanged.

The transcript follows new output while it is at the bottom. After scrolling up,
new output does not move the current reading position; the command lane shows
`new output  PgDn to follow` until the transcript returns to the bottom. Long
tool and MCP responses are visibly bounded with an `output truncated` marker.

The startup hint is not editable content. Begin typing normally; Backspace and
Delete affect only text you entered. Typing `/` opens the command palette, and
Backspace immediately removes a bare slash and closes the palette. Use
Shift+Enter or Ctrl+J for a newline and Enter to submit non-empty input.

In command and selection overlays, type to filter and use Up/Down to navigate.
After accepting a filter, `j`/`k` also navigate the list. Enter accepts the
highlighted row, Tab places it in the composer, and Escape closes the overlay.
Commands that require an argument are placed in the composer and are not run
until the argument is supplied and the completed command is submitted.

Bare startup loads or creates `.rehamr/config.yaml`, loads optional
`.rehamr/REPHAMR_STATE.md` memory when present, wires the slash-command
environment, creates an MCP manager without autoconnecting servers, and prepares
the Bubble Tea interface. From the TUI, type slash commands such as `/help` or
submit a prompt to the configured OpenAI-compatible backend.

The redesigned TUI opens on a centered launcher with two-tone `RECOMP HAMR` on
wide and 80-column terminals or compact `RecompHamr` at 60-79 columns. It shows
only the domain line, shallow prompt panel, active mode/model/readiness, three
key hints, and actionable `/init-re` guidance when project memory is reported
missing or unsupported. Secondary memory, skill, MCP, context, permission, and
tool detail stays behind commands instead of cluttering startup. After a prompt
or command, the screen becomes
transcript-first with a fixed bottom composer/status panel. Typing `/` opens a
bounded registry-backed command palette directly above the composer. It consumes
transcript space instead of moving the command lane. Typing exact `/models`, `/skills`,
`/skill`, `/mcp`, or `/help` opens a read-only model, skill, MCP, or help
overlay above the composer. The live app runs through Bubble Tea v2 with
declarative view fields for alt screen, mouse mode, focus reporting, title, and
cursor, plus Lip Gloss styling for the visible terminal shell.

Transcript lines are labeled as user, assistant, tool, MCP, verification,
blocked, unsupported, unverified, attachment, status, or note output. The TUI
redacts configured secrets from visible transcript and status text and does not
display token, cost, timing, or reasoning metrics unless documented runtime
state provides verified values.

Each semantic label appears once per block; repeated prefixes such as `user:`
or `assistant:` are normalized from the body. Routine progress stays in the
fixed command lane rather than becoming conversation history.

Composer keys:

- `/` opens the slash command palette.
- `Up` and `Down` move the selected command or modal row when an overlay is open.
- `Tab` completes the selected matching slash command.
- `Up` and `Down` navigate prompt history when no overlay is open.
- `Shift+Enter` or `Ctrl+J` inserts a newline; `Enter` submits.
- `Page Up`, `Page Down`, or the mouse wheel scrolls an active transcript.
- `Ctrl+C` cancels active work or arms quit while idle.
- `Ctrl+D` exits.
- `Esc` clears transient quit/status state.

The supported interactive floor is 60 columns by 18 rows. Below that size the
TUI shows the required dimensions and keeps `Ctrl+D` available for a clean exit.
Mouse input is optional; every action remains keyboard accessible.

Set the standard `NO_COLOR` environment variable to any value before launch to
disable terminal colors. Labels, the focused-input rail, bold warnings, and
reverse-video selection remain visible. Without `NO_COLOR`, RecompHamr follows
the terminal profile reported by Bubble Tea: ANSI16, ANSI256, or truecolor.

At 80 columns and wider, startup uses the full RecompHamr block wordmark. A
compact literal wordmark is used from 60 through 79 columns. Below 60x18 the app
shows a terminal-too-small message instead of clipping interactive controls.
Picker selection always includes a `>` marker, and blocked rows include
`[blocked]`, so navigation and state remain visible without color.

The transcript follows output at the bottom. If you scroll up, incoming output
preserves the visible history and shows `new output` until `Page Down` or the
mouse wheel returns to the bottom. Long tool and MCP results state `output
truncated` after 12 rendered lines. Runtime blocks always state their class.

Deterministic startup evidence remains available without launching the TUI:

```sh
go run ./cmd/recomphamr --summary
```

Diagnostic status remains available:

```sh
go run ./cmd/recomphamr --diagnostic
```

Startup itself does not call a model backend. Prompt submission inside the TUI
uses the active model profile, the agent loop, the built-in tool dispatcher, and
connected MCP tools that are enabled and unlocked by the active skill. MCP
autostart is limited to server configs with explicit autostart metadata.

For a local Windows executable instead of `go run`, build one with:

```powershell
go build -trimpath -o .\dist\recomphamr.exe .\cmd\recomphamr
.\dist\recomphamr.exe --summary
.\dist\recomphamr.exe --diagnostic
.\dist\recomphamr.exe
```

Use the non-interactive flags first when validating a new build. `--summary`
prints runtime composition evidence and `--diagnostic` prints offline status;
running without flags opens the TUI.

Published Windows archives are available from the `v2.0.0` release:

```text
https://github.com/DohmBoy64Bit/RecompHamr2/releases/download/v2.0.0/recomphamr_windows_amd64.zip
https://github.com/DohmBoy64Bit/RecompHamr2/releases/download/v2.0.0/SHA256SUMS
```

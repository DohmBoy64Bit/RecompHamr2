# Quickstart

RecompHamr 2.0 launches the local terminal app by default:

```sh
go run ./cmd/recomphamr
```

The Phase 47 component replacement does not change commands or visible user
behavior. Bubble Tea v2 remains the runtime while official Bubbles components
provide the internal composer, transcript viewport, and key-help mechanics.

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
transcript-first with a bottom composer/status panel. Typing `/` opens a
floating registry-backed command palette. Typing exact `/models`, `/skills`,
`/skill`, `/mcp`, or `/help` opens a read-only model, skill, MCP, or help
overlay above the composer. The live app runs through Bubble Tea v2 with
declarative view fields for alt screen, mouse mode, focus reporting, title, and
cursor, plus Lip Gloss styling for the visible terminal shell.

Transcript lines are labeled as user, assistant, tool, MCP, verification,
blocked, unsupported, unverified, attachment, status, or note output. The TUI
redacts configured secrets from visible transcript and status text and does not
display token, cost, timing, or reasoning metrics unless documented runtime
state provides verified values.

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

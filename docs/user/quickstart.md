# Quickstart

RecompHamr 2.0 launches the local terminal app by default:

```sh
go run ./cmd/recomphamr
```

Bare startup loads or creates `.rehamr/config.yaml`, loads optional
`.rehamr/REPHAMR_STATE.md` memory when present, wires the slash-command
environment, creates an MCP manager without autoconnecting servers, and prepares
the Bubble Tea interface. From the TUI, type slash commands such as `/help` or
submit a prompt to the configured OpenAI-compatible backend.

Composer keys:

- `/` opens the slash command palette.
- `Tab` completes the first matching slash command.
- `Up` and `Down` navigate prompt history.
- `Ctrl+C` cancels active work or arms quit while idle.
- `Ctrl+D` exits.

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

Published `.exe` downloads remain blocked until external release publication
evidence exists.

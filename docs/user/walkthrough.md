# Fresh-Clone Walkthrough

This walkthrough covers the verified RecompHamr 2.0 local path in this checkout.
Live model turns, live tool execution through the product runtime, MCP
autoconnect, remote release downloads, and the interactive Bubble Tea process
remain `unsupported`.

## 1. Verify The Checkout

```powershell
make verify
go run ./cmd/recomphamr --diagnostic
```

`make verify` runs docs coverage, architecture checks, and the 100% statement
coverage gate. Diagnostic mode prints foundation status without mutating model,
tool, or MCP state.

## 2. Compose Local Runtime State

```powershell
go run ./cmd/recomphamr
```

Bare startup creates or loads `.rehamr/config.yaml`, reads optional
`.rehamr/REPHAMR_STATE.md`, wires slash commands, registers MCP server metadata,
and prepares pure TUI state. It does not call a backend model.

## 3. Configure A Model Profile

Edit `.rehamr/config.yaml`:

```yaml
active: lmstudio-amd
logging: false
models:
  lmstudio-amd:
    llm: qwen/qwen3.6-35b-a3b
    url: http://localhost:1234
    key: ""
    context_size: 32768
```

Use `RECOMPHAMR_URL` for a process-local endpoint override:

```powershell
$env:RECOMPHAMR_URL = "http://localhost:1234"
go run ./cmd/recomphamr
```

## 4. Initialize Reverse-Engineering Memory

Through the command registry, `/init-re` creates `.rehamr/` memory and ledger
files. `/status-re` reads those files and reports missing tracked items.

Important generated files include `PROJECT.md`, `REPHAMR_STATE.md`,
`EVIDENCE.md`, `HYPOTHESES.md`, `BLOCKERS.md`, `COMMANDS.md`, `TOOLCHAIN.md`,
`MODELS.md`, `mcp.json`, and `skills/active.md`.

## 5. Use Skills And Tools

`/skills` lists 28 embedded skills plus custom `.rehamr/skills/*.md` files.
Use `/skill ghidra-mcp` or another documented skill name to activate a skill
for the current session.

Tool schemas are `powershell`, `read_file`, `write_file`, `edit_file`,
`repomixr`, and `recomp_reference`. `bash` is retained as a 1.x compatibility
alias, but Windows-focused workflows should use `powershell`.

## 6. Inspect MCP State

Use `/mcp` to list built-in registrations. With manager wiring,
`/mcp connect <server>`, `/mcp tools <server>`, `/mcp enable <server> <tool>`,
`/mcp disable <server> <tool>`, and `/mcp disconnect <server>` are supported
for registered HTTP-capable manager flows. Stdio process spawning and app
autostart remain `unsupported`.

## 7. Run Doctor And Troubleshoot

```text
/doctor
/help /doctor
```

`/doctor` is offline and non-mutating. It reports `verified`, `unsupported`, and
`blocked` sections for runtime, workspace, config, memory, skills, tools, MCP
registration, and operational release files.

Use `docs/user/troubleshooting.md` when output includes `usage:`,
`unsupported:`, `unverified:`, or `blocked:`.

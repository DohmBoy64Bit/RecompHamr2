# Phase 44 TUI Screenshot Evidence

Generated on 2026-07-12 in the local Windows checkout at `E:\RecompHamr2`.

## Capture Method

The screenshot evidence was generated from the production `internal/tui`
renderer through a temporary local harness that imported `internal/tui` and
`internal/commands`, rendered deterministic TUI states with
`Model.RenderWithLayout`, and converted the resulting terminal text into PNG
captures. The temporary harness was removed after capture. The captures are
render evidence for the implemented TUI states; they do not claim external CI or
hosted screenshot publication.

## Captures

| State | Text evidence | PNG evidence |
|---|---|---|
| Startup | `docs/dev/07_ProjectManagement/phase44_screenshots/startup.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/startup.png` |
| Slash palette | `docs/dev/07_ProjectManagement/phase44_screenshots/palette.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/palette.png` |
| Chat/runtime | `docs/dev/07_ProjectManagement/phase44_screenshots/chat.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/chat.png` |
| Compact layout | `docs/dev/07_ProjectManagement/phase44_screenshots/compact.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/compact.png` |
| Blocked/unsupported | `docs/dev/07_ProjectManagement/phase44_screenshots/blocked.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/blocked.png` |
| Model modal | `docs/dev/07_ProjectManagement/phase44_screenshots/model_mcp_modal.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/model_mcp_modal.png` |
| MCP modal | `docs/dev/07_ProjectManagement/phase44_screenshots/mcp_modal.txt` | `docs/dev/07_ProjectManagement/phase44_screenshots/mcp_modal.png` |

## Local Binary Smoke

Command:

```powershell
go build -trimpath -o dist/recomphamr.exe ./cmd/recomphamr
.\dist\recomphamr.exe --summary
.\dist\recomphamr.exe --diagnostic
Get-FileHash .\dist\recomphamr.exe -Algorithm SHA256
```

Observed summary:

```text
RecompHamr product runtime
project: .
config: loaded active=lmstudio-amd profiles=4
memory: unsupported: REPHAMR_STATE.md is missing
mcp: manager wired servers=8 autoconnect=false
tui: ready
agent: wired for interactive turns; no model call made during summary
```

Observed diagnostic:

```text
recomphamr diagnostic mode
phase: architecture-skeleton
phase-0: complete: RecompHamr 1.x reference inventory is recorded
runtime: product wiring available; use bare recomphamr to launch the terminal app
```

Binary hash:

```text
SHA256 dist/recomphamr.exe 7892BFD9F39C7D06D4A90D0B67E355F800A40FBB817CE86892F036C6DB52F0FB
```

## Verification

- `go test ./internal/tui ./internal/app ./internal/agent -cover` passed with
  100.0% statement coverage for each package during Phase 43 closure.
- `make docscheck` passed after Phase 43 documentation updates.
- `make docscheck` passed after this Phase 44 evidence ledger and roadmap
  update.
- `make verify` passed after this Phase 44 evidence ledger and roadmap update;
  every Go package reported 100.0% statement coverage and `archcheck` reported
  that separation-of-concerns boundaries hold.

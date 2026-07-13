# Phase 62 TUI Acceptance Evidence

Generated on 2026-07-13 in the local Windows checkout at `E:\RecompHamr2`.

## Automated Gate

`make verify` passes after the replacement TUI and runtime integration. The
gate reports 100.0% statement coverage for every Go package, passes
`docscheck`, and passes `archcheck`.

Deterministic TUI tests cover startup, chat, palette, model/skill/MCP/help
pickers, blocked and unsupported states, multiline editing, paste, history,
scroll-follow behavior, cancellation, resize, Unicode display width, and
ANSI16, ANSI256, truecolor, and `NO_COLOR` output. Integration tests prove
submit, command, model, skill, MCP, cancel, and quit intents cross the app
boundary exactly once. These tests verify behavior and geometry; they do not
approve visual quality.

## Windows Executable

Build and smoke commands:

```powershell
go build -trimpath -o dist/recomphamr.exe ./cmd/recomphamr
.\dist\recomphamr.exe --summary
.\dist\recomphamr.exe --diagnostic
Get-FileHash .\dist\recomphamr.exe -Algorithm SHA256
```

Observed evidence:

- path: `E:\RecompHamr2\dist\recomphamr.exe`;
- size: `14183424` bytes;
- build timestamp: `2026-07-13T16:04:53.5790699Z`;
- SHA-256: `dc2a69e4e5ca64989a3aa9aae20f9072f30ded7ebb636c26caf9e1347daf98ab`;
- summary first line: `RecompHamr product runtime`;
- diagnostic first line: `recomphamr diagnostic mode`.

## Live Startup Correction

The first manual launch exposed a nil-pointer panic in `renderStartup`. Source
inspection and the Bubbles textarea contract verified that `Cursor()` returns
no cursor while blurred; Bubble Tea focus reporting can deliver that state
before the first frame. Startup and chat now hide the terminal cursor when the
component cursor is absent. Focus restores the component-owned cursor.

Regression evidence covers blurred startup, blurred chat, and the production
app boundary receiving `BlurMsg` before its first `View`. The focused packages
and canonical gate pass at 100% statement coverage. The executable evidence
above replaces the panic-prone build.

## Manual Acceptance

Manual acceptance is pending. The user must capture the real executable in
WezTerm for startup, slash palette, active chat, model picker, blocked state,
and an 80x24 terminal. Automated screenshots and deterministic frames cannot
close this gate. Rejected screenshots reopen the phase that owns the affected
component; accepted screenshots close Phase 62 and permit the preserved roadmap
to resume.

## Documentation Impact

Architecture, parity, traceability, roadmap, status, stable-release, and this
evidence ledger changed. User commands and configuration did not change;
`docs/user/quickstart.md` remains the correct launch, key, and color-profile
guide and is intentionally unchanged.

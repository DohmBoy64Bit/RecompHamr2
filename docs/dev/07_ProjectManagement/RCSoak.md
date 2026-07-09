# RC Soak And Bugfix Freeze

Phase 26 records the local release-candidate soak process. It adds no product
features and does not change runtime behavior.

## Freeze Rule

During RC soak, allowed changes are limited to:

- verified release blockers;
- documentation corrections that remove drift or unsupported claims;
- test fixes for existing intended behavior;
- security fixes with regression evidence.

Disallowed changes:

- new commands, tools, skills, MCP servers, config keys, or user-visible
  behaviors;
- feature polish outside parity scope;
- published artifact claims without local checksum evidence;
- silent fallback behavior or fake success paths.

## Local Soak Matrix

Run the matrix repeatedly before stable release:

```powershell
make verify
go run ./cmd/recomphamr --diagnostic
go test ./internal/release ./internal/update ./internal/doctor -cover
rg -n "TODO|placeholder-only|fake success path|speculative API|silent unsupported" . -g "!/.git/**" -g "!/.reference/**" -g "!internal/skills/embedded/**"
```

Expected local result:

- `make verify` passes with docscheck, archcheck, and 100% statement coverage.
- Diagnostic mode exits successfully.
- Release, update, and doctor packages report 100% statement coverage.
- The placeholder-policy scan reports only policy text, not implementation
  placeholders.

## Blocker Policy

A blocker must include:

- exact command output or test failure;
- affected files or packages;
- whether the issue is parity, security, docs, release, or platform related;
- a regression test or documented no-code rationale;
- updated `StatusReports.md` and `TraceabilityMatrix.md`.

After three failed attempts at the same blocker, record `blocked:` evidence and
stop instead of speculating.

## Current Soak Result

Local soak is complete for this checkout when the Phase 26 final verification
commands pass. Stable release still requires the remaining Phase 27 gate,
including intentional publication decisions and any external CI evidence.

# Known Limits

This ledger records verified limits for the current release-candidate
preparation state. These are explicit `unsupported` boundaries, not hidden
success paths.

## Runtime Limits

- Stdio MCP process spawning is implemented and security-sensitive: configured
  commands run with the user's local permissions.
- Persistent user MCP config is limited to the documented `.rehamr/mcp.json`
  keys: `servers`, `name`, `command`, `args`, `url`, `allowed_tools`,
  `autostart`, and `require_skill`.
- MCP autoconnect is limited to configs with explicit `autostart` metadata.

## Release Limits

- Local Windows `.exe` build and launch are supported from the checkout with
  `go build -trimpath -o .\dist\recomphamr.exe .\cmd\recomphamr`; local
  `--summary`, `--diagnostic`, archive, and `SHA256SUMS` verification evidence
  are recorded in `StableRelease.md`.
- Published tags and uploaded release artifacts are verified for `v2.0.0` at
  `https://github.com/DohmBoy64Bit/RecompHamr2/releases/tag/v2.0.0`.
- Remote release downloads are `unsupported`.
- Remote checksum fetching is `unsupported`.
- Automatic replacement of the running executable is `unsupported`.
- Installer execution tests on every platform are `unsupported`.
- Dependency audit remains `unsupported`.
- Stable release publication is no longer blocked for `v2.0.0`: external CI,
  artifact, checksum, release, and timestamp evidence are recorded in
  `StableRelease.md`.

## Platform Limits

- Final union coverage across Linux, macOS, and Windows CI remains required
  before stable release.
- OS-specific process-group termination guarantees beyond Go `CommandContext`
  behavior are `unsupported`.

## Evidence

Current local evidence is `make verify`, diagnostic output, release package
tests, `PlatformMatrix.md`, `PerformanceBenchmarks.md`, `SecurityAudit.md`,
`ParityClosure.md`, `RCSoak.md`, `StableRelease.md`, and this file.

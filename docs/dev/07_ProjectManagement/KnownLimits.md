# Known Limits

This ledger records verified limits for the current release-candidate
preparation state. These are explicit `unsupported` boundaries, not hidden
success paths.

## Runtime Limits

- Real backend prompt/model turns are `unsupported` in product runtime until
  Phase 28 live end-user runtime integration closes.
- Real product-runtime tool execution is `unsupported` until Phase 28 wires
  tool dispatch with cancellation and documented status.
- MCP autoconnect is `unsupported`.
- Stdio MCP process spawning is `unsupported`.
- Persistent user MCP config files beyond current documented metadata are
  `unsupported`.
- Interactive Bubble Tea process launch is `unsupported` until Phase 28;
  current TUI coverage is the pure model and adapter tests.

## Release Limits

- Published tags and uploaded release artifacts are `unsupported` until the
  stable publication gate is intentionally cut. The local `v2.0.0` tag decision
  is release evidence, not upload evidence.
- Remote release downloads are `unsupported`.
- Remote checksum fetching is `unsupported`.
- Automatic replacement of the running executable is `unsupported`.
- Installer execution tests on every platform are `unsupported`.
- Dependency audit remains `unsupported`.
- Stable release publication is `blocked:` until external platform, upload, and
  publication destination evidence exist. Local artifact, checksum, Windows
  installer smoke, and stable tag decision evidence are recorded in
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

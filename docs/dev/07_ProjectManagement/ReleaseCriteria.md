# Release Criteria

Release requires all parity rows passed, all commands/tools/MCP/skills implemented and documented, `.rehamr/` secure behavior complete, docs and code coverage gates passing, security audit passing, and known limits verified.

Release-candidate preparation additionally requires `ReleaseCandidate.md`,
`KnownLimits.md`, current `CHANGELOG.md`, checksum guidance, packaged-docs
contents, `make verify`, and diagnostic evidence. Published artifacts and tags
must not be claimed until the RC gate is intentionally cut.

Stable publication additionally requires a stable tag decision, generated
release artifacts, generated and verified `SHA256SUMS`, fresh-install evidence,
external CI or platform evidence where required, and upload evidence. Local
stable-gate readiness is not the same as a published stable release.

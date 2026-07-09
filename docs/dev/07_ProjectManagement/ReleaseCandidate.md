# Release Candidate

Phase 25 prepares the release-candidate record for the local checkout at commit
`4b15aa6`. This document is an RC preparation ledger, not a published release.

## RC Scope

Included:

- RecompHamr 1.x parity inventory and closure docs.
- Governance, memory, verification, architecture, testing, parity, and project
  management docs.
- Config and `.rehamr/` workspace behavior.
- OpenAI-compatible LLM streaming and context packing helpers.
- Built-in tool schemas with `powershell` as the primary Windows-first shell
  tool and `bash` as the compatibility alias.
- Agent loop, pure TUI model, slash command registry, embedded/custom skills,
  MCP protocol/manager foundations, doctor diagnostics, and local release
  helpers.
- Fresh-clone walkthrough and RecompHamr 1.x migration notes.

Excluded:

- Published tags, uploaded artifacts, and remote release downloads.
- Automatic binary replacement.
- Installer execution claims on every platform.
- Live backend prompt/model turns.
- Product-runtime tool execution.
- MCP autoconnect and stdio MCP process spawning.
- Interactive Bubble Tea process launch.

## RC Verification Commands

Run before any RC tag or artifact publication:

```powershell
make verify
go run ./cmd/recomphamr --diagnostic
go test ./internal/release ./internal/update ./internal/doctor -cover
```

Optional local artifact dry run:

```powershell
go test ./internal/release -run TestBuildBinary -count=1
```

The dry run builds local binaries only through test-controlled temporary
directories. It does not publish, install, download, or replace executables.

## Checksum Guidance

Use `internal/release.GenerateManifest`, `ManifestText`, `WriteManifest`, and
`VerifyManifest` for local artifacts already present on disk. The manifest file
is `SHA256SUMS`. Artifact paths must be relative to the release directory and
must not contain `..`, absolute paths, or root-relative paths.

Canonical archive names:

```text
recomphamr_windows_amd64.zip
recomphamr_windows_arm64.zip
recomphamr_linux_amd64.tar.gz
recomphamr_linux_arm64.tar.gz
recomphamr_darwin_amd64.tar.gz
recomphamr_darwin_arm64.tar.gz
```

## Packaged Docs Set

The RC docs package must include:

- `README.md`
- `CHANGELOG.md`
- `SECURITY.md`
- `CONTRIBUTING.md`
- `.docs-index.md`
- `docs/user/`
- `docs/dev/05_FeatureParity/ParityClosure.md`
- `docs/dev/07_ProjectManagement/ReleaseCandidate.md`
- `docs/dev/07_ProjectManagement/KnownLimits.md`

## RC Gate

Do not cut an RC tag unless `make verify` passes, known limits are current,
release notes are current, and every unsupported release/runtime claim is still
documented as `unsupported`.

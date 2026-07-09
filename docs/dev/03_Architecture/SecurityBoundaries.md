# Security Boundaries

Security boundaries include filesystem paths, config permissions, shell execution, network fetches, MCP execution, logs, generated files, and release verification.

Phase 3 filesystem boundary:

- `.rehamr/` and `.rehamr/config.yaml` are checked with `Lstat` so symlinks are refused before reading or writing.
- Config saves write a sibling temporary file with owner-only permissions, then rename over `config.yaml`.
- Workspace files are created with owner-only permissions on POSIX systems.
- `/init-re` never overwrites existing files, which protects hand-maintained evidence and memory.
- `/status-re` reads tracked workspace files and reports missing files instead of creating them.
- Direct memory/status reads through `internal/project` use `Lstat` and refuse
  symlinked `.rehamr/` workspaces. `LoadMemory` also refuses a symlinked
  `REPHAMR_STATE.md`.

Phase 20 audit boundary:

- `docs/dev/03_Architecture/SecurityAudit.md` records the current audit
  results, including passed boundaries, partial boundaries, verified
  unsupported limits, regression evidence, and verification commands.

Phase 12 release checksum boundary:

- `internal/release` builds, archives, and verifies only local checkout files
  and local files already present on disk.
- Binary builds run `go build -trimpath` with explicit `GOOS`, `GOARCH`, and
  `CGO_ENABLED=0`, write to deterministic output filenames, and refuse to
  overwrite existing binaries.
- Archive creation reads an already-built binary, creates the output directory
  when needed, and refuses to overwrite an existing archive.
- Archives contain only `recomphamr.exe` for Windows targets or `recomphamr` for
  Linux/macOS targets.
- Manifest generation reads only caller-specified relative artifact paths under
  the release directory and writes deterministic `SHA256SUMS` text.
- Installer scripts install only caller-specified local artifacts and verify
  `SHA256SUMS` before extraction.
- Self-update planning verifies a local artifact against a local manifest and
  returns a dry-run plan; it does not replace the running executable.
- Operational file validation reads only repository-relative Phase 12 files and
  checks required markers.
- Checksum manifest rows must use SHA-256 and a relative artifact path.
- Absolute paths, root-relative paths, and `..` traversal are blocked before
  file reads.
- Digest mismatches, missing files, malformed manifests, and unsafe paths are
  reported as blocked verification results or explicit errors.
- Release helpers do not download, fetch remote checksums, or execute built
  artifacts.

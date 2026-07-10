# Release Parity

Release parity covers install scripts, self-update, release manifests,
checksums, devcontainer support, and release dry-runs.

Implemented Phase 12 checksum slice:

- `internal/release.DefaultTargets` defines the supported artifact targets:
  Windows amd64/arm64 zip archives, Linux amd64/arm64 tar.gz archives, and
  macOS amd64/arm64 tar.gz archives.
- `internal/release.ArtifactName` returns names in the format
  `recomphamr_<goos>_<goarch>.<archive>`.
- `internal/release.BinaryFileName` returns `recomphamr.exe` for Windows
  archives and `recomphamr` for Linux/macOS archives.
- `internal/release.BinaryOutputName` returns deterministic local build output
  names such as `recomphamr_windows_amd64.exe` and `recomphamr_linux_amd64`.
- `internal/release.BuildBinary` runs `go build -trimpath` with `GOOS`,
  `GOARCH`, and `CGO_ENABLED=0`, defaults to `./cmd/recomphamr` and `dist`,
  and refuses to overwrite an existing binary.
- `internal/release.CreateArchive` creates local zip or tar.gz archives from an
  already-built binary and refuses to overwrite existing archive paths.
- `internal/release.ManifestName` defines the canonical manifest filename
  `SHA256SUMS`.
- `internal/release.GenerateManifest(rootDir, artifactPaths)` reads local
  artifacts under `rootDir`, hashes them with SHA-256, rejects unsafe relative
  paths, and returns entries sorted by artifact path.
- `internal/release.ManifestText(entries)` renders deterministic
  `SHA256SUMS` text from validated entries.
- `internal/release.WriteManifest(rootDir, artifactPaths, manifestPath)` writes
  the generated manifest; an empty `manifestPath` writes `<rootDir>/SHA256SUMS`.
- `internal/release.ParseManifest` parses `SHA256SUMS`-style manifests with
  `sha256 path` and `sha256 *path` rows.
- `internal/release.VerifyManifest(rootDir, manifestPath)` verifies local
  artifacts under `rootDir` against the manifest.
- Artifact paths must be relative and stay inside the release directory.
- Malformed manifests, missing artifacts, unsafe paths, and digest mismatches
  return explicit errors or blocked result rows.
- Reports expose verified/blocked counts and a user-facing string format.
- `scripts/install.ps1` installs a local Windows zip artifact after
  `SHA256SUMS` verification.
- `scripts/install.sh` installs a local POSIX tar.gz artifact after
  `SHA256SUMS` verification.
- `.goreleaser.yaml` defines local release builds, archives, and checksum
  output.
- `.devcontainer/devcontainer.json` defines a Go devcontainer that runs
  `make verify`.
- `.github/workflows/verify.yml` defines Linux, Windows, and macOS CI
  verification with `make verify`.
- `internal/release.ValidateOperationalFiles` verifies required operational
  files and required markers locally.
- `internal/update.PlanLocal` verifies a local artifact against `SHA256SUMS`
  and returns a self-update dry-run plan without replacing the executable.
- Phase 34 local Windows executable evidence built
  `%TEMP%\recomphamr-phase34\recomphamr.exe` with `go build -trimpath`, ran
  `--summary` and `--diagnostic`, archived it as
  `recomphamr_windows_amd64.zip`, and verified the local `SHA256SUMS` row.

Phase 25 RC preparation adds local release-candidate notes and a known-limits
ledger. Phase 27/35 publication repair created the public GitHub repository,
pushed commit `8d96724`, published tag `v2.0.0`, uploaded six archives and
`SHA256SUMS`, and recorded successful `verify` CI. Still `unsupported`: remote
checksum fetching inside the app, automatic replacement of the running
executable, dependency audit, and platform-wide installer execution tests.

Verification evidence: `go test ./internal/release -cover` and `make verify`.

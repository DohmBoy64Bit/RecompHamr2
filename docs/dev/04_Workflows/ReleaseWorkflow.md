# Release Workflow

Release work is split into audited Phase 12 slices. Each slice must add tests,
docs, traceability, and explicit unsupported boundaries before broader release
automation is allowed.

Implemented checksum slice:

- Use `internal/release.DefaultTargets` and `ArtifactName` for canonical
  archive names.
- Use `internal/release.BuildBinary` to run local `go build -trimpath` for a
  supported target.
- Use `internal/release.CreateArchive` to package an already-built local binary.
- Use `internal/release.ManifestName` for the checksum manifest filename.
- Use `internal/release.GenerateManifest` to hash local artifacts after archive
  creation.
- Use `internal/release.ManifestText` for deterministic `SHA256SUMS` content.
- Use `internal/release.WriteManifest` to write `<releaseDir>/SHA256SUMS` or an
  explicit manifest path.
- Use `internal/release.ParseManifest` for local `SHA256SUMS` parsing.
- Use `internal/release.VerifyManifest(rootDir, manifestPath)` for local
  artifact verification.
- Use `internal/release.ValidateOperationalFiles(repoRoot)` before claiming
  Phase 12 operational files are present.
- Use `internal/update.PlanLocal` for self-update dry-runs from verified local
  artifacts.
- Keep artifact paths relative to the release directory.
- Treat malformed manifests, path traversal, missing artifacts, and checksum
  mismatches as blocked verification results.

Verified unsupported release limits:

- release downloads;
- remote checksum fetching;
- automatic replacement of the running executable;
- installer execution tests on every platform;
- dependency audit.

Phase 25 RC preparation adds local release notes and known-limits docs, but it
does not publish tags, upload artifacts, fetch remote checksums, or execute
installers on every platform.

# Install

Phase 12 provides local installers and release verification for artifacts that
already exist on disk. Remote downloads and automatic binary replacement are
not performed by these scripts.

```sh
make verify
go run ./cmd/recomphamr --diagnostic
```

## Local Installers

Windows PowerShell installer:

```powershell
.\scripts\install.ps1 -ReleaseDir .\dist -InstallDir "$env:LOCALAPPDATA\RecompHamr\bin" -Artifact recomphamr_windows_amd64.zip
```

POSIX installer:

```sh
scripts/install.sh ./dist "$HOME/.local/bin/recomphamr" recomphamr_linux_amd64.tar.gz
```

Both installers require the artifact to be present in the release directory.
They verify `SHA256SUMS` before extraction and fail on missing artifacts,
missing manifest entries, and checksum mismatches.

## Checksum Verification

The local checksum helpers exist in `internal/release` for future release
plumbing and tests. They generate and verify `SHA256SUMS`-style manifests for
files already present on disk; they do not download, install, update, or execute
anything.

Manifest format:

```text
<64 lowercase or uppercase sha256 hex>  recomphamr_windows_amd64.zip
<64 lowercase or uppercase sha256 hex> *recomphamr_linux_amd64.tar.gz
```

Rules:

- Artifact paths must be relative to the release directory.
- Absolute paths and `..` traversal are blocked.
- Generated manifests are sorted by artifact path and use lowercase SHA-256.
- Missing files, malformed hashes, and mismatches are reported as blocked
  verification results.
- A report is verified only when every manifest entry matches.

Generation helpers:

- `GenerateManifest(rootDir, artifactPaths)` hashes local artifacts under
  `rootDir`.
- `ManifestText(entries)` renders deterministic manifest text.
- `WriteManifest(rootDir, artifactPaths, manifestPath)` writes the manifest.
  When `manifestPath` is empty, the output is `<rootDir>/SHA256SUMS`.

Release downloads and remote checksum fetching remain `unsupported`.

## Self-Update Dry Run

`internal/update.PlanLocal` verifies a local artifact against a local
`SHA256SUMS` manifest and returns a dry-run plan:

```text
current: v1.0.0
candidate: v1.1.0
artifact: recomphamr_windows_amd64.zip
```

The dry run proves the candidate artifact is present and verified. It does not
replace the running executable; replacement must happen through an installer or
an explicit user action.

## Release Artifact Names

Release artifact naming is defined and tested before release automation exists.
The canonical archive names are:

```text
recomphamr_windows_amd64.zip
recomphamr_windows_arm64.zip
recomphamr_linux_amd64.tar.gz
recomphamr_linux_arm64.tar.gz
recomphamr_darwin_amd64.tar.gz
recomphamr_darwin_arm64.tar.gz
```

The checksum manifest filename is `SHA256SUMS`. These names are metadata only
until full release automation exists.

## Archive Creation

`internal/release.CreateArchive` can create local archives from an already-built
binary:

- Windows targets store `recomphamr.exe` in `.zip` archives.
- Linux and macOS targets store `recomphamr` in `.tar.gz` archives.
- Existing archive files are not overwritten.
- The function creates the output directory when needed.

Example archive inputs:

```text
target: windows/amd64
binary: dist/recomphamr.exe
output: dist/
archive: dist/recomphamr_windows_amd64.zip
```

The archive helper does not build the binary, download releases, install files,
or update an existing installation.

## Binary Builds

`internal/release.BuildBinary` can run a local `go build` for a supported target.
It writes deterministic binary names to an output directory:

```text
recomphamr_windows_amd64.exe
recomphamr_windows_arm64.exe
recomphamr_linux_amd64
recomphamr_linux_arm64
recomphamr_darwin_amd64
recomphamr_darwin_arm64
```

Defaults:

- Package: `./cmd/recomphamr`
- Output directory: `dist`
- Build flags: `go build -trimpath`
- Environment: `GOOS`, `GOARCH`, and `CGO_ENABLED=0`

Existing binary outputs are not overwritten. This helper builds only from the
local checkout; it does not package, install, download, or update releases.

## Release Config, Devcontainer, And CI

Operational files are part of Phase 12:

- `.goreleaser.yaml` defines GoReleaser builds, archives, and `SHA256SUMS`.
- `.devcontainer/devcontainer.json` uses a Go devcontainer and runs
  `make verify` after creation.
- `.github/workflows/verify.yml` runs `make verify` on Linux, Windows, and
  macOS.

`/doctor` checks these files locally and reports `verified` only when required
markers are present.

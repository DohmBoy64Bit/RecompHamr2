# Install

Phase 12 provides local installers and release verification for artifacts that
already exist on disk. Remote downloads and automatic binary replacement are
not performed by these scripts.

```sh
make verify
go run ./cmd/recomphamr --diagnostic
```

## Local Installers

Build a local Windows executable from this checkout:

```powershell
go build -trimpath -o .\dist\recomphamr.exe .\cmd\recomphamr
.\dist\recomphamr.exe --summary
.\dist\recomphamr.exe --diagnostic
.\dist\recomphamr.exe
```

`--summary` and `--diagnostic` are non-interactive smoke commands. Running the
`.exe` without flags launches the Bubble Tea terminal app and does not contact a
model backend until you submit a prompt.

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

Manual local Windows archive smoke:

```powershell
$release = "$env:TEMP\recomphamr-release"
New-Item -ItemType Directory -Force -Path $release | Out-Null
go build -trimpath -o "$release\recomphamr.exe" .\cmd\recomphamr
Compress-Archive -LiteralPath "$release\recomphamr.exe" -DestinationPath "$release\recomphamr_windows_amd64.zip"
$hash = (Get-FileHash -Algorithm SHA256 "$release\recomphamr_windows_amd64.zip").Hash.ToLowerInvariant()
Set-Content -Encoding ascii -NoNewline -Path "$release\SHA256SUMS" -Value "$hash  recomphamr_windows_amd64.zip`n"
```

Verify the manifest before installing:

```powershell
$line = Get-Content "$release\SHA256SUMS" -Raw
$parts = $line.Trim() -split '\s+', 2
$parts[0] -eq (Get-FileHash -Algorithm SHA256 (Join-Path $release $parts[1])).Hash.ToLowerInvariant()
```

The expected verification result is `True`.

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

Published `.exe` downloads remain `blocked:` until an external release page,
artifact URL, checksum URL, CI/platform run, and publication timestamp are
recorded and validated.

## Release Config, Devcontainer, And CI

Operational files are part of Phase 12:

- `.goreleaser.yaml` defines GoReleaser builds, archives, and `SHA256SUMS`.
- `.devcontainer/devcontainer.json` uses a Go devcontainer and runs
  `make verify` after creation.
- `.github/workflows/verify.yml` runs `make verify` on Linux, Windows, and
  macOS.

`/doctor` checks these files locally and reports `verified` only when required
markers are present.

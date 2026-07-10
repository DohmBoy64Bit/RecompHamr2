# Stable Release Gate

Phase 27 records stable-release readiness for the local checkout at commit
`6a095dc`, plus local artifact, checksum, Windows installer smoke, and local
stable tag decision evidence generated after commit `95269b3`. This is a local
gate record, not an uploaded or externally published stable release.

## Local Gate Result

Local stable-release readiness is satisfied when these commands pass:

```powershell
make verify
go run ./cmd/recomphamr --diagnostic
go test ./internal/release ./internal/update ./internal/doctor -cover
```

The local gate covers:

- parity documentation and closure;
- docs coverage and exported Go doc comments;
- 100% statement coverage for every package;
- architecture separation checks;
- release/update/doctor package coverage;
- diagnostic output;
- known-limits and RC soak docs.

## Local Artifact Evidence

Local release artifacts were generated under `dist/` for:

- `recomphamr_windows_amd64.zip`
- `recomphamr_windows_arm64.zip`
- `recomphamr_linux_amd64.tar.gz`
- `recomphamr_linux_arm64.tar.gz`
- `recomphamr_darwin_amd64.tar.gz`
- `recomphamr_darwin_arm64.tar.gz`

`dist/SHA256SUMS` was generated from those six archives and each manifest entry
was verified locally with `Get-FileHash -Algorithm SHA256`.

Windows installer smoke evidence:

```powershell
.\scripts\install.ps1 -ReleaseDir .\dist -InstallDir <temp-install-dir> -Artifact recomphamr_windows_amd64.zip
& <temp-install-dir>\recomphamr.exe --diagnostic
```

The smoke installed `recomphamr.exe` into a temporary directory and the
installed binary printed diagnostic output successfully.

## Phase 34 Local Executable Evidence

Phase 34 re-verified the direct end-user Windows executable path from this
checkout:

```powershell
go build -trimpath -o "$env:TEMP\recomphamr-phase34\recomphamr.exe" .\cmd\recomphamr
& "$env:TEMP\recomphamr-phase34\recomphamr.exe" --summary
& "$env:TEMP\recomphamr-phase34\recomphamr.exe" --diagnostic
```

Observed evidence:

- executable path:
  `C:\Users\SeanS\AppData\Local\Temp\recomphamr-phase34\recomphamr.exe`;
- executable size: `12109312` bytes;
- executable SHA-256:
  `782c7d9193af6368652a5862cb7e8ea39bcd3baa8994ae2898fea6eec97a6ad1`;
- `--summary` first line: `RecompHamr product runtime`;
- `--diagnostic` first line: `recomphamr diagnostic mode`;
- local archive:
  `C:\Users\SeanS\AppData\Local\Temp\recomphamr-phase34\recomphamr_windows_amd64.zip`;
- archive SHA-256:
  `93e5c0f8dbbf90b4c1423a45f239d0f085a1a4e14e91a5aa6f0fb58df8065218`;
- local `SHA256SUMS` verification result: `True`.

## Local Stable Tag Decision

The local stable tag decision is `v2.0.0`. The tag must be created only after
the release-memory docs are committed and `make verify` passes for that commit.
This tag is local evidence until it is pushed or otherwise published with the
release artifacts.

## Blocked Publication Conditions

Stable release publication remains `blocked:` until the release owner records:

- external CI or platform matrix evidence where required;
- publication destination and upload evidence;
- external artifact and checksum URLs.

No uploaded artifact, remote download, remote checksum fetch, automatic
replacement, external CI result, publication destination, or platform-wide
installer execution claim exists in this checkout.

Local inspection with `git remote -v` produced no remote output during Phase
34, so external publication remains `blocked:` until a release owner supplies
hosted evidence.

Phase 35 gate audit found local `HEAD` `6b17cf2`, local tag `v2.0.0`, 13 local
`dist/` entries, and local `dist/SHA256SUMS`, but `git remote -v` produced no
output and `gh release view v2.0.0` returned `no git remotes found`. This is
not stable publication evidence because there is still no external artifact
URL, checksum URL, CI URL, or publication timestamp.

`internal/release.ValidatePublicationEvidence` validates the required
publication fields without claiming an upload: version, commit, external CI URL,
external artifact URL, external checksum URL, and publication timestamp. Local
paths, localhost URLs, empty values, and missing timestamps are reported as
`blocked`.

## Corrective Runtime, MCP, And Post-Parity Feature Gate

Phase 28 is the corrective live end-user runtime integration phase. Phase 29 is
the corrective live MCP agent integration phase. Phases 30-34 are corrective
TUI and Windows executable hardening. Post-parity feature intake moves to Phase
35 and remains blocked until Phase 28, Phase 29, corrective TUI hardening,
local `.exe` launch polish, and stable publication evidence are recorded. Local
readiness alone does not open feature planning. The Phase 35 gate audit did not
open feature intake because publication evidence is still missing.

## Release Owner Checklist

Before publishing:

1. Re-run the local gate commands from a clean checkout.
2. Generate local release artifacts.
3. Generate and verify `SHA256SUMS`.
4. Record platform/install evidence.
5. Create the stable tag intentionally.
6. Publish artifacts and checksums.
7. Validate publication evidence with `internal/release`.
8. Update `KnownLimits.md`, `StatusReports.md`, and this file with publication
   evidence.

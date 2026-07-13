# Stable Release Gate

Phase 27 records stable-release readiness and publication evidence for
RecompHamr 2.0.0. The current published tag points at commit `8d96724`, and the
GitHub release is public at
`https://github.com/DohmBoy64Bit/RecompHamr2/releases/tag/v2.0.0`.

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

## Phase 62 Replacement TUI Executable Evidence

The complete Bubble Tea replacement was built and smoked locally from this
checkout:

- executable path: `E:\RecompHamr2\dist\recomphamr.exe`;
- executable size: `14183424` bytes;
- build timestamp: `2026-07-13T16:04:53.5790699Z`;
- executable SHA-256:
  `dc2a69e4e5ca64989a3aa9aae20f9072f30ded7ebb636c26caf9e1347daf98ab`;
- `--summary` first line: `RecompHamr product runtime`;
- `--diagnostic` first line: `recomphamr diagnostic mode`.

This is local replacement-build evidence, not a new hosted release. Manual
WezTerm visual acceptance remains pending in
`Phase62TUIAcceptanceEvidence.md`; the published `v2.0.0` evidence below is
unchanged.

This metadata supersedes the earlier Phase 62 local binary, which panicked when
a terminal blur event preceded its first frame. The replacement has focused
startup/chat and app-boundary regression coverage for that message order.

## Stable Publication Evidence

External publication is verified for `v2.0.0`:

- repository: `https://github.com/DohmBoy64Bit/RecompHamr2`;
- remote: `origin https://github.com/DohmBoy64Bit/RecompHamr2.git`;
- published commit: `8d96724`;
- tag: `v2.0.0`;
- release URL:
  `https://github.com/DohmBoy64Bit/RecompHamr2/releases/tag/v2.0.0`;
- CI URL:
  `https://github.com/DohmBoy64Bit/RecompHamr2/actions/runs/29096150903`;
- additional successful verify run:
  `https://github.com/DohmBoy64Bit/RecompHamr2/actions/runs/29096149468`;
- publication timestamp: `2026-07-10T13:19:48Z`;
- checksum URL:
  `https://github.com/DohmBoy64Bit/RecompHamr2/releases/download/v2.0.0/SHA256SUMS`;
- primary Windows artifact URL:
  `https://github.com/DohmBoy64Bit/RecompHamr2/releases/download/v2.0.0/recomphamr_windows_amd64.zip`.

Uploaded artifacts:

- `recomphamr_windows_amd64.zip`
- `recomphamr_windows_arm64.zip`
- `recomphamr_linux_amd64.tar.gz`
- `recomphamr_linux_arm64.tar.gz`
- `recomphamr_darwin_amd64.tar.gz`
- `recomphamr_darwin_arm64.tar.gz`
- `SHA256SUMS`

## Local Stable Tag Decision

The stable tag is `v2.0.0`. The tag was force-updated from local evidence onto
commit `8d96724` after the portable CI test fix and pushed to `origin`.

## Blocked Publication Conditions

The previous publication blocker is resolved for `v2.0.0`: external
repository, release, artifact, checksum, CI, and publication timestamp evidence
now exist. Still unsupported: automatic replacement of the running executable,
remote checksum fetching inside the app, dependency audit, and installer
execution tests on every platform.

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
local `.exe` launch polish, and stable publication evidence are recorded. The
publication evidence gate is now satisfied for `v2.0.0`; Phase 35 feature
intake may open only through its documented intake criteria.

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

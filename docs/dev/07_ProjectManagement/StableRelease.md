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

## Local Stable Tag Decision

The local stable tag decision is `v2.0.0`. The tag must be created only after
the release-memory docs are committed and `make verify` passes for that commit.
This tag is local evidence until it is pushed or otherwise published with the
release artifacts.

## Blocked Publication Conditions

Stable release publication remains `blocked:` until the release owner records:

- external CI or platform matrix evidence where required;
- publication destination and upload evidence.

No uploaded artifact, remote download, remote checksum fetch, automatic
replacement, external CI result, publication destination, or platform-wide
installer execution claim exists in this checkout.

## Post-Parity Feature Gate

Post-parity feature intake remains blocked until stable publication evidence is
recorded. Local readiness alone does not open Phase 28 feature planning.

## Release Owner Checklist

Before publishing:

1. Re-run the local gate commands from a clean checkout.
2. Generate local release artifacts.
3. Generate and verify `SHA256SUMS`.
4. Record platform/install evidence.
5. Create the stable tag intentionally.
6. Publish artifacts and checksums.
7. Update `KnownLimits.md`, `StatusReports.md`, and this file with publication
   evidence.

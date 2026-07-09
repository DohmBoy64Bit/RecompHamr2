# Release Candidate Notes

This checkout has release-candidate preparation docs, but no published RC tag or
uploaded artifacts.

## What Is Ready Locally

- `make verify` is the canonical verification gate.
- Local release metadata, archive naming, checksum manifest generation, and
  checksum verification are implemented.
- Installer scripts verify local artifacts against `SHA256SUMS`.
- The fresh-clone walkthrough and RecompHamr 1.x migration notes are available.

## What Is Not Published

The following remain `unsupported` until an explicit RC cut:

- remote downloads;
- remote checksum fetching;
- automatic self-replacement;
- installer execution claims on every platform;
- published release artifacts.

## Local Verification

```powershell
make verify
go run ./cmd/recomphamr --diagnostic
```

Use `docs/dev/07_ProjectManagement/ReleaseCandidate.md` for the release manager
checklist and `docs/dev/07_ProjectManagement/KnownLimits.md` for verified
limits.

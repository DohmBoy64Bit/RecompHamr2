# Stable Release Gate

Phase 27 records stable-release readiness for the local checkout at commit
`6a095dc`. This is a local gate record, not a published stable release.

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

## Blocked Publication Conditions

Stable release publication remains `blocked:` until the release owner records:

- an intentional stable tag decision;
- generated release artifacts for supported targets;
- `SHA256SUMS` generated from those artifacts;
- checksum verification of the generated artifacts;
- fresh-install validation evidence;
- external CI or platform matrix evidence where required;
- publication destination and upload evidence.

No published stable release, uploaded artifact, remote download, remote checksum
fetch, automatic replacement, or platform-wide installer execution claim exists
in this checkout.

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

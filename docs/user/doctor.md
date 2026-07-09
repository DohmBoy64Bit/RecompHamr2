# Doctor

`/doctor` reports deterministic local diagnostics without network probes,
external process launches, or workspace mutation.

Run it from the slash-command registry:

```text
/doctor
```

The report uses explicit evidence labels:

- `verified`: proven from local runtime, source registrations, or workspace files.
- `unsupported`: intentionally outside the current phase or not initialized.
- `blocked`: local state prevented the check, such as unreadable config or skill files.

Current checks:

- Runtime: Go version, OS, and architecture.
- Workspace: whether `.rehamr/` exists and is a directory.
- Config: active model and profile count from `.rehamr/config.yaml`.
- Memory: `.rehamr/REPHAMR_STATE.md` path, byte count, and truncation status.
- Skills: embedded and custom skill counts.
- Tools: primary and compatibility tool counts.
- MCP: built-in server count and autostart count from environment settings.
- Install/update/release: required operational files for installers,
  GoReleaser, devcontainer, and CI.

Example output:

```text
RecompHamr doctor
[verified] runtime: go=go1.26.4 os=windows arch=amd64
[verified] tools: primary=6 compatibility=1
[verified] install-update-release: 5 operational files verified
```

The executable diagnostic mode remains available:

```sh
go run ./cmd/recomphamr --diagnostic
```

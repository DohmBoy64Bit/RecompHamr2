# Doctor Parity

Doctor parity covers the `/doctor` slash command and local operational checks.

Implemented Phase 12 slice:

- `internal/doctor.Run` collects deterministic local diagnostics without
  network probes, external process launches, or workspace mutation.
- Runtime reports Go version, OS, and architecture.
- Workspace reports whether `.rehamr/` exists and is a directory.
- Config reports active model and profile count from `.rehamr/config.yaml`.
- Memory reports `.rehamr/REPHAMR_STATE.md`, byte count, and truncation status.
- Skills, tools, and MCP checks report registered local counts.
- Install/update/release checks validate required operational files for
  installers, GoReleaser, devcontainer, and CI.
- `/doctor` command dispatch renders the doctor report through
  `internal/commands`.

Status labels:

- `verified`: local runtime, source registration, or workspace evidence exists.
- `unsupported`: the workspace/config/memory is not initialized.
- `blocked`: local files or directories prevent the check from completing.

Still `unsupported`: network release downloads, remote checksum fetching, and
installer execution tests on every platform.

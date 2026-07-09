# Security Audit

Next 20 Phase 20 audits implemented security boundaries against local source,
tests, and docs. The audit scope is limited to current parity behavior and does
not approve post-parity features.

## Audit Results

| Boundary | Status | Evidence |
|---|---|---|
| Config directory and file symlinks | passed | `internal/config.Bootstrap` refuses symlinked `.rehamr/` and `config.yaml`; config tests cover unsafe paths. |
| Workspace memory and status symlinks | passed | `internal/project.LoadMemory` and `project.Status` use `Lstat` and refuse symlinked `.rehamr/`; `LoadMemory` also refuses symlinked `REPHAMR_STATE.md`; project tests cover these regressions. |
| Owner-only generated files | passed | Config saves use `0600`; workspace files use `0600`; workspace directories use `0700`; tests check POSIX permissions where applicable. |
| Atomic config writes | passed | `Config.Save` writes `config.yaml.tmp`, chmods it, then renames over `config.yaml`; save failure tests cover cleanup paths. |
| Shell command execution | passed for current tool surface | `powershell` uses `-NoProfile -NonInteractive`, bounded timeout, cancellation checks, and explicit failure text; tool tests cover timeout, cancellation, and compatibility alias behavior. |
| File tools | passed for current scope | `read_file` caps output at 1 MiB; `write_file` and `edit_file` use owner-only writes and explicit failures; tool tests cover empty, missing, ambiguous, and write failures. |
| Reference network fetches | passed for current scope | `recomp_reference` accepts only `http` and `https`, writes sanitized cache names under the output directory, and returns explicit fetch/read/status errors. |
| Repository cache | passed for current scope | `repomixr` accepts only `github.com/owner/repo` style URLs and keeps clone targets inside the configured cache directory. |
| MCP execution | partial with verified limits | MCP manager registration, streamable HTTP, skill gating, allowlists, and command dispatch are tested. Stdio process spawning, autoconnect, persistent user MCP config, and agent-loop MCP exposure remain unsupported. |
| Release verification | partial with verified limits | Local build/archive/checksum/install-script/dry-run behavior is tested. Remote downloads, remote checksum fetching, automatic binary replacement, and platform installer execution remain unsupported. |
| Secret redaction | passed for current scope | `internal/security`, `internal/logging`, and `internal/tui` tests cover configured secret redaction in logs and debug rendering. |
| Product startup | passed for current scope | Bare startup does not call a model backend, execute tools, autoconnect MCP servers, or launch a terminal process; app tests and diagnostic output cover this boundary. |

## Verification

Phase 20 closure requires:

- `go test ./internal/project -cover`
- `make verify`
- `go run ./cmd/recomphamr --diagnostic`
- Security keyword scan over code and docs
- Placeholder-policy scan limited to policy text
- Documentation hash evidence for changed security and project docs

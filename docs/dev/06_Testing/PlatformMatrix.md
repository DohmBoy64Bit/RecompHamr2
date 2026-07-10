# Platform Matrix

Phase 22 validates platform behavior from local source, docs, and tests. Windows
is the primary target; Linux and macOS are validated where current code can do
so without pretending installer execution or live terminal behavior.

## Validation Areas

| Area | Windows evidence | Linux/macOS evidence | Current limit |
|---|---|---|---|
| Shell tool | `internal/tools` selects `powershell -NoProfile -NonInteractive -Command` when `GOOS=windows`; tests force the Windows branch. | `internal/tools` selects `pwsh` for the `powershell` schema when `GOOS` is not Windows; tests force the non-Windows branch. | Real child-process tree termination beyond Go `CommandContext` behavior is unsupported. |
| Compatibility shell alias | `bash` remains a compatibility alias and maps to the Windows PowerShell path. | `bash` uses the documented shell compatibility path on non-Windows. | The primary documented schema remains `powershell`. |
| Paths and permissions | Workspace and config tests use `Lstat` symlink refusal and owner-only mode requests. Windows does not expose POSIX mode guarantees. | Owner-only `0o700` directories and `0o600` files are requested and checked where POSIX modes are meaningful. | Union coverage across real Linux/macOS CI is required before stable release. |
| Install scripts | `scripts/install.ps1` is marker-validated by `internal/release.ValidateOperationalFiles` and documented in `docs/user/install.md`. | `scripts/install.sh` is marker-validated by the same release check and documented in `docs/user/install.md`. | Executing installers on every platform is unsupported in this checkout. |
| Release artifacts | `internal/release.DefaultTargets` includes `windows/amd64` and `windows/arm64`; archive tests verify `.zip` contents include `recomphamr.exe`. | Targets include `linux` and `darwin` for `amd64` and `arm64`; archive tests verify `.tar.gz` contents include `recomphamr`. | Remote downloads and published artifact verification remain unsupported. |
| TUI rendering and launch | `internal/tui` pure render tests cover compact/wide layouts; `internal/app` tests verify bare CLI launch wiring and live wrapper behavior without requiring an attached terminal. | The same pure render tests apply because rendering is string based; real terminal attachment on Linux/macOS still requires CI smoke evidence. | Platform-specific terminal attachment claims outside local Windows remain unverified until external CI evidence exists. |
| Cancellation | Agent, LLM, MCP, app smoke, and tool tests cover Go context cancellation and timeout reporting. | Same code paths are covered by platform-independent tests. | OS-specific process-group cleanup is not yet claimed. |

## Required Stable-Release Evidence

- Windows, Linux, and macOS CI jobs must run `make verify`.
- Final release coverage must use a union report across platform jobs.
- Installer execution claims require platform-specific smoke evidence before
  they can move from `unsupported` to verified.
- Any platform-specific behavior change must update this matrix, user docs,
  release docs, and the traceability matrix in the same change.

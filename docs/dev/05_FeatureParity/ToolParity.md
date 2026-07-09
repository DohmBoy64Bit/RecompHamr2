# Tool Parity

RecompHamr 1.x evidence captured the built-in tools `bash`, `read_file`,
`write_file`, `edit_file`, `repomixr`, and `recomp_reference`.

Phase 5 implements the same tool families with one intentional Windows-focused
interface change: the primary shell schema is `powershell`. The 1.x `bash` name
remains as a compatibility alias and maps to PowerShell on Windows. This avoids
requiring a Unix shell for the Windows-first rewrite while preserving parity
handling for historical prompts and embedded references.

Parity rows:

| 1.x behavior | 2.0 behavior | Evidence |
|---|---|---|
| `bash` schema ran `/bin/sh -c` with timeout and cancellation. | Primary schema is `powershell`; `bash` remains a compatibility alias. | `internal/tools`, `docs/user/tools.md`, `tools` package tests. |
| `read_file` returned file text. | Returns file text with a 1 MiB context-flood limit. | `internal/tools` tests. |
| `write_file` created parents and wrote content. | Creates parents and writes owner-only files where supported. | `internal/tools` tests. |
| `edit_file` required one exact replacement. | Empty, missing, identical, and ambiguous edits fail. | `internal/tools` tests. |
| `repomixr` accepted GitHub references. | Accepts only contained `github.com/owner/repo` cache targets. | `internal/tools` tests. |
| `recomp_reference` cached web references. | Accepts only `http` or `https` URLs and contained cache paths. | `internal/tools` tests. |

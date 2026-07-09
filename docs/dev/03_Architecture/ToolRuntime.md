# Tool Runtime

Tools expose documented schemas, explicit validation, timeout and cancellation
behavior, cache path containment, examples, and tests before runtime exposure.

## Primary Tools

- `powershell`: Windows-first shell execution. It runs `powershell -NoProfile -NonInteractive -Command` on Windows and `pwsh` on non-Windows platforms. The default timeout is 120 seconds and the maximum timeout is 1 hour.
- `read_file`: returns file content and truncates output after 1 MiB with a marker.
- `write_file`: creates parent directories with owner-only permissions and writes files with owner-only permissions where supported.
- `edit_file`: performs one exact replacement only; empty paths, empty `old_string`, identical replacement text, missing matches, and multiple matches fail.
- `repomixr`: accepts only `github.com/owner/repo` URLs and writes cloned repositories inside the configured output directory.
- `recomp_reference`: accepts only `http` and `https` URLs and writes sanitized cache files inside the configured output directory.

## Compatibility Alias

RecompHamr 1.x exposed `bash`. RecompHamr 2.0 keeps `bash` as a compatibility
alias for parity, but it is not the primary schema. On Windows, the alias maps
to the PowerShell execution path so Windows users are not required to install a
Unix shell.

## Failure Contract

Tool failures are user-visible strings beginning with `(` or containing exit,
timeout, or cancellation markers. Callers must treat these as failures and must
not convert them into fake success paths.

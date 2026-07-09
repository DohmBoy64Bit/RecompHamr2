# Tools

Built-in tool foundations are implemented and test-covered. The primary tool
schemas are:

- `powershell`: run a Windows PowerShell command with timeout and cancellation. Use `cmd` for the command text and `timeout_seconds` for an optional limit. Example: `powershell {"cmd":"Get-ChildItem .","timeout_seconds":30}`.
- `read_file`: read file contents from `path`. Results larger than 1 MiB are truncated with an explicit marker.
- `write_file`: write `content` to `path`, creating parent directories with owner-only permissions where the platform supports them.
- `edit_file`: replace exactly one `old_string` in `path` with `new_string`; ambiguous or missing matches fail.
- `repomixr`: clone a `github.com` repository `url` into `output_dir`. Only two-segment `owner/repo` paths are accepted.
- `recomp_reference`: fetch an `http` or `https` `url` into `output_dir` for offline reading.

Compatibility: RecompHamr 1.x exposed `bash`. RecompHamr 2.0 keeps `bash` as
a legacy alias for parity, but new Windows-focused behavior should use
`powershell`.

Failures return explicit tool-style strings beginning with `(` or containing
exit, timeout, or cancellation markers.

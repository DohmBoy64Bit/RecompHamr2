# Security And Sandboxing

Security rules are mandatory from the start. Sensitive files require owner-only permissions, atomic writes, symlink refusal, path traversal rejection, bounded command execution, and redaction.

Phase 3 config and workspace controls:

- `.rehamr/` must be a real directory, not a symlink.
- `.rehamr/config.yaml` must be a real file, not a symlink.
- `.rehamr/REPHAMR_STATE.md` must be a real file when loaded for prompt memory;
  direct memory reads reject symlinked state files.
- Config saves use an adjacent temporary file and rename to avoid torn writes.
- Config files and generated workspace files are written owner-only on POSIX systems.
- `RECOMPHAMR_URL` is process-local and never persisted to disk.

Phase 5 shell and file tool controls:

- `powershell` is the primary shell tool. It uses `-NoProfile -NonInteractive`, defaults to a 120 second timeout, and caps timeouts at 1 hour.
- The legacy `bash` name remains a compatibility alias and maps to PowerShell on Windows.
- `read_file` truncates output after 1 MiB to avoid context flooding.
- `write_file` and `edit_file` use owner-only file permissions where supported.
- `repomixr` accepts only `github.com/owner/repo` URLs and keeps clone targets inside the configured cache directory.
- `recomp_reference` accepts only `http` or `https` URLs and keeps sanitized cache files inside the configured cache directory.

Current security-sensitive runtime behavior:

- Startup does not call a model or execute tools until a user submits a prompt
  or command.
- MCP tool exposure in the agent loop is limited to connected, enabled tools
  unlocked by active skills.
- MCP autostart is limited to server configs with explicit autostart metadata.
- Stdio MCP process spawning is supported only through configured server
  commands and runs with the user's local permissions.
- Persistent `.rehamr/mcp.json` config uses strict JSON and blocks unknown keys.
- Remote release downloads, remote checksum fetching, automatic binary
  replacement, and platform installer execution tests.

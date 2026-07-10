# RecompHamr 1.x Migration Notes

RecompHamr 2.0 is a clean Go rewrite. It preserves observable RecompHamr 1.x
behavior through parity rows and tests, but it does not copy 1.x internals 1:1.

## What Carries Forward

- The 11 slash commands are preserved: `/clear`, `/models`, `/skills`, `/skill`,
  `/skill-audit`, `/skill-new`, `/init-re`, `/status-re`, `/doctor`, `/mcp`, and
  `/help`.
- The 28 embedded reverse-engineering skills are present.
- `.rehamr/` remains the project workspace and memory root.
- `REPHAMR_STATE.md` remains the persistent project memory file.
- OpenAI-compatible model profile concepts remain in `.rehamr/config.yaml`.
- MCP server names from 1.x remain registered and skill-gated.

## Intentional Interface Updates

- `powershell` is the primary shell tool because RecompHamr 2.0 is
  Windows-first. `bash` remains as a compatibility alias for 1.x parity.
- Runtime startup now launches the terminal app. It does not call a model or
  execute tools until a prompt or command requests that behavior.
- MCP autostart is explicit and can use streamable HTTP URLs or configured
  stdio commands from `.rehamr/mcp.json`.
- `/skill-new` fetches and caches skill Markdown for review. It does not
  silently install or activate custom skills.
- Release helpers verify local files and `SHA256SUMS`; remote release downloads
  and automatic binary replacement remain `unsupported`.

## Migration Checklist

1. Keep your RecompHamr 1.x project backup unchanged.
2. Run `make verify` in the RecompHamr 2.0 checkout.
3. Run `go run ./cmd/recomphamr` to create `.rehamr/config.yaml` if needed.
4. Copy only reviewed memory and evidence text into the new `.rehamr/` files.
5. Recreate model profiles in `.rehamr/config.yaml`; verify `active`,
   `models.<name>.llm`, `models.<name>.url`, `models.<name>.key`, and
   `models.<name>.context_size`.
6. Copy custom skills into `.rehamr/skills/*.md`, then run `/skills` and
   `/skill <name>`.
7. Run `/status-re` and `/doctor` before relying on the migrated workspace.

## Verified Limits

Remote release downloads and installer execution tests on every platform remain
`unsupported` in this checkout. Stdio MCP process spawning and persistent user
MCP config are implemented but require explicit local configuration.

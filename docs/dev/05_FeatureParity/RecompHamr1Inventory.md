# RecompHamr 1.x Inventory

Reference source: `https://github.com/DohmBoy64Bit/RecompHamr`

Inspected commit: `259a450e93af48437ee23663e5ca66cdc1ab8569`

## Inventoried Surface

- Slash commands: `/help`, `/clear`, `/models`, `/skills`, `/skill`, `/skill-audit`, `/skill-new`, `/init-re`, `/status-re`, `/doctor`, `/mcp`.
- Built-in tools: `bash`, `read_file`, `write_file`, `edit_file`, `repomixr`, `recomp_reference`.
- MCP servers: `ghidra`, `n64-debug-mcp`, `pcrecomp`, `mcp-pine`, `objdiff`, `pcsx2`, `bizhawk`, `sega2asm`.
- Embedded skills: 28 markdown skills imported as parity data under `internal/skills/embedded/`.
- Config files: `.rehamr/config.yaml`, `.rehamr/mcp.json`, `.rehamr/REPHAMR_STATE.md`.
- Release and install evidence: reference repo contains `Makefile`, `.goreleaser.yaml`, `install.sh`, `install.cmd`, and release workflow.

The durable inventory is encoded in `internal/parity` and covered by tests.

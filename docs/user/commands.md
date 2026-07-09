# Commands

Bare startup composes local runtime state:

```sh
recomphamr
```

It loads config, optional memory, slash commands, MCP manager state, and pure
TUI state, then prints a launch summary without making model calls.

Diagnostic command-line mode is also supported:

```sh
recomphamr --diagnostic
```

The slash command registry is implemented and wired into the composed runtime
environment. Deterministic fake-runtime smoke covers slash command dispatch
through the TUI model. Real backend prompt/model turns remain unsupported.

## Slash Command Reference

The registry implements the RecompHamr 1.x parity set:

| Command | Usage | Side Effects |
|---|---|---|
| `/clear` | `/clear` | Clears transient conversation state. |
| `/models` | `/models [name]` | Lists profiles or updates `.rehamr/config.yaml`. |
| `/skills` | `/skills` | Reads embedded and custom skill directories. |
| `/skill` | `/skill <name>` | Updates active skill state for the current session. |
| `/skill-audit` | `/skill-audit <name>` | Classifies a skill name into a template category. |
| `/skill-new` | `/skill-new <url>` | Fetches HTTP(S) skill Markdown and caches `.rehamr/fetched/<name>.md` for review. |
| `/init-re` | `/init-re` | Creates `.rehamr/` config, memory, MCP, evidence, and ledger files. |
| `/status-re` | `/status-re` | Reads `.rehamr/` state and reports missing tracked files. |
| `/doctor` | `/doctor` | Reads local runtime, workspace, config, memory, skill, tool, MCP registration, and install/update/release operational file state. |
| `/mcp` | `/mcp [connect\|disconnect\|tools\|enable\|disable] <server> [tool]` | Lists built-ins; uses MCP manager wiring for lifecycle and tool mutation. |
| `/help` | `/help [command]` | Shows generated command help. |

Examples:

```text
/help /models
/models ollama-amd
/skill ghidra-mcp
/skill-audit n64-debug-mcp
/skill-new https://example.com/SKILL.md
/init-re
/mcp connect ghidra
/mcp tools ghidra
/mcp disable ghidra decompile
/doctor
```

Non-success output is explicit: `usage:` for malformed commands,
`unsupported:` for intentionally unavailable phase behavior, `unverified:` for
missing evidence or unknown names, and `blocked:` for local failures such as
config, workspace, fetch, or cache-write errors. `/skill-new` does not silently
install or activate skills; review the cached file and copy approved content to
`.rehamr/skills/<name>.md` before loading it with `/skill <name>`.

`/mcp` lifecycle commands require MCP manager wiring. Without a manager they
return `unsupported:`; with a manager, lifecycle and mutation failures return
`blocked:`.

`/doctor` is offline and non-mutating. It reports `verified`, `unsupported`, and
`blocked` sections. Phase 12 operational file checks cover installer scripts,
GoReleaser, devcontainer, and CI workflow files.

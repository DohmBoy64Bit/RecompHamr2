# Skills

RecompHamr embeds the 28 RecompHamr 1.x reverse-engineering skills:

`bizhawk`, `build-fix-loop`, `cdb-debug`, `core-re`, `evidence-mode`, `file-format-reversing`, `function-discovery`, `gb-recomp`, `gc-decomp`, `gen-decomp`, `ghidra-mcp`, `imhex`, `mcp-pine`, `n64-debug-mcp`, `n64-decomp`, `objdiff`, `pcrecomp`, `pcsx2`, `project-handoff`, `ps2recomp`, `ps3recomp`, `recomp-foundations`, `sega2asm`, `snesrecomp`, `vb-decomp`, `windows-game-decomp`, `xbox360-decomp`, and `xboxrecomp`.

Custom skills load from `.rehamr/skills/*.md`. A custom skill with the same name as an embedded skill overrides the embedded copy for that workspace.

## Resolution

Skill names resolve by exact name, case-insensitive name, or a `.md` suffix:

```text
/skill ghidra-mcp
/skill GHIDRA-MCP.md
```

Unknown names return `unverified:`. Custom directory read failures return `blocked:`.

## Listing And Active State

`/skills` reports embedded, custom, and active counts. The lower-level skill list renderer marks active skills with `*` and labels custom overrides with `(custom)`.

## Skill Audit

`/skill-audit <name>` keeps the RecompHamr 1.x name classifier:

- names containing `mcp` or `debug`: `runtime-integration`
- names containing `recomp` or `decomp`: `reverse-engineering-workflow`
- all other names: `methodology`

The Phase 9 draft classifier also scores fetched bodies as `full_workflow`, `micro_skill`, `tool_bridge`, or `none`.

## Skill-New Workflow

`/skill-new <url>` accepts only `http` and `https` URLs. It fetches Markdown, classifies it, caches the reviewed source under `.rehamr/fetched/<name>.md`, and reports the manual approval target `.rehamr/skills/<name>.md`.

Example:

```text
/skill-new https://example.com/SKILL.md
```

The command does not silently activate or install the custom skill. After reviewing the cached file, write the approved Markdown to `.rehamr/skills/<name>.md`, run `/skills`, then activate it with `/skill <name>`.

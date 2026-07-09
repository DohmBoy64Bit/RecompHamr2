# Skill Parity

Skill parity requires all 28 embedded skills from the workflow plus custom skill behavior, resolver behavior, audit behavior, and skill-new behavior.

## Embedded Inventory

`internal/skills/embedded/` contains the 28 RecompHamr 1.x skill names recorded in the workflow:

`bizhawk`, `build-fix-loop`, `cdb-debug`, `core-re`, `evidence-mode`, `file-format-reversing`, `function-discovery`, `gb-recomp`, `gc-decomp`, `gen-decomp`, `ghidra-mcp`, `imhex`, `mcp-pine`, `n64-debug-mcp`, `n64-decomp`, `objdiff`, `pcrecomp`, `pcsx2`, `project-handoff`, `ps2recomp`, `ps3recomp`, `recomp-foundations`, `sega2asm`, `snesrecomp`, `vb-decomp`, `windows-game-decomp`, `xbox360-decomp`, and `xboxrecomp`.

## Implemented Behavior

| Behavior | Status | Evidence |
|---|---|---|
| Embedded skill bundle | implemented | `Embedded()` test requires exactly 28 non-empty skills. |
| Custom skill discovery | implemented | `LoadCustom()` reads `.md` files only, ignores directories and other extensions, and treats a missing custom directory as empty. |
| Custom precedence | implemented | `Resolve()` prepends custom skills so `.rehamr/skills/ghidra-mcp.md` overrides the embedded `ghidra-mcp`. |
| Resolution forms | implemented | Exact, case-insensitive, and `.md`-suffixed names are tested. |
| Listing and active state | implemented | `ListMarkdown()` marks active names with `*` and labels custom skills with `(custom)`. |
| Name audit | implemented | `Audit()` preserves runtime-integration, reverse-engineering-workflow, and methodology categories. |
| Skill-new classification | implemented | `Classify()` scores fetched content as `full_workflow`, `micro_skill`, `tool_bridge`, or `none`. |
| Skill-new draft | implemented | `/skill-new` fetches HTTP(S), caches `.rehamr/fetched/<name>.md`, reports `.rehamr/skills/<name>.md`, and requires manual approval before activation. |

## Security And Boundaries

`/skill-new` accepts only `http` and `https` URLs. It does not silently write approved custom skills, activate skills, or execute fetched content. Failed fetches and cache writes return `blocked:`; invalid URLs or too-short fetched bodies return `unverified:`.

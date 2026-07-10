# RecompHamr 2.0

RecompHamr 2.0 is a clean Go rewrite of RecompHamr 1.x. Its mission is a local-first terminal coding agent for reverse engineering, decompilation, and static recompilation support.

The rewrite is under a strict feature-parity freeze. New features are blocked until RecompHamr 1.x observable behavior is inventoried, documented, parity-tested, and traceable.

## Current Status

The repository has a usable local terminal runtime while remaining under the
feature-parity freeze:

- Phase 0 inventory uses `https://github.com/DohmBoy64Bit/RecompHamr` at commit `259a450e93af48437ee23663e5ca66cdc1ab8569`.
- Phase 1 governance and durable docs memory are present.
- Phase 2 diagnostic-only Go skeleton is present.
- Phase 3 config and `.rehamr/` workspace parity are implemented as independently testable packages.
- Phase 4 LLM streaming and context packing parity are implemented as independently testable packages.
- Phase 5 built-in tools are implemented with `powershell` as the primary Windows-first shell tool and `bash` retained only as a 1.x compatibility alias.
- Phase 6 agent loop and Phase 7 TUI shell foundations are implemented as independently testable packages.
- Phase 8 slash command parity is implemented as an independently testable package.
- Phase 9 skill foundations are implemented as independently testable packages.
- Product runtime wiring launches the Bubble Tea terminal app, dispatches slash
  commands, sends prompts through the active OpenAI-compatible model profile,
  routes built-in and connected skill-visible MCP tool calls through the agent
  loop, supports cancellation, and keeps deterministic fake-runtime smoke
  coverage.

## Verify

```sh
make verify
```

## User Walkthrough

Fresh-clone setup and migration notes live in:

- `docs/user/walkthrough.md`
- `docs/user/migration.md`

## Diagnostic Mode

```sh
go run ./cmd/recomphamr --diagnostic
```

Help is available with:

```sh
go run ./cmd/recomphamr --help
```

This mode reports foundation status only. Bare startup launches the terminal app:

```sh
go run ./cmd/recomphamr
```

Startup does not call a model backend. MCP autostart is explicit and limited to
URL-backed server configs. Use `go run ./cmd/recomphamr --summary` to print the
deterministic runtime summary without opening the TUI.

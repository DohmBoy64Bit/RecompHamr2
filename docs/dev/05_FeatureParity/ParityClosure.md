# Parity Closure

Next 20 Phase 19 audits the parity matrix against local source, tests, and
developer docs. This file is the closure ledger for the current implementation
state. It does not approve new features.

## Closure Rule

Each parity area must be one of:

- `passed`: implemented, documented, tested, and traceable for the stated
  scope.
- `partial`: implemented for local, deterministic behavior, with verified
  remaining limits listed.
- `blocked`: not implementable without missing evidence, external execution, or
  a later approved phase.

Unsupported behavior must stay user-visible as `unsupported:`, `unverified:`,
or `blocked:` rather than pretending success.

## Audit Results

| Area | Closure | Evidence |
|---|---|---|
| Phase 0 source inventory | passed | Reference commit `259a450e93af48437ee23663e5ca66cdc1ab8569`, `internal/parity`, `RecompHamr1Inventory.md`, `make verify`. |
| Governance docs | passed | `AGENTS.md`, workflow docs, docscheck, traceability rows. |
| Diagnostic skeleton | passed | `cmd/recomphamr`, `internal/app`, diagnostic tests, diagnostic CLI output. |
| Config/workspace and memory | partial | Config, workspace bootstrap, memory loading, agent injection, startup status, and fake-runtime memory smoke are covered. Real backend memory consumption remains unsupported. |
| LLM/context helpers | passed | `internal/llm`, streaming/provider/context tests, protocol and context docs. |
| Built-in tools | passed | `internal/tools`, schemas, Windows-first `powershell`, `bash` compatibility alias, tool docs, security tests. |
| Agent loop | passed | `internal/agent` fake-model tests for tool pairing, nudges, cancellation, and round limits. |
| TUI shell contract | partial | Pure TUI model and Bubble Tea adapter are covered. Launching the live Bubble Tea process loop remains unsupported. |
| Slash commands | passed | All 11 parity commands have registry metadata, generated help, docs, examples, side effects, and tests. |
| Skills | passed | All 28 embedded skills, custom precedence, resolution, active listing, audit, and skill-new fetch/cache workflow are covered. |
| MCP protocol foundation | passed | JSON-RPC, stdio injected-stream transport, streamable HTTP transport, payloads, and fake transport tests are covered. |
| MCP manager runtime | partial | Manager lifecycle, streamable HTTP connector, tool gating, and `/mcp` command dispatch are covered. Stdio process spawning, app autoconnect, persistent user MCP config files, and agent-loop MCP exposure remain unsupported. |
| Doctor diagnostics | passed | Offline diagnostics and operational file validation are covered. |
| Release verification | partial | Local build/archive/checksum/install-script/dry-run behavior is covered. Remote downloads, remote checksum fetching, automatic binary replacement, and platform installer execution tests remain unsupported. |
| Product runtime | partial | Bare startup and deterministic fake-runtime smoke are covered. Real backend prompt turns, real tool execution, MCP autoconnect, and Bubble Tea process launch remain unsupported. |

## Verification

Closure requires:

- `make verify`
- `go run ./cmd/recomphamr --diagnostic`
- Stale parity wording scan for obsolete "remains next phase" claims
- Placeholder-policy scan limited to policy text
- Documentation hash evidence for changed parity and status docs

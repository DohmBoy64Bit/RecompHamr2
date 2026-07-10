# Next 20 Phase Plan

This document preserves and reconciles the user-approved "RecompHamr 2.0 Next
20 Phases" plan that continues after completed Phase 9. It is durable project
memory. Agents must read it when selecting work after Phase 12.

## Reconciliation Rule

The original workflow in `recomphamr_2_rewrite_workflow.md` remains the
historical source for phases 0 through 15. This plan is the forward execution
roadmap for post-Phase-9 work, but some phase numbers differ because several
large workflow phases were split into smaller implementation phases.

When names conflict, use the phase title and scope, not the number alone.
Record both names in goal packets, traceability rows, and status reports.

## Current Mapping

| Next 20 phase | Title | Workflow mapping | Current status |
|---|---|---|---|
| 10 | MCP Protocol Foundation | Workflow Phase 10 MCP subset | complete |
| 11 | MCP Manager And `/mcp` Runtime | Workflow Phase 10 MCP manager subset | complete |
| 12 | Reverse-Engineering Memory Runtime | Workflow Phase 11 memory subset | complete |
| 13 | Doctor Diagnostics | Workflow Phase 12 doctor subset | complete |
| 14 | Product Runtime Wiring | New forward phase before full hardening | complete |
| 15 | Interactive Smoke And Golden Runtime | Runtime hardening before parity closure | complete |
| 16 | Install Scripts | Workflow Phase 12 install subset | complete for local scripts |
| 17 | Release Metadata And Checksums | Workflow Phase 12 release subset | complete for local metadata/checksums |
| 18 | CI Gate Expansion | Workflow Phase 12 CI subset and Phase 13 gate prep | complete for baseline verify matrix |
| 19 | Full Parity Matrix Closure | Workflow Phase 13 | complete |
| 20 | Security Hardening Audit | Workflow Phase 13 security audit | complete |
| 21 | Documentation Coverage Hardening | Workflow Phase 13 docs coverage | complete |
| 22 | Cross-Platform Validation | Workflow Phase 13 platform validation | complete |
| 23 | Performance And Local-First Budgeting | Workflow Phase 13 performance baseline | complete |
| 24 | User Walkthrough And Migration Guide | Workflow Phase 14 release candidate prep | complete |
| 25 | Release Candidate | Workflow Phase 14 | complete for local RC prep |
| 26 | RC Soak And Bugfix Freeze | Workflow Phase 14 stabilization | complete for local soak |
| 27 | RecompHamr 2.0 Stable Release | Stable release gate after workflow Phase 14 | partially satisfied; blocked pending publication evidence |
| 28 | Live End-User Runtime Integration | Corrective runtime integration phase | next; blocks feature intake |
| 29 | Post-Parity Feature Intake | Workflow Phase 15 | blocked until Phase 28 and stable publication |
| 30 | Extension Architecture Planning | Workflow Phase 15 extension planning | blocked until Phase 29 intake |

## Global Gates

Every phase starts with `create_goal`, refreshes `AGENTS.md`,
`recomphamr_2_rewrite_workflow.md`, `docs/dev/04_Workflows/PhaseGoals.md`,
parity/status docs, this roadmap, and relevant subsystem docs. Every phase
closes with `make verify`, updated docs, traceability rows, status evidence,
and `update_goal`.

Mandatory gates:

- 100% statement coverage.
- 100% docs, docstrings, help, command, config, and examples coverage.
- No placeholders, fake success paths, speculative APIs, or silent unsupported
  branches.
- Zero hallucination: unsupported claims are labeled `unverified`,
  `unsupported`, or `blocked`.
- No scope creep before parity release.

## Remaining Forward Phases

### Phase 14 — Product Runtime Wiring

Replace diagnostic-only default runtime with the integrated TUI/app path:
config load, project bootstrap detection, LLM client, agent loop, tools, skills,
commands, MCP manager, and cancellation ownership.

### Phase 15 — Interactive Smoke And Golden Runtime

Add end-to-end fake-runtime tests for startup, first prompt, slash command
dispatch, tool call loop, cancellation, memory injection, and transcript
rendering. Keep network tests fake by default.

### Phase 19 — Full Parity Matrix Closure

Audit every parity row against RecompHamr 1.x evidence. Each row must be
passed, removed by explicit decision, or marked `blocked:` with evidence. No new
feature planning is allowed in this phase.

### Phase 20 — Security Hardening Audit

Run a focused security pass over filesystem boundaries, symlinks, permissions,
command execution, MCP process spawning, network fetches, secret redaction, and
logs. Add regression tests for every fixed issue.

### Phase 21 — Documentation Coverage Hardening

Upgrade docs coverage checks so commands, tools, config keys, environment
variables, MCP settings, generated files, exported symbols, and help text are
all mechanically verified.

### Phase 22 — Cross-Platform Validation

Validate Windows behavior first, then Linux/macOS where available. Cover paths,
permissions, process cancellation, install scripts, TUI rendering, and release
artifacts. Phase 22 closes with `docs/dev/06_Testing/PlatformMatrix.md` as the
durable evidence ledger and records unsupported limits for installer execution,
remote release verification, live Bubble Tea process launch, and OS-specific
process-group termination guarantees.

### Phase 23 — Performance And Local-First Budgeting

Measure startup, context packing, tool execution overhead, MCP listing, and TUI
render cost. Add documented baselines and tests that prevent obvious
regressions. Phase 23 closes with deterministic package benchmarks and
`docs/dev/06_Testing/PerformanceBenchmarks.md` recording the local
`windows/amd64` baseline.

### Phase 24 — User Walkthrough And Migration Guide

Create a fresh-clone walkthrough covering install, config, model profile setup,
`/init-re`, skills, tools, MCP, memory, doctor, and troubleshooting. Add
migration notes from RecompHamr 1.x. Phase 24 closes with
`docs/user/walkthrough.md` and `docs/user/migration.md`, both limited to
verified behavior and explicit `unsupported` runtime/release boundaries.

### Phase 25 — Release Candidate

Produce the release candidate branch/tag artifacts, release notes, checksums,
packaged docs, and final known-limits document. Only parity fixes and
documentation corrections are allowed after this point. In this checkout, Phase
25 closes as local RC preparation: release notes, checksum guidance,
packaged-docs list, and known-limits docs are complete, while published tags,
uploaded artifacts, and remote downloads remain `unsupported`.

### Phase 26 — RC Soak And Bugfix Freeze

Run the full verification matrix repeatedly, fix only verified release blockers,
and update status reports after each fix. No feature additions. In this
checkout, Phase 26 closes as local soak: freeze rules and blocker policy are
documented in `RCSoak.md`, and stable release remains gated on Phase 27.

### Phase 27 — RecompHamr 2.0 Stable Release

Cut the stable release when all parity, docs, coverage, security, install, and
smoke gates pass. Publish artifacts only after checksum verification and
fresh-install validation. In this checkout, Phase 27 records local stable-gate
readiness, generated artifacts for six targets, verified `dist/SHA256SUMS`, and
Windows installer smoke evidence in `StableRelease.md`. The local stable tag
decision is `v2.0.0`; publication remains `blocked:` because no external
CI/platform evidence, uploaded artifact evidence, or publication destination
evidence exists locally.

### Phase 28 — Live End-User Runtime Integration

Correct the runtime integration gap by making `recomphamr` launch the usable
terminal app instead of only printing a startup summary. Wire the Bubble Tea
program loop, prompt submission, slash command dispatch, real LLM client,
agent loop, built-in tool dispatcher, cancellation, status updates, debug
redaction, and graceful quit path into the CLI while preserving strict package
boundaries. This is parity repair, not new feature intake.

Phase 28 closes only when a user can run the CLI, see the TUI, run slash
commands, submit a prompt through the agent loop, cancel work, and exit cleanly
with tests, docs, help text, security notes, and 100% statement coverage.

### Phase 29 — Post-Parity Feature Intake

Open feature planning after live runtime integration and stable publication.
Create a decision register for candidate enhancements such as safer permission
prompts, session export/import, richer reverse-engineering dashboards,
ACP/editor integration, and optional desktop shell.

### Phase 30 — Extension Architecture Planning

Design post-parity extension boundaries for optional Rust helpers, external
analyzers, plugin-style tools, richer MCP integrations, and future UI surfaces.
No implementation begins until each feature has its own approved goal packet.

## Public Interfaces And Types

- MCP protocol/client/manager interfaces live under `internal/mcp` and must
  document request, response, transport, server, tool, and lifecycle types.
- Runtime composition options belong under `internal/app` so tests can inject
  fake LLMs, fake tools, fake MCP managers, and fake filesystem paths.
- Doctor report types must keep stable section, status, and error fields for
  `/doctor`, docs, and tests.
- Release/version metadata exposed by the CLI must be documented in install and
  release docs.

## Test Plan

- Every phase must keep `make verify` green with 100% statement coverage.
- MCP phases require fake server integration tests and JSON-RPC golden tests.
- Runtime phases require fake LLM/tool/MCP end-to-end tests without real network
  dependencies.
- Install/release phases require dry-run tests, checksum mismatch tests, and
  platform-script checks.
- Final parity phases require docs coverage, command/help coverage, config-doc
  coverage, race smoke, fuzz smoke, and security regression tests.

## Assumptions

- "Next 20 phases" now means the reconciled forward plan from phases 10
  through 30 after completed Phase 9; Phase 28 was inserted as a corrective
  runtime integration phase, moving the previous phases 28 and 29 to phases 29
  and 30.
- Go remains the implementation language and `make verify` remains canonical.
- Windows remains the primary target, with Linux/macOS validated where scripts
  and CI are available.
- No post-parity feature work begins before Phase 28 live runtime integration
  and stable publication evidence both pass.

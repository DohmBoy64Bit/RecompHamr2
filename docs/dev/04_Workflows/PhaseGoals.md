# Phase Goal Packets

Every phase starts by copying or refreshing the matching packet. Update evidence and verification commands with exact local facts before closing a phase.

For post-Phase-9 work, reconcile phase numbers against
`docs/dev/07_ProjectManagement/Next20PhasePlan.md`. If the original workflow
phase number and the Next 20 phase number differ, record both the phase title
and mapping in the goal packet.

## Phase 0: Source Inventory And Parity Freeze

Outcome: definitive RecompHamr 1.x behavior inventory and parity matrix.
Scope: commands, tools, MCP, skills, config, memory, LLM streaming, context packing, diagnostics, install, update, release, docs, and golden outputs.
Out of scope: new RecompHamr 2.0 features.
Evidence required: RecompHamr 1.x source location, inventory output, golden captures, parity rows.
Verification commands: inventory script once created; `make verify`.
Stop condition: blocked if RecompHamr 1.x source is unavailable.

## Phase 1: Governance, AGENTS.md, And Docs Memory

Outcome: mandatory rules are durable outside the workflow document.
Scope: `AGENTS.md`, root docs, user docs, developer docs, docs index, docscheck, memory refresh, zero-hallucination, no-placeholder, coverage, security, and final report policy.
Out of scope: product runtime behavior.
Evidence required: docs tree, `.docs-index.md`, docscheck output, traceability rows.
Verification commands: `make docscheck`; `make verify`.
Stop condition: blocked if any mandatory rule is missing from durable docs.

## Phase 2: Architecture Skeleton

Outcome: Go package boundaries and diagnostic-only executable exist without fake functionality.
Scope: Go module, `Makefile`, `cmd/recomphamr`, `internal/` packages, tests, package docs, diagnostic mode, redaction foundation.
Out of scope: config, LLM, tools, TUI, slash commands, skills, MCP, and product runtime.
Evidence required: package docs, tests, diagnostic output, unsupported default runtime.
Verification commands: `make verify`; `go test -cover ./...`; `go run ./cmd/recomphamr --diagnostic`.
Stop condition: blocked if any package has fake success behavior or uncovered implemented statements.

## Phase 3: Config And Workspace Parity

Outcome: secure config and `.rehamr/` workspace behavior matches parity evidence.
Scope: strict YAML, default profiles, env overrides, atomic writes, permissions, symlink refusal, traversal checks, config docs and examples.
Out of scope: LLM network behavior and tools not required for config verification.
Evidence required: parity captures, config examples, security tests, docs coverage.
Verification commands: `make verify`; config parity tests.
Stop condition: blocked if reference config behavior is unavailable.

## Phase 4: LLM Client And Context Packing

Outcome: OpenAI-compatible streaming and context budgeting are implemented with evidence.
Scope: request/response types, SSE parser, streaming assembly, provider errors, idle timeout, cancellation, context packing, truncation, and invalid tool-call protection.
Out of scope: TUI presentation and real tool execution.
Evidence required: golden streaming tests, fake provider tests, context packing tests, provider docs.
Verification commands: `make verify`; LLM golden tests.
Stop condition: blocked if provider behavior cannot be reproduced or documented.

## Phase 5: Built-In Tools

Outcome: all six built-in tools are implemented with parity, schemas, docs, and security tests.
Scope: `bash`, `read_file`, `write_file`, `edit_file`, `repomixr`, `recomp_reference`, schemas, validation, cancellation, redaction, path rules, examples.
Out of scope: MCP tools and slash command UI.
Evidence required: reference tool schemas, golden outputs, security tests, parameter docs.
Verification commands: `make verify`; tool schema and output golden tests.
Stop condition: blocked if a tool exposes undocumented or untested filesystem, network, or shell behavior.

## Phase 6: Agent Loop

Outcome: deterministic model-tool turn loop works independently from the TUI.
Scope: turn state machine, tool dispatch, max rounds, failure nudges, verification nudges, transcript state, cancellation ownership, fake LLM tests.
Out of scope: Bubble Tea rendering and command palette.
Evidence required: fake LLM scripts, loop tests, cancellation tests, docs.
Verification commands: `make verify`; agent loop tests.
Stop condition: blocked if any loop can continue indefinitely or claim success without evidence.

## Phase 7: TUI Shell

Outcome: terminal interface renders and dispatches state without owning core logic.
Scope: Bubble Tea model, transcript, prompt editor, status footer, slash palette, completions, resize, paste, history, cancellation, quit, debug log redaction.
Out of scope: changing agent, tool, LLM, MCP, or config semantics.
Evidence required: golden renders, key handling tests, resize tests, docs and help.
Verification commands: `make verify`; TUI golden tests.
Stop condition: blocked if UI logic duplicates core behavior or cannot be tested without a terminal.

## Phase 8: Slash Commands

Outcome: all 11 parity slash commands are implemented with generated help and docs.
Scope: registry, `/clear`, `/models`, `/skills`, `/skill`, `/skill-audit`, `/skill-new`, `/init-re`, `/status-re`, `/doctor`, `/mcp`, `/help`, docs generation.
Out of scope: new commands outside parity.
Evidence required: command parity rows, help output, examples, parsing tests, error tests.
Verification commands: `make verify`; command golden tests.
Stop condition: blocked if any command lacks docs, help, tests, or parity evidence.

## Phase 9: Skills

Outcome: embedded and custom skill behavior matches parity requirements.
Scope: 28 embedded skills, custom discovery, precedence, exact/case-insensitive/`.md` resolution, active state, listing, audit classification, and skill-new workflow.
Out of scope: MCP server implementation beyond skill-gating contracts.
Evidence required: embedded content golden tests, resolver tests, custom precedence tests, audit tests, network-fetch docs.
Verification commands: `make verify`; skill golden tests.
Stop condition: blocked if any skill behavior is undocumented, untested, or network-fetching without explicit user-visible reason.

## Phase 10: MCP Protocol Foundation

Outcome: JSON-RPC 2.0 MCP protocol and transport foundations are implemented without taking over server manager lifecycle.
Scope: request, response, notification, initialize, tools/list, tools/call, tool schema, stdio transport over injected streams, streamable HTTP request/response transport, context cancellation, HTTP errors, RPC errors, docs, and tests.
Out of scope: process spawning, persistent manager state, `/mcp` lifecycle mutation, tool execution through the agent, and real external MCP server dependencies.
Evidence required: RecompHamr 1.x MCP source evidence, fake transport tests, fake HTTP server tests, protocol docs, user docs, parity rows, and traceability rows.
Verification commands: `go test ./internal/mcp -cover`; `make verify`.
Stop condition: blocked if protocol behavior requires an undocumented external server or if any transport path cannot be tested without a real MCP process.

## Plan Phase 11: MCP Manager And `/mcp` Runtime

Outcome: MCP lifecycle state, tool visibility, and `/mcp` runtime commands are implemented over explicit manager wiring.
Scope: manager registration, connect, disconnect, status, enabled tool allowlists, skill-gated visibility, `server.tool` calls, HTTP connector initialization, protocol-backed client sessions, and `/mcp connect|disconnect|tools|enable|disable` command behavior.
Out of scope: stdio process spawning, app startup autoconnect, agent-loop MCP tool exposure, persistent user MCP config files, and external server dependency tests.
Evidence required: RecompHamr 1.x manager evidence, fake connector tests, fake HTTP connector tests, command dispatch tests, user docs, developer docs, parity rows, and traceability rows.
Verification commands: `go test ./internal/mcp -cover`; `go test ./internal/commands -cover`; `make verify`.
Stop condition: blocked if lifecycle behavior can only be verified against a real external MCP server or if unsupported stdio process spawning would have to pretend success.

## Memory Runtime Phase: RE Workspace Prompt Context

Outcome: `.rehamr/REPHAMR_STATE.md` can be loaded and injected into model context
without weakening package boundaries.
Scope: workspace memory loader, explicit missing-memory errors, UTF-8-safe read
cap, prompt injection helper, token-budget trimming, agent-loop memory fields,
docs, parity rows, traceability, and status report updates.
Out of scope: executable startup auto-load, TUI runtime wiring, MCP memory sync,
and hidden background mutation of workspace memory.
Evidence required: project, LLM, and agent tests; user memory docs; architecture
contracts; parity and traceability rows; before/after doc hashes.
Verification commands: `go test ./internal/project -cover`; `go test
./internal/llm -cover`; `go test ./internal/agent -cover`; `make verify`.
Stop condition: blocked if runtime memory would require filesystem access from
`internal/llm` or `internal/agent`, or if any behavior change lacks matching
docs and tests.

## Phase 12 Slice: Doctor Local Diagnostics

Outcome: `/doctor` reports real offline diagnostics without pretending install,
update, release, or checksum parity.
Scope: `internal/doctor`, `/doctor` command dispatch, runtime/workspace/config/
memory/skills/tools/MCP registration checks, status labels, docs, parity rows,
traceability rows, and architecture boundary updates.
Out of scope: installer scripts, self-update, release asset checks, checksum
verification, devcontainer checks, CI workflow diagnostics, network probes, and
external process launches.
Evidence required: doctor tests, command dispatch tests, architecture docs, user
doctor docs, `DoctorParity.md`, status report, doc hash comparison, and full
verification output.
Verification commands: `go test ./internal/doctor -cover`; `go test
./internal/commands -cover`; `go test ./internal/archcheck -cover`; `make
verify`.
Stop condition: blocked if a diagnostic would require mutation, network access,
external tools, or an undocumented release/install assumption.

## Phase 12 Slice: Release Artifact Names And Checksum Verification

Outcome: release artifact names are deterministic, supported targets can be
built locally with `go build`, already-built binaries can be archived locally,
local SHA-256 manifests can be generated deterministically, and local release
artifacts can be verified against a SHA-256 manifest without implementing
downloads, installers, or self-update.
Scope: `internal/release`, supported target metadata, artifact naming,
local build orchestration, archive creation from already-built binaries,
SHA256SUMS generation and parsing, local artifact hashing, relative path safety,
mismatch reporting, user install docs, release workflow docs, release parity,
traceability, security notes, and architecture boundaries.
Out of scope: network downloads, remote checksum fetching, installer execution,
self-update, GoReleaser config, devcontainer generation, CI release dry-run,
and command-line release UI.
Evidence required: release package tests, 100% statement coverage, docs hash
comparison, parity rows, traceability rows, and full verification output.
Verification commands: `go test ./internal/release -cover`; `go test
./internal/archcheck -cover`; `make verify`.
Stop condition: blocked if verification would require network access, artifact
execution, mutation, or unverified release metadata.

## Phase 12 Completion: Doctor, Install, Update, Release, Devcontainer, And CI

Outcome: Phase 12 operational parity is complete for local, testable behavior:
doctor diagnostics, local installers, release config, self-update dry-run,
checksum generation and verification, devcontainer config, and CI workflow.
Scope: installer scripts, GoReleaser config, devcontainer config, CI workflow,
operational file validation, local self-update dry-run planning, doctor
operational diagnostics, user docs, developer docs, parity rows, traceability,
security notes, and architecture boundaries.
Out of scope: remote release downloads, remote checksum fetching, automatic
replacement of the running executable, installer execution on every platform,
release notes, migration guide, and dependency audit.
Evidence required: release, update, doctor, and archcheck tests; 100% statement
coverage; docs hash comparison; parity rows; traceability rows; full
verification output.
Verification commands: `go test ./internal/release ./internal/update
./internal/doctor ./internal/archcheck -cover`; `make verify`; `go run
./cmd/recomphamr --diagnostic`.
Stop condition: blocked if completing any item would require network access,
remote metadata, artifact execution, or pretending a platform installer ran.

## Next 20 Phase 14: Product Runtime Wiring

Outcome: the default executable path composes real local runtime state instead
of returning diagnostic-only unsupported output.
Scope: `internal/app` runtime composition, config bootstrap/load, project memory
detection, command environment, TUI shell state, MCP manager wiring without
autoconnect, startup output, help text, docs, parity rows, traceability, status,
and architecture contracts.
Out of scope: live model calls, interactive Bubble Tea process loop, external
MCP process spawning, network probes, automatic workspace initialization beyond
config bootstrap, and end-to-end fake prompt smoke tests reserved for Next 20
Phase 15.
Evidence required: app and command-line tests, 100% statement coverage,
diagnostic command evidence, docs hash comparison, parity and traceability rows,
and full verification output.
Verification commands: `go test ./internal/app ./cmd/recomphamr -cover`; `make
verify`; `go run ./cmd/recomphamr --diagnostic`; `go run ./cmd/recomphamr`.
Stop condition: blocked if runtime wiring would need network access, a real
model backend, terminal control, MCP autoconnect, or a fake successful agent
turn.

## Next 20 Phase 15: Interactive Smoke And Golden Runtime

Outcome: deterministic fake-runtime smoke tests cover startup-adjacent
interactive behavior without network, real model, real tool, or terminal
process dependencies.
Scope: `internal/app.RunSmoke`, slash-command dispatch through composed TUI
state, first prompt execution through injected fake model and tool runner,
cancellation before agent execution, project memory injection, rendered
transcript evidence, docs, parity rows, traceability, and status evidence.
Out of scope: live backend model calls, real tool execution, MCP autoconnect,
external network tests, and launching the interactive Bubble Tea process.
Evidence required: app smoke tests, architecture boundary docs, user runtime
docs, parity and traceability rows, status report, docs hash comparison, and
full verification output.
Verification commands: `go test ./internal/app ./cmd/recomphamr
./internal/archcheck -cover`; `make verify`; `go run ./cmd/recomphamr
--diagnostic`.
Stop condition: blocked if smoke behavior requires real network/model/tool
dependencies, terminal control, or an undocumented app-to-subsystem import.

## Next 20 Phase 19: Full Parity Matrix Closure

Outcome: every current parity matrix row is audited against local evidence and
marked passed, partial, or blocked without adding post-parity features.
Scope: top-level parity matrix, detailed feature parity docs, Phase 19 closure
ledger, traceability rows, status report, docs index, stale wording cleanup,
documentation hashes, and full verification evidence.
Out of scope: implementing remaining unsupported real-backend, live terminal,
stdio process spawning, remote release, or platform execution behavior.
Evidence required: `ParityClosure.md`, updated parity rows, traceability and
status rows, stale wording scan, placeholder-policy scan, docs hash comparison,
and full verification output.
Verification commands: `make verify`; `go run ./cmd/recomphamr --diagnostic`;
stale parity wording scan; placeholder-policy scan.
Stop condition: blocked if a row cannot be tied to source, docs, tests, or an
explicit verified unsupported limit.

## Next 20 Phase 20: Security Hardening Audit

Outcome: implemented security boundaries are audited, one direct project-memory
symlink gap is closed, and remaining security-sensitive unsupported behavior is
documented with evidence.
Scope: filesystem symlink checks, workspace memory/status reads, command/tool
execution boundaries, MCP execution boundaries, release verification
boundaries, redaction, product startup limits, user security docs, developer
security docs, traceability, status, and verification evidence.
Out of scope: implementing real backend prompt turns, real product tool
execution, MCP autoconnect, stdio process spawning, persistent user MCP config
files, remote release downloads, binary replacement, or platform installer
execution tests.
Evidence required: project symlink regression tests, `SecurityAudit.md`,
updated security docs, traceability and status rows, security keyword scan,
placeholder-policy scan, docs hash comparison, and full verification output.
Verification commands: `go test ./internal/project -cover`; `make verify`;
`go run ./cmd/recomphamr --diagnostic`; security keyword scan.
Stop condition: blocked if a security claim cannot be tied to code, tests,
docs, or an explicit verified unsupported limit.

## Next 20 Phase 21: Documentation Coverage Hardening

Outcome: documentation coverage checks mechanically verify durable docs,
user-visible command/tool/config/MCP/workspace/release/help terms, and exported
Go doc comments.
Scope: `internal/docscheck`, docscheck tests, documentation coverage docs,
coverage requirements, user docs terms exposed by the new checker, parity,
traceability, status, docs index, and verification evidence.
Out of scope: implementing docs site generation, public API extraction beyond
current Go exported-symbol checks, or post-parity feature documentation.
Evidence required: docscheck tests at 100% coverage, updated docs coverage
policy, traceability and status rows, docs hash comparison, placeholder-policy
scan, and full verification output.
Verification commands: `go test ./internal/docscheck -cover`; `make verify`;
`go run ./cmd/recomphamr --diagnostic`; placeholder-policy scan.
Stop condition: blocked if any implemented command, tool, config key, MCP
setting, generated file, release file, help flag, or exported Go symbol cannot
be documented and mechanically verified.

## Next 20 Phase 22: Cross-Platform Validation

Outcome: Windows-first behavior and available Linux/macOS behavior are audited
against existing source, tests, and docs without adding unsupported runtime
claims.
Scope: platform matrix, shell selection, path and permission rules, process
cancellation evidence, install script validation, TUI render portability,
release artifact targets, parity rows, traceability, status, docs hashes, and
full verification evidence.
Out of scope: executing installers on every platform, remote release download
tests, OS-specific process-group termination guarantees, launching the live
Bubble Tea process, or claiming CI results that did not run locally.
Evidence required: updated `PlatformMatrix.md`, source/test references for each
platform area, traceability and status rows, docs hash comparison,
placeholder-policy scan, and full verification output.
Verification commands: `go test ./internal/release ./internal/tools
./internal/tui ./internal/config ./internal/project ./internal/agent
./internal/mcp ./internal/app -cover`; `make verify`; `go run
./cmd/recomphamr --diagnostic`; `go env GOOS GOARCH`; placeholder-policy scan.
Stop condition: blocked if a platform claim cannot be tied to code, docs, tests,
or an explicit `unsupported` limit.

## Next 20 Phase 23: Performance And Local-First Budgeting

Outcome: startup composition, context packing, tool/runtime overhead, MCP
listing, and TUI render costs have repeatable local benchmark coverage and a
documented baseline.
Scope: benchmark functions in existing owner packages, `PerformanceBenchmarks.md`,
docs index, parity, traceability, status, benchmark command output, docs hashes,
and full verification evidence.
Out of scope: live model benchmarks, network benchmarks, installer benchmarks,
terminal process benchmarks, universal performance claims, or CI-published
release numbers.
Evidence required: local benchmark output with OS/architecture/CPU, benchmark
source in owner packages, updated performance docs, traceability and status
rows, docs hash comparison, placeholder-policy scan, and full verification
output.
Verification commands: `go test ./internal/llm ./internal/tui ./internal/mcp
./internal/tools ./internal/app -bench
"Benchmark(PackLargeHistory|RenderWideLayout|RenderCompactLayout|ManagerToolsForSkills|Schemas|ReadFileSmall|ComposeRuntimeStartup)$"
-benchmem -run "^$"`; `go test ./internal/llm ./internal/tui ./internal/mcp
./internal/tools ./internal/app -cover`; `make verify`; `go run
./cmd/recomphamr --diagnostic`.
Stop condition: blocked if any benchmark requires network, a real model, a live
terminal, or unsupported release infrastructure.

## Next 20 Phase 24: User Walkthrough And Migration Guide

Outcome: users have a verified fresh-clone walkthrough and RecompHamr 1.x
migration guide that cover implemented behavior and explicit unsupported limits.
Scope: `docs/user/walkthrough.md`, `docs/user/migration.md`, README and docs
index links, config/model setup, `/init-re`, memory, skills, tools, MCP,
doctor, troubleshooting, migration checklist, parity, traceability, status,
docs hashes, and full verification evidence.
Out of scope: release-candidate notes, published artifacts, remote downloads,
installer execution claims, live backend walkthroughs, and post-parity feature
guidance.
Evidence required: user docs with commands and examples, links from README and
docs index, traceability and status rows, docs hash comparison,
placeholder-policy scan, diagnostic command output, and full verification
output.
Verification commands: `make verify`; `go run ./cmd/recomphamr --diagnostic`;
placeholder-policy scan.
Stop condition: blocked if a walkthrough step cannot be tied to current docs,
source, tests, runtime output, or an explicit `unsupported` limit.

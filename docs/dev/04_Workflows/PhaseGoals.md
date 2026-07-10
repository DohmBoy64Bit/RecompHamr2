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

## Next 20 Phase 25: Release Candidate

Outcome: release-candidate preparation docs, release notes, checksum guidance,
packaged-docs list, and verified known limits are current without pretending a
published RC exists.
Scope: `ReleaseCandidate.md`, `KnownLimits.md`, user RC notes, changelog,
release workflow, release criteria, docs index, parity, traceability, status,
docs hashes, diagnostic evidence, and full verification evidence.
Out of scope: published tags, uploaded artifacts, remote release downloads,
remote checksum fetching, automatic executable replacement, installer execution
claims on every platform, and dependency audit.
Evidence required: RC preparation docs, known-limits ledger, changelog entry,
checksum guidance, packaged-docs list, traceability and status rows, docs hash
comparison, placeholder-policy scan, diagnostic output, and `make verify`.
Verification commands: `make verify`; `go run ./cmd/recomphamr --diagnostic`;
`go test ./internal/release ./internal/update ./internal/doctor -cover`;
placeholder-policy scan.
Stop condition: blocked if any RC claim requires a tag, upload, remote network
metadata, installer execution, or unsupported runtime behavior.

## Next 20 Phase 26: RC Soak And Bugfix Freeze

Outcome: RC soak and bugfix freeze rules are documented, local repeated
verification evidence is recorded, and the roadmap advances without adding
features.
Scope: `RCSoak.md`, docs index, phase goals, roadmap, parity, traceability,
status, known limits, diagnostic evidence, repeated verification commands,
placeholder-policy scan, and docs hash comparison.
Out of scope: feature work, published artifacts, remote downloads, automatic
binary replacement, installer execution claims on every platform, and stable
release publication.
Evidence required: soak/freeze ledger, repeated `make verify` evidence,
diagnostic output, targeted release/update/doctor coverage, placeholder-policy
scan, traceability and status rows, and docs hash comparison.
Verification commands: `make verify` twice; `go run ./cmd/recomphamr
--diagnostic`; `go test ./internal/release ./internal/update ./internal/doctor
-cover`; placeholder-policy scan.
Stop condition: blocked if any verification failure cannot be reproduced,
classified, and tied to a release blocker without adding feature scope.

## Next 20 Phase 27: RecompHamr 2.0 Stable Release

Outcome: stable-release readiness is documented with verification evidence,
local artifacts/checksums/Windows installer smoke are recorded, a stable tag is
published when intentionally cut, and external CI/platform, upload, checksum,
artifact, and publication timestamp evidence are recorded when available.
Scope: `StableRelease.md`, docs index, phase goals, roadmap, parity,
traceability, status, release criteria, known limits, diagnostic evidence,
release/update/doctor coverage, local release archives, generated
`SHA256SUMS`, Windows installer smoke, stable tag decision, publication
evidence, placeholder-policy scan, and docs hash comparison.
Out of scope: remote checksum fetching inside the app, automatic binary
replacement, installer execution claims on every platform, and opening
post-parity feature intake without Phase 35 intake approval.
Evidence required: local stable-gate docs, `make verify`, diagnostic output,
release/update/doctor coverage, local artifact list, checksum verification,
Windows installer smoke, tag/status evidence, publication evidence when
available, placeholder-policy scan, traceability and status rows, and docs hash
comparison.
Verification commands: `make verify`; `go run ./cmd/recomphamr --diagnostic`;
`go test ./internal/release ./internal/update ./internal/doctor -cover`; `git
tag --list`; placeholder-policy scan.
Stop condition: blocked if stable publication requires credentials, remote
configuration, hosted artifacts, or CI evidence that cannot be produced or
verified.

## Next 20 Phase 28: Live End-User Runtime Integration

Outcome: running `recomphamr` launches a usable Bubble Tea terminal app that
connects prompt input, slash commands, the agent loop, the OpenAI-compatible LLM
client, built-in tool dispatch, cancellation, status rendering, and clean quit
behavior without violating package boundaries.
Scope: `cmd/recomphamr`, `internal/app`, `internal/tui`, `internal/agent`,
`internal/llm`, `internal/tools`, `internal/commands`, user runtime docs,
TUI/runtime architecture docs, parity/status/traceability rows, help text,
security notes, golden render tests, command-flow tests, fake backend tests,
and live-program launch tests where locally possible.
Out of scope: new post-parity features, desktop UI, MCP autoconnect, stdio MCP
process spawning, remote release publishing, external CI claims, or unbounded
tool execution without cancellation and documented user-visible status.
Evidence required: live TUI launch evidence, slash command execution from the
live runtime, prompt-to-agent-loop evidence with fake and configurable real
provider paths, tool dispatch tests, cancellation tests, docs hash comparison,
100% statement coverage, docs/help/docstring coverage, and explicit security
boundaries for real local tool execution.
Verification commands: `go test ./cmd/recomphamr ./internal/app
./internal/tui ./internal/agent ./internal/llm ./internal/tools
./internal/commands -cover`; `make verify`; `go run ./cmd/recomphamr --help`;
manual or scripted TUI smoke with documented terminal constraints.
Stop condition: blocked if live runtime requires undocumented network access,
unbounded local command execution, fake success paths, package-boundary
violations, or terminal behavior that cannot be verified locally.

## Next 20 Phase 29: Live MCP Agent Integration

Outcome: connected, enabled, skill-visible MCP tools are exposed to live agent
turns and dispatched through the MCP manager without bypassing skill gates,
allowlists, cancellation, or package boundaries.
Scope: `internal/app` MCP tool schema merging, `internal/app` MCP tool-call
dispatch through `internal/mcp.Manager`, explicit URL-backed autostart policy,
MCP user docs, MCP architecture docs, security notes, parity/status/traceability
rows, and app tests.
Out of scope: stdio MCP process spawning, persistent user MCP config files,
external MCP server dependency tests, post-parity features, and stable
publication.
Evidence required: live prompt MCP schema exposure tests, MCP tool-call dispatch
tests, MCP tool error tests, active-skill gating evidence, autostart policy
tests, docs hash comparison, 100% statement coverage, and `make verify`.
Verification commands: `go test ./internal/app ./internal/mcp ./internal/commands
-cover`; `make verify`; `go run ./cmd/recomphamr --summary`.
Stop condition: blocked if MCP exposure would require an undocumented external
server, silent tool enablement, fake connection success, package-boundary
violation, or untested network/tool behavior.

## Corrective MCP And Publication Evidence Hardening

Outcome: stdio MCP process spawning, persistent user MCP config files, and
stable publication evidence validation are implemented where locally possible
without claiming external publication.
Scope: `internal/mcp` stdio connector and config merge behavior,
`internal/app` persistent MCP config loading and autostart, `internal/release`
publication evidence validation, MCP user docs, MCP architecture docs,
release/known-limit docs, parity/status/traceability rows, security notes, and
coverage tests.
Out of scope: uploading artifacts, creating remote CI evidence, inventing a
publication destination, broad post-parity feature intake, or bypassing MCP
skill gates and allowlists.
Evidence required: helper-process stdio tests, strict `.rehamr/mcp.json` tests,
runtime config-loading tests, publication evidence validation tests, git remote
inspection, docs hash comparison, 100% statement coverage, and `make verify`.
Verification commands: `go test ./internal/mcp ./internal/app
./internal/release -cover`; `make verify`; `git remote -v`.
Stop condition: blocked if external publication evidence requires credentials,
remote repository configuration, or hosted artifacts that are not available in
the local checkout.

## Next 20 Phase 30: TUI Reference And Parity Specification

Outcome: the TUI polish work has a non-copying, evidence-backed design and
parity specification before rendering code changes begin.
Scope: current TUI audit, RecompHamr 1.x TUI parity requirements,
user-provided reference screenshots, OpenCode public UI concepts, screen-state
inventory, visual tokens, responsive breakpoints, command palette behavior,
footer/status metrics, startup/welcome state, golden render acceptance
criteria, docs, traceability, and status reports.
Out of scope: copying OpenCode or RecompHamr 1.x source/design 1:1, changing
runtime semantics, adding new slash commands, adding desktop UI, or claiming
unverified token/cost data.
Evidence required: source/docs references, screenshot-derived observations,
non-copying design rationale, parity checklist, test plan, docs hash
comparison, and `make verify`.
Verification commands: `make verify`; stale wording scan for post-parity phase
numbers; docs coverage check.
Stop condition: blocked if a UI requirement cannot be tied to current parity,
user reference input, local source/docs, or explicitly labeled inspiration.

## Next 20 Phase 31: TUI Visual System And Responsive Layout

Outcome: RecompHamr has a polished terminal visual system with branded startup,
dark theme, prompt panel, footer/status band, and responsive wide/compact
layouts while preserving RecompHamr identity.
Scope: `internal/tui` rendering, theme tokens, startup banner, layout
composition, MCP/skill/memory indicators, status/footer rendering, wide and
compact golden outputs, user docs, architecture docs, parity/status/traceability
rows, and 100% statement coverage.
Out of scope: changing agent/tool/LLM/MCP semantics, copying OpenCode colors or
layout exactly, adding fake metrics, or adding desktop UI.
Evidence required: golden renders for startup, wide chat, compact chat,
memory-missing state, active skill state, MCP disconnected/connected states,
responsive narrow behavior, docs hash comparison, and `make verify`.
Verification commands: `go test ./internal/tui ./internal/app -cover`; `make
verify`; screenshot or golden render review.
Stop condition: blocked if visual polish requires untestable terminal behavior,
hard-coded fake state, package-boundary violations, or copied third-party UI.

## Next 20 Phase 32: TUI Composer, Palette, And Completion UX

Outcome: prompt input, slash command palette, command completion, history,
paste chips, and keybinding hints are end-user polished and registry-driven.
Scope: composer editing, multiline prompt behavior, `/` palette opening,
Tab completion, selected-row styling, argument completion, command
descriptions, history navigation, large paste chips, cancellation/quit hints,
docs/help coverage, golden renders, and tests.
Out of scope: new command semantics, model/tool behavior changes, or hard-coded
palette entries not generated from the command registry.
Evidence required: key handling tests, command registry coverage, golden
palette renders, docs/examples, 100% statement coverage, and `make verify`.
Verification commands: `go test ./internal/tui ./internal/commands
./internal/app -cover`; `make verify`.
Stop condition: blocked if any user-visible keybinding, command entry,
argument rule, or help text is undocumented or untested.

## Next 20 Phase 33: TUI Transcript, Tool Blocks, And Runtime Feedback

Outcome: conversation, command, tool, MCP, blocked, unsupported, and streaming
states render as professional end-user transcript blocks without fake data or
private reasoning history.
Scope: transcript block rendering, markdown-safe output, PowerShell/tool
blocks, MCP tool blocks, question/prompt blocks, blocked/unsupported states,
streaming/thinking status, redacted debug output, context/token/cost rendering
only when verified, docs, parity/status/traceability rows, and golden tests.
Out of scope: storing private reasoning, inventing token/cost values, changing
agent loop policy, or adding unapproved telemetry.
Evidence required: fake-runtime tests for assistant replies, tool calls, MCP
results, tool errors, blocked states, cancellation, streaming status, redaction,
golden renders, docs hash comparison, and `make verify`.
Verification commands: `go test ./internal/tui ./internal/app ./internal/agent
-cover`; `make verify`.
Stop condition: blocked if transcript rendering would need unavailable metrics,
private reasoning storage, unredacted secrets, or untestable terminal behavior.

## Next 20 Phase 34: Windows Executable And End-User Launch Polish

Outcome: users can build or install a local `recomphamr.exe`, launch the
polished TUI through it, and verify local checksums without relying on
published artifacts.
Scope: Windows build docs, installer walkthrough, local `.exe` smoke evidence,
release artifact naming docs, checksum verification docs, optional version/about
output if needed, user quickstart/install docs, release docs, parity/status/
traceability rows, and verification evidence.
Out of scope: claiming public downloads, uploading artifacts, creating external
CI evidence, replacing the running executable, or adding a desktop app.
Evidence required: local Windows `.exe` build or archive evidence,
`SHA256SUMS` verification, installer smoke where locally possible, docs hash
comparison, 100% statement coverage for touched code, and `make verify`.
Verification commands: `make verify`; `go test ./internal/release
./internal/update ./internal/app -cover`; local build command for
`recomphamr_windows_amd64.exe`; checksum verification.
Stop condition: blocked if `.exe` claims require remote publication,
credentials, hosted artifacts, or platform behavior that cannot be verified
locally.

## Next 20 Phase 35: Post-Parity Feature Intake

Outcome: after Phase 28, Phase 29, corrective TUI hardening, local `.exe`
launch polish, and stable publication evidence pass, candidate post-parity
features are registered with evidence, risk, user value,
configuration/docs impact, and explicit approval requirements.
Scope: decision register, feature intake criteria, docs impact matrix,
security/risk notes, user-visible examples, and phase goal packets for any
approved feature candidates.
Out of scope: implementing new features before approval, reopening parity
scope, or bypassing live runtime and publication gates.
Evidence required: current stable publication evidence, Phase 28 and Phase 29 closure
evidence, intake register, traceability rows, status report, docs hashes, and
`make verify`.
Verification commands: `make verify`; docs coverage check; decision-register
review.
Stop condition: blocked if Phase 28, Phase 29, corrective TUI hardening, local
`.exe` launch polish, or stable publication evidence is missing.

## Next 20 Phase 36: Extension Architecture Planning

Outcome: approved post-parity extension boundaries are designed without
implementation, with package boundaries, protocol contracts, configuration
docs, security rules, and test strategy recorded.
Scope: optional Rust helpers, external analyzers, plugin-style tools, richer
MCP integrations, future UI surfaces, ADRs, architecture docs, config examples,
and traceability rows.
Out of scope: implementation before each extension has an approved goal packet.
Evidence required: approved Phase 35 intake item, ADR or architecture doc,
config/help/docs coverage plan, security analysis, and `make verify`.
Verification commands: `make verify`; docs coverage check; architecture review.
Stop condition: blocked if an extension lacks approved intake evidence or would
violate separation of concerns.

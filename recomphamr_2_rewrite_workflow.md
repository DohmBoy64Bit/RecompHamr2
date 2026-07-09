# RecompHamr 2.0 Rewrite Workflow

## Mission

**Tagline:** Local-first terminal coding agent for reverse engineering, decompilation, and static recompilation support.

RecompHamr 2.0 is a clean rewrite, not a line-for-line port. The mission is to preserve observable RecompHamr 1.x behavior while rebuilding the internals around strict evidence, small-context operation, security, testability, and complete documentation.

A coding agent built for local LLMs makes different decisions than one built for frontier cloud models. Context is precious; every tool call has to earn its place. RecompHamr 2.0 should pick simplicity over complexity on purpose and stay small so the context window stays the user's.

Primary target hardware is a local developer workstation with at least **16 GB VRAM** and **32 GB system RAM**, using local or OpenAI-compatible inference endpoints.

## Rewrite Principle

Do not copy RecompHamr 1.x code 1:1. Treat the old implementation as an executable behavior reference and feature inventory. Rebuild each subsystem with a new design, tests, docs, and security review.

Feature parity must come before new features. New features are blocked until every RecompHamr 1.x user-visible behavior, command, tool, profile, skill, MCP behavior, config behavior, memory behavior, install path, and diagnostic behavior has a passing parity test and complete docs.

## Immediate Stack Recommendation

Use **Go** for RecompHamr 2.0 unless the project deliberately accepts a longer rewrite for lower-level ownership benefits.

The practical recommendation is:

- **Language:** Go
- **TUI:** Bubble Tea, Bubbles, Lip Gloss, Glamour, ANSI helpers
- **CLI args/config:** standard library plus a small parser only where needed
- **Serialization:** YAML for user config, JSON for protocol payloads
- **HTTP/SSE:** Go standard `net/http` plus explicit SSE parser
- **MCP:** in-tree JSON-RPC 2.0 stdio and streamable HTTP clients
- **Testing:** Go unit tests, integration tests, golden tests, fuzz tests, race tests, platform tests
- **Docs:** Markdown-first docs with generated coverage checks
- **Build/release:** Makefile or justfile, GoReleaser, shell/PowerShell installers, container/devcontainer

The reason to stay with Go is feature-parity risk. RecompHamr 1.x is already Go, already uses Charm/Bubble Tea, already ships as a small terminal binary, and already has a broad test suite. Rust is attractive for future binary-analysis engines, but the agent harness is mostly TUI, process control, JSON, HTTP streaming, files, and docs. Go gets us to audited parity faster and leaves Rust available for optional helper tools, MCP servers, or native analysis libraries later.

## Language Decision Record

### Go advantages for this rewrite

- Fastest route to feature parity with the current RecompHamr behavior.
- Strong fit for simple local-first CLIs and TUIs.
- Existing RecompHamr concepts map directly to Go packages.
- Bubble Tea ecosystem already matches the desired opencode-like terminal interface.
- Easier for local LLMs to modify safely due to simpler language surface and shorter compile cycles.
- Single static-ish binary distribution is straightforward.
- Excellent table-driven testing style for command/tool/config parity.

### Go risks

- Less compile-time ownership enforcement than Rust.
- More runtime discipline required around path safety, cancellation, and concurrency.
- TUI complexity can become hard to reason about if the Bubble Tea model grows without strict module boundaries.

### Rust advantages

- Strong memory-safety and ownership guarantees.
- Excellent fit for future reverse-engineering helpers, binary parsers, lifters, and static-analysis engines.
- Ratatui can build fast, polished TUIs.
- Strong package ecosystem for parsers, binary formats, async IO, and CLI tooling.

### Rust risks

- Larger rewrite surface before reaching parity.
- Ratatui gives lower-level control but less built-in app architecture than Bubble Tea.
- More compile-time friction for agent-driven work.
- Higher chance of architectural overbuild before the simple local-first agent is complete.

### Decision

Build the RecompHamr 2.0 agent in **Go**. Keep the architecture protocol-oriented so future Rust helpers can be added as external tools, MCP servers, or subprocess backends without rewriting the agent shell.

## Current RecompHamr 1.x Feature Inventory

The uploaded RecompHamr repository contains:

- 65 Go files, including 27 Go test files.
- 78 Markdown files, including 46 files under `docs/`.
- 11 slash commands.
- 6 built-in LLM tools.
- 8 built-in MCP server configurations.
- 28 embedded reverse-engineering skills.
- Config bootstrapping under `.rehamr/`.
- Persistent project memory via `.rehamr/REPHAMR_STATE.md`.
- Local-first OpenAI-compatible chat streaming.
- Token/context packing logic.
- Shell command execution with cancellation and timeout behavior.
- File read/write/edit tools.
- `repomixr` and `recomp_reference` reference tools.
- Doctor diagnostics.
- Self-update and release plumbing.
- Linux/macOS/Windows install scripts.

### Slash commands to preserve

| Command | Parity requirement |
|---|---|
| `/clear` | Reset conversation state without corrupting project state. |
| `/models` | List and switch model profiles with active profile display. |
| `/skills` | List built-in and custom skills with active state. |
| `/skill <name>` | Resolve and load a skill by exact, case-insensitive, or `.md` name. |
| `/skill-audit <name>` | Classify a skill and suggest the proper template. |
| `/skill-new <url>` | Fetch, classify, and scaffold a new skill through the documented workflow. |
| `/init-re` | Create the `.rehamr/` reverse-engineering workspace safely and idempotently. |
| `/status-re` | Summarize current reverse-engineering state. |
| `/doctor` | Run environment diagnostics with clear verified/unverified sections. |
| `/mcp` | Show and manage MCP server/tool status. |
| `/help` | Print canonical TUI and command help. |

### Built-in tools to preserve

| Tool | Parity requirement |
|---|---|
| `bash` | Run commands with timeout, cancellation, output capture, and safe result formatting. |
| `read_file` | Read files with size limits and precise errors. |
| `write_file` | Write files, create parent directories, and report exact status. |
| `edit_file` | Perform exact, unique string replacement with near-miss diagnostics. |
| `repomixr` | Clone and pack repositories into `.rehamr/repos/` with documented flags and output shape. |
| `recomp_reference` | Fetch/cache reference pages for offline reading. |

### Built-in MCP servers to preserve

| Server | Parity requirement |
|---|---|
| `ghidra` | Skill-gated tools for function, xref, decompile, memory, import/export, and data-region workflows. |
| `n64-debug-mcp` | Skill-gated emulator/debug bridge. |
| `pcrecomp` | Skill-gated PC recompilation tools with default allowlist. |
| `mcp-pine` | Skill-gated RPCS3/PINE workflow support. |
| `objdiff` | Skill-gated object-diff workflow support. |
| `pcsx2` | Skill-gated PCSX2 debug bridge. |
| `bizhawk` | Skill-gated multi-system emulator debug bridge. |
| `sega2asm` | Skill-gated Genesis disassembly support. |

### Embedded skills to preserve

The 28 embedded skills are:

`bizhawk`, `build-fix-loop`, `cdb-debug`, `core-re`, `evidence-mode`, `file-format-reversing`, `function-discovery`, `gb-recomp`, `gc-decomp`, `gen-decomp`, `ghidra-mcp`, `imhex`, `mcp-pine`, `n64-debug-mcp`, `n64-decomp`, `objdiff`, `pcrecomp`, `pcsx2`, `project-handoff`, `ps2recomp`, `ps3recomp`, `recomp-foundations`, `sega2asm`, `snesrecomp`, `vb-decomp`, `windows-game-decomp`, `xbox360-decomp`, and `xboxrecomp`.

## Documentation Model

AetherX86's useful pattern is a strict, numbered governance tree with living registers. RecompHamr 2.0 should adapt that style while keeping user docs and developer docs separate.

Recommended layout:

```text
AGENTS.md
README.md
SECURITY.md
CONTRIBUTING.md
CHANGELOG.md
LICENSE

cmd/recomphamr/
internal/
pkg/                         optional public packages only when truly stable
assets/
scripts/
testdata/

.docs-index.md               generated docs coverage index

docs/
  user/
    quickstart.md
    install.md
    configuration.md
    model-profiles.md
    commands.md
    tools.md
    skills.md
    mcp.md
    memory.md
    doctor.md
    security-sandboxing.md
    troubleshooting.md
  dev/
    00_Governance/
      Constitution.md
      AntiHallucination.md
      DefinitionOfDone.md
      ScopeControl.md
      ChangeControl.md
      CodingStandard.md
      ReviewPolicy.md
      SecurityPolicy.md
    01_Memory/
      DecisionLog.md
      AssumptionRegistry.md
      ProblemRegistry.md
      UnresolvedFacts.md
      SessionHistory.md
      LessonsLearned.md
      ArchitectureMemory.md
    02_Verification/
      EvidenceTemplate.md
      ConfidencePolicy.md
      Sources.md
      RegressionRules.md
      AuditChecklist.md
      CrossReferenceRules.md
    03_Architecture/
      Overview.md
      ModuleContracts.md
      DataFlow.md
      Protocols.md
      ContextPacking.md
      ToolRuntime.md
      MCPArchitecture.md
      TUIArchitecture.md
      ErrorHandling.md
      LoggingPolicy.md
      SecurityBoundaries.md
      ADR_TEMPLATE.md
    04_Workflows/
      SessionStart.md
      GoalLifecycle.md
      TaskLifecycle.md
      DocumentationWorkflow.md
      TestWorkflow.md
      ReleaseWorkflow.md
      SessionEnd.md
    05_FeatureParity/
      ParityMatrix.md
      RecompHamr1Inventory.md
      CommandParity.md
      ToolParity.md
      MCPParity.md
      SkillParity.md
      TUIParity.md
      ConfigParity.md
      MemoryParity.md
    06_Testing/
      CoverageRequirements.md
      UnitTests.md
      IntegrationTests.md
      GoldenOutputs.md
      Fuzzing.md
      RaceTests.md
      PlatformMatrix.md
      PerformanceBenchmarks.md
    07_ProjectManagement/
      Milestones.md
      TraceabilityMatrix.md
      ReleaseCriteria.md
      StatusReports.md
    08_References/
      ExternalTools.md
      LocalLLMBackends.md
      MCPReferences.md
      ReverseEngineeringReferences.md
```

## AGENTS.md Requirements

`AGENTS.md` is mandatory and must stay short enough for every coding agent to read before work. It is the entrypoint, not the full documentation set.

It must include:

1. Mission and feature-parity freeze.
2. Mandatory startup reading order.
3. Goal system rules.
4. Documentation coverage rule.
5. Test coverage rule.
6. Zero-hallucination policy.
7. Security and sandbox rules.
8. Stop conditions.
9. Required final report format.

### Mandatory startup reading order

Before any phase or code-changing task, the agent must read:

1. `AGENTS.md`
2. `README.md`
3. `docs/dev/00_Governance/AntiHallucination.md`
4. `docs/dev/00_Governance/DefinitionOfDone.md`
5. `docs/dev/05_FeatureParity/ParityMatrix.md`
6. `docs/dev/07_ProjectManagement/TraceabilityMatrix.md`
7. The relevant subsystem docs for the files being changed
8. The current build/test result file

A task is invalid if the agent cannot name the relevant docs it read.

## Goal System

Every phase and every non-trivial task must start with a written goal packet.

### Goal packet template

```md
# Goal: <short outcome name>

## Outcome
<What will be true when this goal is complete.>

## Scope
<Files, modules, docs, and behaviors included.>

## Out of scope
<Anything explicitly blocked during this goal.>

## Evidence required
- Source evidence:
- Documentation evidence:
- Test evidence:
- Security evidence:

## Verification commands
- <command>
- <command>

## Stop condition
<The exact blocker or repeated failure condition that requires stopping and recording evidence.>
```

Goal closure requires:

- Code complete.
- Tests complete.
- User docs updated.
- Developer docs updated.
- Traceability matrix updated.
- Session history updated.
- Build/test evidence recorded.
- No placeholder files or vague future-work comments introduced.

## Zero-Hallucination Policy

RecompHamr 2.0 must never claim unsupported behavior. If a command, tool, MCP server, model profile, external dependency, or reverse-engineering claim cannot be verified, report it as `unverified` or `blocked` with the missing evidence.

Rules:

- Repository facts require local source or docs evidence.
- External facts require current official docs or authoritative source evidence.
- Runtime behavior requires a reproducible test, trace, or command result.
- Decompiler output is never source truth by itself.
- MCP tool output is evidence, not authority, unless cross-checked.
- After three failed attempts at the same blocker, stop, record exact evidence, and do not keep guessing.

## No Placeholder Policy

The repository must reject placeholder-only files, fake implementations, vague future-work comments, unimplemented stubs that pretend success, speculative APIs, silent unsupported branches, and generated docs that are not checked against actual commands.

Allowed forms of incomplete knowledge are:

- `blocked:` with evidence and stop condition.
- `unverified:` with the exact missing verification.
- `unsupported:` with a clear user-facing error and a test proving that error.

## Documentation Coverage Definition

Documentation coverage is 100% only when every user-visible and developer-visible item is linked from the docs index and covered by the appropriate user and developer pages.

Coverage includes:

- Every slash command.
- Every CLI flag.
- Every environment variable.
- Every config key.
- Every generated config file.
- Every built-in tool and tool parameter.
- Every MCP server and environment override.
- Every embedded skill and skill-loading rule.
- Every persistent file under `.rehamr/`.
- Every security boundary and permission rule.
- Every install and release path.
- Every error class users can act on.
- Every exported package, type, function, and method.

A generated docs coverage checker should fail CI if any command/tool/config/schema/exported symbol is undocumented.

## Code Documentation Standard

For Go:

- Every exported package, type, function, method, constant, and variable must have a Go doc comment beginning with the exported identifier.
- Every non-exported function longer than a small helper must have a purpose comment unless its behavior is obvious from name and tests.
- Every command handler must document accepted args, side effects, failure modes, and user-visible output.
- Every tool schema must document parameters, defaults, limits, path rules, security risks, and examples.
- Every concurrency boundary must document cancellation and ownership.
- Every filesystem boundary must document symlink, permissions, path traversal, and workspace behavior.

For Rust, if selected instead:

- Every public item must have rustdoc.
- `#![deny(missing_docs)]` must be enabled for public crates.
- Unsafe code is denied unless an ADR permits it with tests and safety comments.

## Test Coverage Standard

The user asked for 100% test coverage. Treat that as a hard CI gate.

Required test layers:

1. Unit tests for every package/module.
2. Golden tests for current RecompHamr 1.x visible outputs.
3. Integration tests for LLM streaming, tool calls, MCP lifecycle, config bootstrap, memory, TUI command dispatch, and install scripts.
4. Fuzz tests for config parsing, JSON streaming, MCP payloads, path handling, edit replacement, and context packing.
5. Race tests for streaming, cancellation, prompt history, MCP manager, and skill registry.
6. Platform tests for Linux, macOS, and Windows behavior.
7. Security regression tests for symlinks, traversal, permission tightening, command cancellation, and secret redaction.
8. Docs coverage tests.

Coverage rule:

```text
No package can merge below 100% statement coverage.
No exported symbol can merge without docs.
No user-visible behavior can merge without a parity or behavior test.
No fixed bug can close without a regression test.
```

Because platform-specific files can only execute on their platform, CI should measure a union coverage report across Linux, macOS, and Windows jobs.

## Security Requirements

Security is part of parity, not a later hardening pass.

Required controls:

- Refuse symlinked `.rehamr/` and symlinked config files.
- Owner-only permissions for sensitive config and state files.
- Atomic config writes.
- Path traversal rejection for repo-packing and cache tools.
- Workspace boundary checks for file tools, with explicit escape policy if any escape is allowed.
- Shell command cancellation that terminates child processes where the OS supports it.
- Bounded command timeout.
- Secrets redaction in logs, debug files, errors, and final reports.
- MCP tools off by default unless skill-gated or explicitly enabled.
- MCP server command and URL overrides documented and tested.
- No network fetching without a user-visible tool reason.
- No silent fallback from local model to cloud model.
- Clear warning that shell tools execute with local filesystem permissions.
- Dependency audit in CI.
- Generated release checksum verification.

## Architecture

Recommended Go package layout:

```text
cmd/recomphamr/              process entrypoint, flags, version, update check
internal/app/                app wiring and dependency construction
internal/tui/                Bubble Tea model, rendering, key handling
internal/commands/           slash command registry and handlers
internal/agent/              turn loop, tool loop, verification nudges
internal/llm/                OpenAI-compatible client, SSE, provider quirks
internal/contextpack/        token estimates, packing, truncation, anchoring
internal/tools/              built-in tool schemas and execution
internal/mcp/                MCP clients, manager, config, tool gating
internal/skills/             embedded and custom skills
internal/config/             config schema, bootstrap, persistence
internal/project/            `.rehamr/` workspace and memory files
internal/doctor/             diagnostics
internal/security/           path, permission, redaction helpers
internal/update/             self-update and release verification
internal/docscheck/          generated docs coverage checker
internal/testharness/        golden/parity helpers
```

Key rule: UI is not the agent. The TUI should render app state and dispatch user intent. Agent loop, tools, LLM, MCP, config, and docs checks must be independently testable without a terminal.

## TUI Design Direction

The interface should be opencode-like in polish while staying RecompHamr-small.

Required UI elements:

- Chat transcript with markdown rendering.
- Multiline prompt editor.
- Slash command palette.
- Argument completion popover.
- Model/status footer.
- Active skill indicators.
- MCP status indicators.
- Streaming tool-call status line.
- Reasoning/progress animation without storing private reasoning in history.
- Copy/paste handling with large-paste chips or attachments.
- Clear cancellation model for Ctrl+C.
- Clear quit model for Ctrl+D or double Ctrl+C.
- Debug log mode that redacts secrets.

Avoid a full desktop UI until feature parity is complete. A desktop interface multiplies scope and threatens the small-context mission.

## Feature-Parity Workflow

### Phase 0 — Source inventory and parity freeze

Goal: produce the definitive RecompHamr 1.x behavior inventory.

Outputs:

- `docs/dev/05_FeatureParity/RecompHamr1Inventory.md`
- `docs/dev/05_FeatureParity/ParityMatrix.md`
- Golden command-output captures.
- Golden config/bootstrap captures.
- Golden tool schema captures.
- Golden docs index.

Verification:

- Inventory script can enumerate commands, tools, skills, MCP servers, config keys, env vars, and docs files.
- Every discovered item has a parity row.
- No new feature work begins.

### Phase 1 — Governance, AGENTS.md, and docs skeleton

Goal: create the rules the rewrite will obey before writing the product code.

Outputs:

- Root `AGENTS.md`.
- User docs skeleton.
- Developer docs skeleton.
- Goal lifecycle docs.
- Definition of done.
- Anti-hallucination policy.
- Docs coverage checker initial version.

Verification:

- Docs index generated.
- Every skeleton file contains real policy or contract text, not filler.
- CI fails on undocumented commands/tools once code begins.

### Phase 2 — Architecture skeleton

Goal: build package boundaries without fake functionality.

Outputs:

- `cmd/recomphamr` entrypoint.
- Package contracts.
- Dependency wiring.
- Test harness.
- Logging/redaction foundation.

Verification:

- Empty app starts only in a documented diagnostic mode.
- No fake successful commands exist.
- All packages have docs and tests.

### Phase 3 — Config and project workspace parity

Goal: implement `.rehamr/` bootstrap and profile behavior.

Outputs:

- Config schema.
- Default local profiles.
- Strict YAML decode.
- Atomic save.
- Permissions and symlink defenses.
- Environment URL override behavior.

Verification:

- Golden parity tests from RecompHamr 1.x.
- Security tests for symlinks and loose permissions.
- Docs for every config key and env var.

### Phase 4 — LLM client and context packing

Goal: implement OpenAI-compatible streaming and local-first context budgeting.

Outputs:

- Chat request/response types.
- SSE parser.
- Tool-call streaming assembly.
- Reasoning delta handling without history persistence.
- Context window and budget header handling.
- Context packing and truncation.

Verification:

- Golden streaming tests.
- Fragmented tool-call tests.
- Idle timeout tests.
- Provider error tests.
- Context packing tests with orphan/dangling tool calls.

### Phase 5 — Built-in tools

Goal: implement the six built-in tools with security and parity tests.

Outputs:

- `bash`
- `read_file`
- `write_file`
- `edit_file`
- `repomixr`
- `recomp_reference`

Verification:

- Schema golden tests.
- Tool output golden tests.
- Security tests.
- Cancellation tests.
- Docs for every parameter and failure mode.

### Phase 6 — Agent loop

Goal: implement the model-tool loop independently from the TUI.

Outputs:

- Turn state machine.
- Tool-call dispatch.
- Max tool round policy.
- Failure nudge policy.
- Verification nudge policy.
- Cancellation ownership.

Verification:

- Scripted fake LLM tests.
- Tool loop tests.
- Cancellation tests.
- No infinite tool loop cases.

### Phase 7 — TUI shell

Goal: implement the polished terminal interface.

Outputs:

- Bubble Tea model.
- Prompt editor.
- Transcript renderer.
- Footer/status bar.
- Command palette.
- Completion popovers.
- Paste handling.
- Resize handling.
- History persistence.

Verification:

- TUI model tests.
- Golden render tests.
- Key handling tests.
- Prompt history tests.
- Cross-platform terminal behavior tests.

### Phase 8 — Slash commands

Goal: implement all 11 commands with parity.

Outputs:

- Command registry.
- Command help generation.
- Command docs generation.
- Command tests.

Verification:

- Every command has unit tests, docs, help text, and parity rows.
- `/help` output is generated from the command registry.
- User docs include examples and errors.

### Phase 9 — Skills

Goal: implement embedded and custom skill behavior.

Outputs:

- Embedded skill bundle.
- Skill resolver.
- Custom skill directory.
- Skill list rendering.
- Skill audit classifier.
- Skill-new workflow.

Verification:

- All 28 skills present.
- Skill docs and embedded text match expected content.
- Custom skill precedence tested.
- Classifier has golden tests.

### Phase 10 — MCP

Goal: implement MCP server registration, connection, tool listing, tool gating, and execution.

Outputs:

- MCP config schema.
- Built-in server configs.
- stdio client.
- streamable HTTP client.
- Manager lifecycle.
- Skill-gated tool exposure.
- `/mcp` subcommands.

Verification:

- JSON-RPC tests.
- Fake MCP server integration tests.
- Env override tests.
- Skill gating tests.
- Tool enable/disable tests.
- Docs for every server and override.

### Phase 11 — Reverse-engineering workspace and memory

Goal: implement `.rehamr/` project state and RE-specific workflows.

Outputs:

- `/init-re`
- `/status-re`
- `REPHAMR_STATE.md`
- Evidence workspace directories.
- Memory prompt injection.

Verification:

- Idempotent initialization tests.
- State file template tests.
- Token budget tests.
- User docs for memory lifecycle.

### Phase 12 — Doctor, install, update, and release

Goal: implement operational parity.

Outputs:

- `/doctor`
- install scripts
- release config
- self-update
- checksum verification
- devcontainer
- CI workflow

Verification:

- Doctor section tests.
- Platform matrix tests.
- Release asset naming tests.
- Checksum mismatch tests.
- Install script shellcheck/PowerShell checks.

### Phase 13 — Full parity hardening

Goal: close every parity row.

Outputs:

- Complete parity matrix.
- Complete docs matrix.
- Complete test coverage report.
- Security audit report.
- Performance baseline.
- Known limits document with only verified limits.

Verification:

- 100% test coverage gate passes.
- 100% docs coverage gate passes.
- Every parity row is passed or explicitly removed by user decision.
- No unsupported behavior pretends success.

### Phase 14 — RecompHamr 2.0 release candidate

Goal: produce a release candidate with no new features beyond parity.

Outputs:

- Tagged RC.
- Release notes.
- Migration guide.
- Install artifacts.
- Checksums.
- User docs site or packaged docs.

Verification:

- Fresh clone install test.
- Local LM Studio/Ollama/OpenAI-compatible smoke tests.
- MCP fake server smoke test.
- Cross-platform launch tests.
- User-doc command walkthrough.

### Phase 15 — Post-parity feature planning

Goal: only after parity, decide new features.

Possible directions:

- Richer RE project dashboards.
- Safer tool permission prompts.
- Session export/import.
- Optional desktop wrapper.
- Optional Rust binary-analysis helpers.
- Optional ACP or editor integration.
- Better model capability probes.

No item in this phase is allowed before Phase 14 closes.

## CI Gates

Minimum required jobs:

- `format`
- `lint`
- `docscheck`
- `test-linux`
- `test-macos`
- `test-windows`
- `race-linux`
- `fuzz-smoke`
- `coverage-union`
- `security-scan`
- `dependency-audit`
- `golden-parity`
- `release-dry-run`

Required commands should be wrapped in a single documented entrypoint:

```bash
make verify
```

or:

```bash
just verify
```

The exact command must be in `AGENTS.md` so Codex, Claude Code, local agents, and human contributors all run the same checks.

## Final Report Format for Every Agent Session

Every agent session must end with:

```md
## Changed
<Files and behavior changed.>

## Documented
<User docs and developer docs updated.>

## Verified
<Commands run and exact results.>

## Coverage
<Test and docs coverage status.>

## Security
<Security checks or relevant boundary notes.>

## Known limits
<Only verified limits, blocked facts, or none.>
```

No session can claim completion without this report.

## Release Criteria

RecompHamr 2.0 can be called complete only when:

- All RecompHamr 1.x parity rows pass.
- All slash commands are implemented and documented.
- All built-in tools are implemented and documented.
- All MCP built-ins are implemented and documented.
- All 28 embedded skills are present and documented.
- `.rehamr/` behavior is secure and documented.
- Config profiles and env overrides match parity expectations.
- TUI behavior is covered by tests.
- Install and update paths are covered by tests.
- User docs and developer docs pass 100% coverage checks.
- Code coverage gate passes.
- Security audit gate passes.
- No placeholder-only files, fake success paths, or vague future-work comments exist.
- The final known-limits document contains only verified limitations.

## First Implementation Order

Start with the non-code foundation:

1. Create `AGENTS.md`.
2. Create docs tree.
3. Create goal template and docs coverage checker.
4. Generate current RecompHamr 1.x inventory.
5. Create parity matrix.
6. Freeze features.
7. Only then begin package skeleton and code.

This order matters because the rewrite is meant to be AI-buildable. Agents need the operating rules, goals, docs contracts, and parity matrix before they touch product code.

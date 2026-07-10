# Traceability Matrix

| Requirement | Source | Implementation | Verification |
|---|---|---|---|
| Mandatory memory refresh | `AGENTS.md` | Process rule | `make docscheck` verifies memory docs exist. |
| Phase goal packets | User plan and workflow goal system | `docs/dev/04_Workflows/PhaseGoals.md` | `make docscheck`; manual memory refresh. |
| Next 20 phase roadmap reconciliation | User-provided Next 20 Phases plan; workflow phase split | `docs/dev/07_ProjectManagement/Next20PhasePlan.md`; `PhaseGoals.md`; status reports | `make docscheck`; `make verify`. |
| Diagnostic-only runtime | Workflow Phase 2 | `cmd/recomphamr`, `internal/app` | `go test ./...`, diagnostic command. |
| Phase 2 support infrastructure | Workflow Phase 2 | `internal/logging`; `internal/testharness`; architecture docs | `make verify`. |
| Phase 0 prior waiting state | Workflow Phase 0 | Historical memory docs | Superseded by source inventory row. |
| No product features before parity | Workflow freeze | Unsupported default runtime | App test checks unsupported default. |
| Phase 0 inventory complete | User reference URL | `internal/parity`; inventory docs | `go test ./internal/parity`; `make verify`. |
| 100% coverage gate | Coverage policy | `internal/covercheck`; `Makefile` | `make verify`. |
| Separation of concerns gate | User direction; architecture docs | `internal/archcheck`; `SeparationOfConcerns.md`; `Makefile` | `make archcheck`; `make verify`. |
| Config parity | Phase 3; RecompHamr 1.x `internal/config` evidence | `internal/config`; `configuration.md`; `ConfigParity.md` | `go test ./internal/config`; `make verify`. |
| Workspace parity subset | Phase 3; RecompHamr 1.x `internal/project` evidence | `internal/project`; `memory.md`; `MemoryParity.md` | `go test ./internal/project`; `make verify`. |
| Memory prompt injection subset | Phase 11 reference behavior | `project.LoadMemory`; `llm.WithProjectMemory`; `agent.Loop` memory fields; `memory.md`; `MemoryParity.md` | project, LLM, and agent memory tests; `make verify`. |
| LLM streaming parity | Phase 4; RecompHamr 1.x `internal/llm` evidence | `internal/llm`; `Protocols.md`; `ErrorHandling.md` | `go test ./internal/llm`; `make verify`. |
| Context packing parity | Phase 4; RecompHamr 1.x `internal/ctx` evidence | `internal/llm` packing helpers; `ContextPacking.md` | context packing tests; `make verify`. |
| Provider header/error parity | Phase 4; RecompHamr 1.x `internal/cloud` evidence | `BudgetFromHeaders`; `ContextWindowFromHeaders`; provider error classes | LLM header/error tests; `make verify`. |
| Tool schemas and Windows shell naming | Phase 5; user Windows-shell direction; RecompHamr 1.x `bash` evidence | `internal/tools`; `docs/user/tools.md`; `ToolRuntime.md`; `ToolParity.md` | `go test ./internal/tools`; `make verify`. |
| Tool security boundaries | Phase 5 mandatory gates | Shell timeout cap, cancellation handling, read truncation, cache path containment, GitHub URL validation, HTTP(S)-only reference fetches | Tool security tests; `make verify`. |
| Agent loop foundation | Phase 6; RecompHamr 1.x TUI loop/nudge evidence | `internal/agent`; `AgentLoop.md`; `docs/user/troubleshooting.md` | `go test ./internal/agent`; `make verify`. |
| Agent loop cancellation and no-infinite-loop policy | Phase 6 mandatory gates | Context checks, model-round cap, repeated-failure nudge, runaway-tool nudge, empty-reply blocked retry, verification nudge | Agent loop tests; `make verify`. |
| TUI shell foundation | Phase 7; RecompHamr 1.x TUI behavior inventory | `internal/tui`; `TUIArchitecture.md`; `TUIParity.md` | `go test ./internal/tui`; `make verify`. |
| Unique opencode-inspired TUI | User direction; official OpenCode docs confirm terminal TUI workflow | `internal/tui` initiative layout, evidence rail, compact status band, Bubble Tea adapter | TUI golden/model tests; `make verify`. |
| TUI separation and debug redaction | Phase 7 mandatory gates | Pure TUI state, Bubble Tea adapter, `internal/security` redaction, no agent/tool ownership | `make archcheck`; TUI redaction tests; `make verify`. |
| Slash command foundation | Phase 8; RecompHamr 1.x slash registry evidence | `internal/commands`; `CommandParity.md`; `docs/user/commands.md` | `go test ./internal/commands`; `make verify`. |
| Command docs/help coverage | Phase 8 mandatory gates | Registry metadata, generated help, generated markdown, examples, side effects, error classes | Command help/docs tests; `make verify`. |
| Skills foundation | Phase 9; RecompHamr 1.x skill evidence | `internal/skills`; embedded skill markdown; `docs/user/skills.md`; `SkillParity.md` | `go test ./internal/skills`; `make verify`. |
| Skill-new workflow | Phase 9; RecompHamr 1.x `/skill-new` evidence | `skills.NewDraft`; `skills.ScaffoldCustomSkill`; `/skill-new` command fetch/cache flow | command and skills tests; `make verify`. |
| MCP JSON-RPC protocol foundation | Phase 10; RecompHamr 1.x `internal/mcp/protocol.go` evidence | `internal/mcp/protocol.go`; `MCPArchitecture.md`; `Protocols.md`; `MCPParity.md` | `go test ./internal/mcp`; `make verify`. |
| MCP stdio and HTTP transport foundation | Phase 10; RecompHamr 1.x stdio and streamable HTTP client evidence | `internal/mcp/transport.go`; fake stream tests; fake HTTP server tests | transport tests; `make verify`. |
| MCP manager lifecycle | Plan Phase 11; RecompHamr 1.x `internal/mcp/manager.go` evidence | `internal/mcp/manager.go`; `MCPArchitecture.md`; `MCPParity.md` | manager tests; `make verify`. |
| `/mcp` lifecycle command dispatch | Plan Phase 11; slash command parity evidence | `internal/commands`; manager-backed command environment | command tests; `make verify`. |
| Doctor local diagnostics | Phase 12; slash command parity evidence | `internal/doctor`; `/doctor` dispatch; `DoctorParity.md`; `docs/user/doctor.md` | doctor and command tests; `make verify`. |
| Release asset naming, local binary builds, archive creation, checksum manifest generation, and checksum verification | Phase 12; release criteria and workflow evidence | `internal/release`; `ReleaseParity.md`; `ReleaseWorkflow.md`; `docs/user/install.md` | release package tests; `make verify`. |
| Install scripts, release config, devcontainer, CI, and self-update dry-run | Phase 12 operational parity | `scripts/install.ps1`; `scripts/install.sh`; `.goreleaser.yaml`; `.devcontainer/devcontainer.json`; `.github/workflows/verify.yml`; `internal/release`; `internal/update` | release, update, doctor, and archcheck tests; `make verify`. |
| Product runtime wiring | Next 20 Phase 14; app wiring contract | `cmd/recomphamr`; `internal/app`; `internal/config`; `internal/project`; `internal/commands`; `internal/mcp`; `internal/tui` | app and command-line tests; bare startup; diagnostic command; `make verify`. |
| Interactive fake runtime smoke | Next 20 Phase 15; runtime hardening plan | `internal/app.RunSmoke`; injected fake model and tool runner; composed TUI slash command dispatch | app smoke tests for slash command, prompt/tool loop, cancellation, memory injection, transcript render; `make verify`. |
| Full parity matrix closure | Next 20 Phase 19; Workflow Phase 13 | `docs/dev/05_FeatureParity/ParityClosure.md`; updated parity docs and docs index | stale wording scan; placeholder-policy scan; `make verify`; diagnostic command. |
| Security hardening audit | Next 20 Phase 20; Workflow Phase 13 security audit | `internal/project` symlink refusal; `SecurityAudit.md`; updated security docs | project symlink regression tests; security keyword scan; `make verify`; diagnostic command. |
| Documentation coverage hardening | Next 20 Phase 21; Workflow Phase 13 docs coverage | `internal/docscheck`; docs coverage policy; updated user docs terms | docscheck 100% coverage; `make docscheck`; `make verify`; diagnostic command. |
| Cross-platform validation | Next 20 Phase 22; Workflow Phase 13 platform validation | `PlatformMatrix.md`; release target matrix; PowerShell/pwsh shell selection docs; install script marker validation; pure TUI render contract | targeted platform package tests; `go env GOOS GOARCH`; `make verify`; diagnostic command; placeholder-policy scan. |
| Performance and local-first budgeting | Next 20 Phase 23; Workflow Phase 13 performance baseline | Benchmarks in `internal/llm`, `internal/tui`, `internal/mcp`, `internal/tools`, and `internal/app`; `PerformanceBenchmarks.md` | benchmark command with `-benchmem`; affected package coverage tests; `make verify`; diagnostic command. |
| User walkthrough and migration guide | Next 20 Phase 24; Workflow Phase 14 release candidate prep | `docs/user/walkthrough.md`; `docs/user/migration.md`; README and docs index links | `make verify`; diagnostic command; placeholder-policy scan; docs hash comparison. |
| Release candidate preparation | Next 20 Phase 25; Workflow Phase 14 RC prep | `ReleaseCandidate.md`; `KnownLimits.md`; user RC notes; changelog; release workflow and criteria docs | release/update/doctor package coverage tests; `make verify`; diagnostic command; placeholder-policy scan; docs hash comparison. |
| RC soak and bugfix freeze | Next 20 Phase 26; Workflow Phase 14 stabilization | `RCSoak.md`; updated roadmap, parity, status, and known-limit memory | repeated `make verify`; diagnostic command; release/update/doctor coverage tests; placeholder-policy scan. |
| Stable release gate | Next 20 Phase 27; stable publication gate | `StableRelease.md`; updated roadmap, parity, status, release criteria, and known limits | `make verify`; local artifact build for six targets; `dist/SHA256SUMS` generation and verification; Windows installer smoke from `recomphamr_windows_amd64.zip`; local `v2.0.0` tag evidence; `git tag --list`; placeholder-policy scan. |

# Module Contracts

Modules must expose small, testable contracts. UI code must not own agent behavior. Tools must not bypass security helpers. Config and project state must use explicit validation and atomic persistence.

The detailed import boundary rules live in `docs/dev/03_Architecture/SeparationOfConcerns.md` and are enforced by `make archcheck`.

Phase 2 architecture skeleton includes `internal/logging` for redacted diagnostics and `internal/testharness` for deterministic golden/parity helpers. These packages are support infrastructure only; they must not import product behavior.

Next 20 Phase 14 app wiring contract:

- `internal/app` owns process modes and local runtime composition.
- `internal/app.ComposeRuntime` may bootstrap/load config, load optional project
  memory, create the command environment, create the MCP manager without
  autoconnect, and create pure TUI state.
- `internal/app` must not own slash command behavior, TUI update/render policy,
  model-tool loop policy, tool execution, config schema rules, MCP protocol
  behavior, or project workspace persistence.
- Bare `recomphamr` startup may report a deterministic local launch summary but
  must not probe networks, call a model backend, autoconnect MCP servers, or
  pretend an agent turn completed.
- `internal/app.RunSmoke` may execute deterministic fake-runtime smoke with
  caller-injected model and tool runner dependencies. It may coordinate the
  existing agent loop and pure TUI renderer for tests, but it must not call real
  model backends, execute real tools, autoconnect MCP servers, or launch a
  terminal process.

Phase 3 config and workspace contracts:

- `internal/config` owns `.rehamr/config.yaml`, model profiles, strict YAML decoding, atomic saves, symlink refusal, `RECOMPHAMR_URL`, and config-key docs.
- `internal/project` owns `.rehamr/` project workspace creation, status summaries,
  and `REPHAMR_STATE.md` file loading. It calls `internal/config` for bootstrap
  and must not import commands, TUI, agent, tools, skills, LLM, or MCP.
- Command handlers may call `internal/config` and `internal/project`, but those packages must stay independently testable without command or terminal state.

Memory runtime contract:

- `internal/project.LoadMemory` returns file content and explicit missing
  workspace or missing memory errors. It must not shape chat messages.
- `internal/llm.WithProjectMemory` shapes already-loaded memory into model
  context and owns token-budget trimming. It must not read files.
- `internal/agent.Loop` may apply caller-provided memory fields before a model
  turn. It must not import `internal/project` or load workspace files itself.

Phase 6 agent loop contract:

- `internal/agent` owns model-tool turn execution, transcript pairing,
  repeated-failure nudges, runaway-tool nudges, empty-reply retries,
  verification nudges, and cancellation ownership.
- `internal/agent` may depend on `internal/llm` types and caller-provided tool
  runner interfaces. It must not import TUI, command, or concrete tool packages.
- TUI and command packages must render or invoke the agent loop instead of
  reimplementing its turn policy.

Phase 7 TUI contract:

- `internal/tui` owns terminal state, rendering, key translation, prompt
  history, paste chips, cancellation/quit UI intent, and debug redaction.
- `internal/tui` may import `internal/commands` for slash-command dispatch and
  `internal/security` for redaction. It must not import `internal/agent`,
  `internal/tools`, `internal/config`, `internal/project`, `internal/llm`, or
  `internal/mcp`.
- Bubble Tea integration is limited to a thin adapter around the pure TUI state
  so tests can exercise behavior without launching a terminal.

Phase 8 command contract:

- `internal/commands` owns slash command registry metadata, generated help,
  examples, side-effect docs, error-class docs, and command dispatch.
- Command handlers may coordinate `internal/config`, `internal/project`,
  `internal/skills`, `internal/mcp`, `internal/parity`, and `internal/tools`.
  They must not own terminal rendering or the agent loop.
- Future-phase behavior must return explicit `unsupported`, `unverified`, or
  `blocked` output instead of pretending success.

Phase 9 skills contract:

- `internal/skills` owns embedded skill inventory, custom skill discovery,
  custom-over-embedded precedence, exact/case-insensitive/`.md` resolution,
  active-list rendering, audit classification, fetched-body classification, and
  approved custom skill scaffolding.
- `internal/skills` must not perform network fetches. Command or app wiring must
  fetch user-approved URLs, pass the body into `skills.NewDraft`, cache fetched
  content for review, and only write `.rehamr/skills/<name>.md` after explicit
  approval.
- `/skill-new` may cache fetched source under `.rehamr/fetched/`, but it must not
  silently activate, execute, or install fetched skills.

Phase 10 MCP protocol contract:

- `internal/mcp` owns MCP built-in registration metadata, JSON-RPC protocol
  types, MCP initialize/tools payloads, injected stdio transport, and streamable
  HTTP request/response transport.
- `internal/mcp` protocol transports must be testable without real external MCP
  servers. Stdio process spawning, persistent manager lifecycle, and agent tool
  exposure are separate runtime responsibilities.
- User-visible MCP lifecycle commands must continue to report explicit
  `unsupported`, `unverified`, or `blocked` output until manager behavior is
  implemented and tested.

Plan Phase 11 MCP manager contract:

- `internal/mcp.Manager` owns registered MCP state, connector-driven lifecycle,
  enabled-tool allowlists, skill-gated visibility, status formatting, and full
  `server.tool` calls.
- `internal/commands` may call a provided `*mcp.Manager` from command
  environment for `/mcp` lifecycle commands. It must return `unsupported:` when
  manager wiring is absent and `blocked:` when manager actions fail.
- Stdio process spawning, app autostart, persistent user MCP config files, and
  agent-loop MCP exposure remain outside this contract.

Phase 12 doctor contract:

- `internal/doctor` owns offline local diagnostics for runtime, workspace,
  config, memory, skills, tools, MCP registrations, and Phase 12 operational
  file validation.
- `internal/doctor` may read local files through owning packages but must not
  mutate config, initialize workspaces, probe networks, launch external tools,
  install dependencies, update binaries, execute installers, or execute release
  artifacts.
- `internal/commands` may render `doctor.Run` for `/doctor`.

Phase 12 release checksum contract:

- `internal/release` owns canonical release artifact names, local `go build`
  orchestration for supported targets, local archive creation from already-built
  binaries, local `SHA256SUMS` generation and parsing, artifact path
  validation, local file hashing, operational file validation, and verification
  report formatting.
- `internal/release` must not download artifacts, fetch checksums, execute
  installers, update binaries, read project config, or own command parsing.
- `internal/update` owns local self-update dry-run planning from a verified
  release artifact. It must not replace the running executable or fetch remote
  release metadata.
- Later installer, updater, and release-automation packages may call
  `internal/release` after they independently document and test their own
  security boundaries.

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
  model-tool loop policy, concrete tool internals, config schema rules, MCP
  protocol behavior, or project workspace persistence.
- Bare `recomphamr` startup launches the Bubble Tea app and must not probe
  model backends, call a model backend, or pretend an agent turn completed
  before the user submits a prompt. MCP autostart is allowed only when explicit
  autostart metadata is present.
- `internal/app.Launch` owns the live runtime bridge from TUI submit/cancel/quit
  intent to the existing agent loop, OpenAI-compatible streaming client, and
  built-in tool dispatcher. This bridge is app wiring; the TUI remains pure and
  `internal/agent` remains unaware of concrete tools.
- `--summary` reports deterministic local launch evidence without opening the
  terminal app.
- `internal/app.RunSmoke` may execute deterministic fake-runtime smoke with
  caller-injected model and tool runner dependencies. It may coordinate the
  existing agent loop and pure TUI renderer for tests, but it must not call real
  model backends, execute real tools, autoconnect stdio MCP servers, or launch a
  terminal process.

Phase 28 live runtime contract:

- Prompt submission inside the TUI creates one cancellable agent turn with the
  active model profile from `.rehamr/config.yaml`.
- Slash commands continue to execute through `internal/commands` without model
  calls.
- Built-in tool calls route through `internal/tools` using `powershell`,
  `read_file`, `write_file`, `edit_file`, `repomixr`, `recomp_reference`, and
  the `bash` compatibility alias.
- Ctrl+C during a running turn cancels the agent context; Ctrl+D or double
  Ctrl+C exits cleanly through Bubble Tea.
- MCP agent-loop exposure is supported only through connected, enabled,
  skill-visible manager tools. Autostart is explicit and may use HTTP or
  configured stdio commands.

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
  servers. Manager lifecycle and agent tool exposure remain separate runtime
  responsibilities.
- User-visible MCP lifecycle commands must report explicit `blocked` output when
  manager actions fail.

Plan Phase 11 MCP manager contract:

- `internal/mcp.Manager` owns registered MCP state, connector-driven lifecycle,
  enabled-tool allowlists, skill-gated visibility, status formatting, and full
  `server.tool` calls.
- `internal/commands` may call a provided `*mcp.Manager` from command
  environment for `/mcp` lifecycle commands. It must return `unsupported:` when
  manager wiring is absent and `blocked:` when manager actions fail.
- Stdio process spawning and persistent user MCP config files are implemented in
  `internal/mcp`. App autostart is limited to configs with explicit autostart
  metadata. Agent-loop MCP exposure is owned by the Phase 29 app wiring contract
  and must use manager skill gates and allowlists.

Phase 29 MCP agent integration contract:

- `internal/app` may ask `internal/mcp.Manager` for `ToolsForSkills` when a live
  prompt starts and convert those tool definitions into LLM function schemas.
- `internal/app` may dispatch `server.tool` calls to `Manager.CallTool` as part
  of the app-owned tool runner.
- `internal/app` must not bypass MCP skill gates, enabled-tool allowlists,
  connection state, or MCP protocol ownership.
- MCP tool-level errors must be visible to the agent loop as tool failures.

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

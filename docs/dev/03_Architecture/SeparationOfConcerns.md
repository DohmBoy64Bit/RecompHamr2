# Separation Of Concerns

RecompHamr packages must stay small, directional, and independently testable. Feature work must add behavior to the owning package instead of reaching across layers.

## Allowed Dependency Direction

- `cmd/recomphamr` may call only `internal/app`.
- `internal/app` wires process modes and may compose config, project memory,
  command environment, MCP manager, TUI state, live model clients, and the
  built-in tool dispatcher required by the agent loop. It may run deterministic
  fake-runtime smoke through injected `internal/agent` and `internal/llm`
  contracts. It must not own slash command behavior, TUI render policy, tool
  implementation internals, MCP protocol behavior, stdio MCP process spawning,
  or run model turns during startup.
- `internal/tui` may call `internal/commands` and must not execute tools, mutate config, own the agent loop, or manage MCP lifecycles.
- `internal/commands` may coordinate config, project, skills, tools, MCP registry, and parity metadata.
- `internal/doctor` may read config, project memory, skills, tool schemas, and MCP registrations for offline diagnostics; it must not mutate state, probe networks, launch tools, own command parsing, or perform release/update work.
- `internal/agent` may depend on `internal/llm` and tool runner interfaces, not concrete TUI or command packages.
- `internal/project` may depend on `internal/config` for workspace bootstrap.
- `internal/logging` may depend on `internal/security` for redaction and must not own runtime policy.
- `internal/release` owns local manifest and checksum verification only; it must not download artifacts, run installers, update binaries, read project config, or own command parsing.
- Future extension packages are unsupported until an approved goal adds source,
  docs, tests, and archcheck rules. They must follow
  `ExtensionArchitecture.md`: app wiring owns side effects, agent code consumes
  interfaces, TUI renders snapshots only, and domain packages keep ownership of
  config, tools, MCP, release, security, and project state.
- Core packages such as `config`, `llm`, `mcp`, `parity`, `security`, `skills`, `testharness`, and `tools` must not import higher-level RecompHamr packages.

## Enforcement

`make verify` runs `internal/archcheck`, which parses Go imports and fails if package boundaries drift. New internal packages must be documented here and added to the archcheck allowlist in the same change.

## Rationale

This separation keeps the terminal interface replaceable, lets local-model agents reason about small files, prevents tool execution from bypassing security helpers, and keeps parity tests focused on one subsystem at a time.

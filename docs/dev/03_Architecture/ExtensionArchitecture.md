# Extension Architecture

Phase 36 is architecture planning only. It records boundaries for future
post-parity extension work approved by `FeatureIntake.md`. No extension runtime
behavior is implemented by this document.

## Approved Planning Scope

The only approved Phase 36 candidate is `FI-001 Extension Architecture
Planning`. The following areas may receive future goal packets, but remain
`unsupported` until a separate approved implementation phase exists:

- permission prompts for shell, filesystem, network, and MCP tool actions;
- session export and import with redaction;
- optional Rust helper processes for binary analysis;
- external analyzer adapters;
- richer MCP integrations;
- plugin-style local tools;
- future UI surfaces beyond the current Bubble Tea terminal app.

## Ownership Boundaries

Future extensions must follow the existing package direction:

- `internal/app` may wire extension services into runtime composition after a
  feature-specific goal approves the behavior.
- `internal/agent` may consume extension capabilities only through explicit
  interfaces owned by the caller; it must not import concrete extension
  packages.
- `internal/tui` may render extension state snapshots, prompts, or modals, but
  it must not execute extension actions or persist extension config.
- `internal/commands` may expose user commands for extensions only when the
  command registry includes usage, side effects, examples, and error classes.
- `internal/mcp`, `internal/tools`, `internal/release`, `internal/project`, and
  `internal/config` keep ownership of their existing domains. Extension
  packages must call documented APIs instead of duplicating those behaviors.
- `pkg/` remains unavailable unless an architecture decision approves a stable
  public API with documentation, tests, and compatibility policy.

Proposed future package names such as `internal/extensions`,
`internal/permissions`, `internal/sessionio`, or `internal/analyzers` are
unimplemented planning labels. They are not supported import paths until source,
docs, tests, and `archcheck` rules are added in a later goal.

## Configuration Contract

Every future extension that needs configuration must define:

- file path and format, such as a documented `.rehamr/<feature>.yaml` or JSON
  file;
- every key, type, default, validation rule, and environment override;
- strict decoding behavior and unknown-field failure policy;
- symlink, permission, traversal, and atomic-write rules;
- user examples for creating, editing, validating, and troubleshooting the
  config file;
- generated or registry-backed help text for every command that reads or
  mutates the config.

No extension may silently read arbitrary config files or use hidden defaults.
When config is missing, the user-visible state must be `unsupported`,
`unverified`, or `blocked` with an actionable reason.

## Protocol And Process Contract

Subprocess helpers, Rust binaries, external analyzers, and plugin-style tools
must be treated as untrusted local processes unless a later security review
proves narrower permissions.

Required future protocol rules:

- JSON-RPC or line-delimited JSON is preferred for local helper protocols.
- Every request and response type must be documented and tested with golden
  payloads.
- Helpers must receive bounded input and return bounded output.
- Cancellation must terminate the request and, where supported by the platform,
  the helper process tree.
- Helper paths must be configured explicitly or discovered from documented,
  deterministic locations.
- Network access requires a user-visible reason and docs for every endpoint.
- Release packaging must document helper binaries, checksums, and platform
  support before users can rely on them.

## Permission Prompt Design Boundary

Permission prompts may be designed next, but they must not claim sandboxing
that does not exist. A prompt can explain and gate an action; it cannot make a
dangerous command safe by itself.

Any implementation must cover:

- action class: shell, file read, file write, network fetch, MCP call, helper
  process, or export/import;
- exact command, file path, URL, server/tool name, or destination;
- allow once, deny, and cancel semantics;
- noninteractive behavior for tests and automation;
- transcript and debug redaction;
- help text and troubleshooting docs.

## Session Export And Import Boundary

Session export/import remains unsupported until a future goal defines the file
format and redaction contract. A valid design must exclude private reasoning,
redact configured secrets, preserve evidence labels, document every exported
field, and include golden round-trip tests.

## Test Strategy

Every extension implementation must add tests at the layer it changes:

- unit tests for parsing, validation, and policy decisions;
- golden tests for protocol payloads, command help, and rendered output;
- integration tests with fake helper processes or fake adapters;
- security regression tests for traversal, symlinks, loose permissions,
  cancellation, redaction, and network boundaries;
- platform tests for Windows-first paths and process behavior;
- docs coverage tests for commands, config keys, environment variables, helper
  files, generated artifacts, and exported symbols.

`make verify` remains the canonical gate. No extension goal may close below
100% statement coverage for implemented Go code.

## Phase 36 Decision

The current extension architecture is documented but not implemented. The next
runtime extension must open a new goal packet that references this document,
updates `FeatureIntake.md`, and explicitly names the user value, package
boundaries, config/help/docs impact, security review, and verification commands.

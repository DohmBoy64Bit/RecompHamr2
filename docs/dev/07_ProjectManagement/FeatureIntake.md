# Feature Intake Register

Phase 35 controls post-parity feature selection. This register does not
implement features. It decides which ideas are approved for a future goal packet
and which remain deferred.

## Intake Gates

Every candidate must satisfy all gates before implementation:

- Stable release evidence exists in `StableRelease.md`.
- Phase 28 live runtime, Phase 29 live MCP integration, Phases 30-34 TUI and
  Windows executable hardening, and publication evidence are complete.
- User value is concrete and tied to RecompHamr's reverse-engineering mission.
- Scope is bounded to named packages, docs, commands, config keys, or files.
- Security impact is recorded, including filesystem, shell, network, MCP, and
  secret-handling risks.
- Configuration impact lists every new config key, file, environment variable,
  default, validation rule, example, and help text requirement.
- Documentation impact lists user docs, developer docs, generated help,
  docstrings, examples, and traceability rows.
- Test impact preserves 100% statement coverage and includes relevant unit,
  integration, golden, security, platform, and docs coverage tests.
- Approval status is one of `approved`, `deferred`, or `rejected`.
- Approved work must get its own goal packet before implementation begins.

## Approved Candidates

| ID | Candidate | User Value | Risk | Required Next Step |
|---|---|---|---|---|
| FI-001 | Extension Architecture Planning | Defines safe boundaries for optional Rust helpers, analyzers, richer MCP integrations, plugin-style tools, and future UI surfaces before code exists. | Medium: poor boundaries could violate separation of concerns or create unsupported config surfaces. | Open Phase 36 as an architecture-only goal with ADRs, config/help coverage plan, security analysis, and no implementation. |
| FI-002 | Permission Prompt Design | Gives users clearer control before shell, file-write, network, and MCP tool actions run with local permissions. | High: prompts must not create fake safety or bypass cancellation/tool policies. | Create a design goal covering command/tool/MCP permission states, help text, tests, and security docs before implementation. |
| FI-003 | Session Export And Import Design | Lets users preserve and move transcripts, evidence, and project memory without copying workspace internals manually. | Medium: exports can leak secrets or unsupported private reasoning if redaction is incomplete. | Create a design goal covering redaction, file format, config, docs, and golden tests. |

## Deferred Candidates

| ID | Candidate | Reason Deferred | Reconsider When |
|---|---|---|---|
| FI-004 | Rich Reverse-Engineering Dashboard | Valuable, but risks bloating the small-context terminal mission if designed before extension boundaries. | Phase 36 defines UI/data ownership boundaries. |
| FI-005 | ACP Or Editor Integration | Useful for editor workflows, but it introduces protocol and lifecycle complexity outside the current terminal app. | Extension architecture identifies supported protocol boundaries. |
| FI-006 | Optional Desktop Shell | Explicitly deferred because a desktop UI multiplies scope and was prohibited before parity completion. | Terminal app has stable permission, export, and extension boundaries. |
| FI-007 | Rust Binary-Analysis Helpers | Potentially useful for future parsers/lifters, but requires subprocess, config, packaging, and docs architecture first. | Phase 36 approves helper boundaries and release packaging rules. |

## Rejected Candidates

No candidates are rejected in Phase 35. Rejections must include evidence, owner,
and the rule or risk that caused rejection.

## Phase 35 Decision

Phase 35 approves Phase 36 as the next phase. Phase 36 is architecture planning
only: it may add ADRs, contracts, config/help coverage plans, security analysis,
and test strategy, but it must not implement extension runtime behavior.

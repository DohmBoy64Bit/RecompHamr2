# Memory Parity

Memory parity covers `.rehamr/` initialization, `REPHAMR_STATE.md`, state
persistence, memory prompt injection, and status reporting.

Phase 3 implemented the secure workspace bootstrap and status-reporting subset:

- Directories: `logs/`, `evidence/`, `repos/`, `functions/`, `formats/parsers/`, `formats/tests/`, `recomp/`, `decomp/`, and `skills/`.
- Files: `PROJECT.md`, `REPHAMR_STATE.md`, `EVIDENCE.md`, `HYPOTHESES.md`, `BLOCKERS.md`, `CHANGELOG.md`, `COMMANDS.md`, `TOOLCHAIN.md`, `MODELS.md`, `mcp.json`, function ledgers, format ledgers, recomp ledgers, decomp ledgers, and `skills/active.md`.
- `/init-re` is idempotent and does not overwrite existing project files.
- `/status-re` summarizes tracked files, reports missing files, and truncates long content without splitting UTF-8.

Runtime memory integration now covers the local, testable subset:

- `project.LoadMemory(projectDir, maxBytes)` reads `.rehamr/REPHAMR_STATE.md`,
  returns explicit missing-workspace or missing-memory errors, defaults to a
  24 KiB read cap, and preserves UTF-8 when truncating.
- `llm.WithProjectMemory(messages, source, memory, maxTokens)` injects memory
  into the primary system context, clones caller history, labels the source, and
  trims the memory body by token budget.
- `agent.Loop` accepts `ProjectMemory`, `ProjectMemorySource`, and
  `ProjectMemoryMaxTokens` so runtime callers can add workspace memory without
  importing filesystem packages into the loop.
- Bare executable startup composes runtime state through `internal/app` and
  reports verified, unsupported, or blocked memory status from
  `.rehamr/REPHAMR_STATE.md`.

Fake-runtime smoke now verifies that startup-loaded memory reaches the injected
agent loop through `internal/app.RunSmoke`. Still `unsupported`: real backend
model turns that consume startup-loaded memory.

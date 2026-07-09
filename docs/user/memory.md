# Memory

`/init-re` creates a project workspace under `.rehamr/` with persistent memory
and evidence files. The command registry and `internal/project` package support
secure, idempotent creation during Phase 3.

Important files:

- `.rehamr/REPHAMR_STATE.md`: current phase, goal, blockers, commands, symbols, and learned patterns.
- `.rehamr/EVIDENCE.md`: confirmed facts with source evidence.
- `.rehamr/HYPOTHESES.md`: unconfirmed ideas that need proof.
- `.rehamr/BLOCKERS.md`: missing tools, files, user decisions, or runtime failures.
- `.rehamr/CHANGELOG.md`: workspace changes by date.
- `.rehamr/functions/inventory.csv`: function and symbol ledger.
- `.rehamr/formats/inventory.md`: file format evidence.
- `.rehamr/recomp/runtime_gaps.md`: recompilation runtime gaps.
- `.rehamr/decomp/symbols.md`: decompilation symbol notes.
- `.rehamr/PROJECT.md`, `.rehamr/COMMANDS.md`, `.rehamr/TOOLCHAIN.md`,
  `.rehamr/MODELS.md`, `.rehamr/mcp.json`, and `.rehamr/skills/active.md`:
  project identity, verified commands, toolchain notes, model observations, MCP
  config, and active skill notes.

`/status-re` summarizes those files and marks missing entries. Existing files
are never overwritten by `/init-re`.

Runtime memory support loads `.rehamr/REPHAMR_STATE.md` through
`internal/project.LoadMemory` and injects it into model context through
`internal/llm.WithProjectMemory` and `internal/agent.Loop`. The loader caps
memory at 24 KiB by default, preserves UTF-8 boundaries, and reports
`blocked`/`unsupported`-style errors for missing workspace or missing memory
instead of guessing.

Callers can configure a smaller prompt budget by passing
`ProjectMemoryMaxTokens` to the agent loop. The cap applies to the memory body;
the system note still tells the model to verify facts against current files
before changing behavior.

Example:

```go
mem, err := project.LoadMemory(projectDir, project.DefaultMemoryMaxBytes)
loop := agent.Loop{
    Model: model,
    RunTool: runTool,
    ProjectMemory: mem.Content,
    ProjectMemorySource: mem.Path,
    ProjectMemoryMaxTokens: 2048,
}
```

Bare executable startup attempts to load `.rehamr/REPHAMR_STATE.md` and reports
`verified`, `unsupported`, or `blocked` memory status in the launch summary. It
does not start a live model turn during startup. Agents must still refresh
project memory from `AGENTS.md`, `recomphamr_2_rewrite_workflow.md`, and
relevant docs before every phase or code-changing task.

# Command Parity

Phase 8 implements the RecompHamr 1.x slash command set in `internal/commands`.
The registry is the source for generated help, examples, side effects, and error
classes.

| Command | Phase 8 Status | Evidence |
|---|---|---|
| `/clear` | Implemented as conversation-reset intent. | Registry, tests, user docs. |
| `/models` | Lists profiles and switches active profile through config. | Config tests and command tests. |
| `/skills` | Lists embedded/custom/active skill counts. | Skills package and command tests. |
| `/skill` | Resolves exact, case-insensitive, and `.md` names through skills package. | Command tests. |
| `/skill-audit` | Classifies skill names into documented template categories. | Skills audit tests and command tests. |
| `/skill-new` | Fetches HTTP(S) skill Markdown, classifies it, and caches `.rehamr/fetched/<name>.md` for manual approval. | Command tests, skills tests, user docs. |
| `/init-re` | Creates `.rehamr/` workspace via project package. | Project tests and command tests. |
| `/status-re` | Summarizes `.rehamr/` state via project package. | Project tests and command tests. |
| `/doctor` | Reports offline local diagnostics. | `internal/doctor`, command dispatch tests, and `DoctorParity.md`; install/update/release operational file checks are covered. |
| `/mcp` | Lists built-ins and dispatches manager-backed connect, disconnect, tools, enable, and disable actions when MCP manager wiring is present. | MCP registry, manager tests, and command tests. |
| `/help` | Generates command help from registry metadata. | Help and markdown tests. |

Known phase boundaries are explicit rather than fake success: command handlers
return `unsupported`, `unverified`, or `blocked` for behavior that is not
implemented or cannot be proven.

# Golden Outputs

Golden outputs must capture RecompHamr 1.x visible behavior and RecompHamr 2.0 generated help/docs outputs. Golden updates require traceability notes.

Phase 15 runtime smoke uses deterministic TUI render strings as golden-style
evidence. The app smoke tests assert visible slash-command output, assistant
messages, tool result lines, cancellation state, and memory-injected prompt
context instead of storing external fixtures.

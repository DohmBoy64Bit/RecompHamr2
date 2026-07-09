# Coverage Requirements

100% statement coverage is mandatory for implemented Go packages. `make verify` runs `internal/covercheck`, which fails if any reported package drops below 100.0% statement coverage.

100% documentation, docstring, help, command, config, examples, and docs index coverage is mandatory immediately for implemented behavior.

`make verify` runs `make docscheck`, which now fails on missing required docs,
undocumented user-visible command/tool/config/MCP/workspace/release/help terms,
and exported Go symbols without Go doc comments.

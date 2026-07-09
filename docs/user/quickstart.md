# Quickstart

RecompHamr 2.0 can compose local product runtime state:

```sh
go run ./cmd/recomphamr
```

Bare startup loads or creates `.rehamr/config.yaml`, loads optional
`.rehamr/REPHAMR_STATE.md` memory when present, wires the slash-command
environment, creates an MCP manager without autoconnecting servers, and prepares
pure TUI state. It prints a deterministic launch summary.

Diagnostic status remains available:

```sh
go run ./cmd/recomphamr --diagnostic
```

Startup does not call a model backend, run a live prompt loop, execute tools, or
autoconnect MCP servers. Fake-runtime smoke tests cover prompt flow with
injected dependencies; real backend model turns and the interactive Bubble Tea
process remain unsupported.

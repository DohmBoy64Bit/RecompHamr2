# Configuration

Configuration is stored in `.rehamr/config.yaml`. Phase 3 supports secure
creation, strict loading, model profile selection, and process-local URL
overrides.

Create the workspace and config through the command registry with `/init-re`, or
run bare startup to bootstrap config and print a local runtime summary:

```sh
go run ./cmd/recomphamr
```

You can also verify the config package directly:

```sh
go test ./internal/config ./internal/project
```

Example:

```yaml
active: lmstudio-amd
logging: false
models:
  lmstudio-amd:
    llm: qwen/qwen3.6-35b-a3b
    url: http://localhost:1234
    key: ""
    context_size: 32768
```

Keys:

- `active`: selected profile name. If it points at a missing profile, RecompHamr chooses the first profile alphabetically.
- `logging`: enables redacted diagnostic logging when app wiring supports log files.
- `models.<name>.llm`: OpenAI-compatible model identifier.
- `models.<name>.url`: endpoint base URL.
- `models.<name>.key`: optional API key; local backends usually use `""`.
- `models.<name>.context_size`: prompt packing budget; missing or non-positive values become `32768`.

`RECOMPHAMR_URL` overrides the active profile URL for the current process only and is never saved. Config uses strict YAML: unknown fields fail to load.

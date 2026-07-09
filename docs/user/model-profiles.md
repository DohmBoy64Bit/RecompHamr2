# Model Profiles

Model profiles are implemented in `.rehamr/config.yaml`. The Phase 3 config
package creates four AMD-priority local defaults:

| Profile | Model | URL | Context |
|---|---|---|---|
| `lmstudio-amd` | `qwen/qwen3.6-35b-a3b` | `http://localhost:1234` | `32768` |
| `lmstudio-fast` | `openai/gpt-oss-20b` | `http://localhost:1234` | `32768` |
| `ollama-amd` | `qwen3.6:27b` | `http://localhost:11434` | `32768` |
| `llama-vulkan` | `qwen3.6-35b-a3b` | `http://localhost:8080` | `32768` |

Add, rename, or delete profiles by editing `.rehamr/config.yaml`. RecompHamr
does not silently restore deleted defaults after the first config is created.

Switching is handled by `/models <name>` in the command registry. The saved
`active` key changes only after `.rehamr/config.yaml` is written successfully.

Use `RECOMPHAMR_URL` for temporary endpoint changes:

```sh
RECOMPHAMR_URL=http://host.docker.internal:11434 recomphamr
```

That override affects the current process only. It is not written to
`.rehamr/config.yaml`, which prevents local test endpoints from replacing the
project's saved profile.

Phase 4 LLM behavior uses the active profile with the OpenAI-compatible
`/v1/chat/completions` endpoint. Providers may send `X-Context-Window`; when
present and sane, that value becomes the live context budget signal for packing.
`RECOMPHAMR_IDLE_TIMEOUT` controls how long the stream may stay silent between
SSE frames:

```sh
RECOMPHAMR_IDLE_TIMEOUT=90m recomphamr
```

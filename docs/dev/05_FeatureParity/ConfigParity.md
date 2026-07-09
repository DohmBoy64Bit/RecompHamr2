# Config Parity

Config parity evidence is captured from `.reference/RecompHamr1` commit
`259a450e93af48437ee23663e5ca66cdc1ab8569`, especially
`internal/config/config.go`, `internal/config/config_test.go`, `README.md`, and
`docs/profiles.md`.

Phase 3 implemented behavior:

- `.rehamr/config.yaml` is created under a real `.rehamr/` directory.
- Symlinked `.rehamr/` and symlinked `config.yaml` are refused.
- Saves are atomic through an adjacent temporary file and rename.
- Default profiles are `lmstudio-amd`, `lmstudio-fast`, `ollama-amd`, and `llama-vulkan`.
- Default `context_size` is `32768`; missing or non-positive local values are coerced to `32768`.
- Supported config keys are `active`, `logging`, `models.<name>.llm`, `models.<name>.url`, `models.<name>.key`, and `models.<name>.context_size`.
- Unknown YAML fields fail strict decode.
- A missing active profile is coerced to the first profile alphabetically when profiles exist.
- `RECOMPHAMR_URL` overrides the active URL for the process and does not persist.

Security improvement over 1.x: generated workspace and config files are written
owner-only on POSIX systems instead of world-readable defaults.

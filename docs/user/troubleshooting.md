# Troubleshooting

If a feature is missing, check whether its phase is complete in `docs/dev/07_ProjectManagement/Milestones.md`.

Use:

```sh
make verify
go run ./cmd/recomphamr --diagnostic
```

Report unsupported behavior as `unsupported`, unknown facts as `unverified`, and missing external evidence as `blocked`.

## LLM Streaming

If a model stalls after the HTTP request succeeds, set a larger inter-frame
timeout:

```sh
RECOMPHAMR_IDLE_TIMEOUT=90m recomphamr
```

Common LLM errors:

- `backend unreachable`: endpoint URL, DNS, process, or network failure.
- `invalid or expired token`: HTTP 401 from the provider.
- `budget depleted`: HTTP 402 from a budgeted provider.
- `sse payload`: provider returned malformed stream data.

RecompHamr does not switch from a local profile to a cloud profile automatically.

## Agent Loop Nudges

During long tool-using turns, RecompHamr may add an automated system note to the
transcript. These notes are not user prompts. They keep the model from repeating
the same failing tool call, continuing a runaway loop, ending with no reply, or
claiming a substantial task is done before checking the acceptance criteria.

If the loop returns `blocked`, read the last automated note and the preceding
tool result. The model must either change approach, report the blocker, or mark
the unverified behavior explicitly.

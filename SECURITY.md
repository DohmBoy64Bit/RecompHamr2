# Security Policy

Security is part of parity, not a later hardening pass.

## Required Controls

- Refuse symlinked `.rehamr/` directories and symlinked sensitive config files.
- Use owner-only permissions for sensitive config and state files.
- Write config and state files atomically.
- Reject path traversal in workspace, cache, reference, and repo-packing paths.
- Keep shell command execution bounded by timeout and cancellation.
- Redact secrets in logs, debug files, errors, docs examples, and final reports.
- Keep MCP tools off unless skill-gated or explicitly enabled.
- Do not fetch from the network without a user-visible reason.
- Never silently fall back from a local model to a cloud model.

## Reporting

Record security issues in `docs/dev/01_Memory/ProblemRegistry.md` and link them from `docs/dev/07_ProjectManagement/TraceabilityMatrix.md`.

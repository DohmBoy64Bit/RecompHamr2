# Logging Policy

Logs and debug output must redact secrets. Reasoning deltas may be displayed only according to policy and must not be persisted as private reasoning.

`internal/logging` provides deterministic diagnostic log formatting and delegates secret replacement to `internal/security`. Runtime packages must use that boundary instead of ad hoc string replacement for logs.

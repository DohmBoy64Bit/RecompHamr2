# Repository Guidelines

## Mission

RecompHamr 2.0 is a clean Go rewrite of RecompHamr 1.x for local-first terminal coding support in reverse engineering, decompilation, and static recompilation workflows. Preserve observable RecompHamr 1.x behavior before adding any new feature. Treat `recomphamr_2_rewrite_workflow.md` and the docs under `docs/dev/` as durable project memory.

## Mandatory Startup Reading

Before every phase and every code-changing task, read:

1. `AGENTS.md`
2. `recomphamr_2_rewrite_workflow.md`
3. `docs/dev/00_Governance/AntiHallucination.md`
4. `docs/dev/00_Governance/DefinitionOfDone.md`
5. `docs/dev/05_FeatureParity/ParityMatrix.md`
6. `docs/dev/07_ProjectManagement/TraceabilityMatrix.md`
7. Relevant subsystem docs for the files being changed
8. Latest verification evidence in `docs/dev/07_ProjectManagement/StatusReports.md`

If any required file is missing, stop and record `blocked:` evidence instead of guessing.

## Goal Packets And Stop Conditions

Every phase or non-trivial task starts with a written goal packet: outcome, scope, out of scope, evidence required, verification commands, and stop condition. Phase goal packets live in `docs/dev/04_Workflows/PhaseGoals.md`. Stop after three failed attempts at the same blocker. Do not continue by speculation.

## Project Structure

Current implementation uses Go:

- `cmd/recomphamr/` process entrypoint
- `internal/` private packages
- `docs/user/` user documentation
- `docs/dev/` governance, architecture, parity, testing, and memory docs
- `scripts/` helper scripts
- `testdata/` golden fixtures and parity evidence

Use `pkg/` only for stable public APIs approved by an architecture decision.

Strict separation of concerns is mandatory. Follow `docs/dev/03_Architecture/SeparationOfConcerns.md`; `make verify` runs `make archcheck` and fails on undocumented package imports.

## Build, Test, And Development Commands

- `make verify` runs the canonical local verification gate.
- `make test` runs Go tests.
- `make docscheck` verifies documentation index and required memory docs.
- `make archcheck` verifies package dependency boundaries.
- `go run ./cmd/recomphamr` launches the terminal app.
- `go run ./cmd/recomphamr --summary` prints deterministic runtime composition evidence without launching the TUI.
- `go run ./cmd/recomphamr --diagnostic` prints foundation status.

## Coding Style And Documentation

Use `gofmt`. Package names are short, lowercase, and domain-specific. Every exported package, type, function, method, constant, and variable must have a Go doc comment beginning with the exported identifier. Every command, tool schema, config key, environment variable, generated file, MCP setting, skill, error class, and user-visible behavior must have professional docs, help text where applicable, and examples when configuration is involved.

## Mandatory Coverage

100% statement coverage is mandatory once coverage enforcement begins. 100% documentation, docstring, generated help, command, config, and examples coverage is mandatory with no exceptions. Behavior changes must update docs in the same change or prove no documentation impact by hash comparison and verification evidence.

## Zero-Hallucination And No-Placeholder Policy

Repository facts require local source, docs, test, or runtime evidence. Unsupported claims must be labeled `unverified`, `unsupported`, or `blocked`. Do not add placeholder-only files, TODO-only files, fake success paths, speculative APIs, silent unsupported branches, or vague future-work comments.

## Commit And Pull Request Guidelines

This checkout currently has no Git history. Use concise, imperative commit subjects such as `docs: add governance memory`. Pull requests must include changed behavior, docs updated, verification commands and results, coverage status, security notes, and known limits.

## Required Final Report

Every agent session ends with: Changed, Documented, Verified, Coverage, Security, and Known limits.

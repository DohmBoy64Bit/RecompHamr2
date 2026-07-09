# Performance Benchmarks

Phase 23 creates the first repeatable local performance baseline. These numbers
are evidence for this checkout on this machine only; they are not universal
release guarantees.

## Benchmark Command

Run the current baseline with:

```powershell
go test ./internal/llm ./internal/tui ./internal/mcp ./internal/tools ./internal/app -bench "Benchmark(PackLargeHistory|RenderWideLayout|RenderCompactLayout|ManagerToolsForSkills|Schemas|ReadFileSmall|ComposeRuntimeStartup)$" -benchmem -run "^$"
```

## Local Baseline

Environment:

- `goos`: `windows`
- `goarch`: `amd64`
- CPU: `AMD Ryzen 5 7600 6-Core Processor`

| Area | Benchmark | Result | Allocation evidence |
|---|---|---:|---:|
| Context packing | `BenchmarkPackLargeHistory-12` | `10528 ns/op` | `45826 B/op`, `2 allocs/op` |
| TUI wide render | `BenchmarkRenderWideLayout-12` | `1028 ns/op` | `1296 B/op`, `17 allocs/op` |
| TUI compact render | `BenchmarkRenderCompactLayout-12` | `13457 ns/op` | `21228 B/op`, `219 allocs/op` |
| MCP listing | `BenchmarkManagerToolsForSkills-12` | `22971 ns/op` | `59168 B/op`, `145 allocs/op` |
| Tool schema lookup | `BenchmarkSchemas-12` | `0.2438 ns/op` | `0 B/op`, `0 allocs/op` |
| Small file read tool | `BenchmarkReadFileSmall-12` | `49536 ns/op` | `9453 B/op`, `6 allocs/op` |
| Startup composition | `BenchmarkComposeRuntimeStartup-12` | `1606005 ns/op` | `58650 B/op`, `284 allocs/op` |

## Regression Policy

- Keep benchmark inputs deterministic and free of network, real model, and live
  terminal dependencies.
- Treat a repeated 2x slowdown in any benchmark as a regression requiring
  investigation before release candidate work.
- Do not claim published performance until the same benchmark command runs on
  the release CI matrix.
- If a benchmark result changes because behavior changed, update this document,
  the status report, and traceability in the same change.

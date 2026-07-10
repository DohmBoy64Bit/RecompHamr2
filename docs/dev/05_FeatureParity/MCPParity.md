# MCP Parity

MCP parity requires built-in registrations, JSON-RPC protocol support, stdio and streamable HTTP transports, manager lifecycle, skill-gated tools, `/mcp` command behavior, and documentation for every override.

Required built-ins are `ghidra`, `n64-debug-mcp`, `pcrecomp`, `mcp-pine`, `objdiff`, `pcsx2`, `bizhawk`, and `sega2asm`.

## Phase 10 Implemented

| Behavior | Status | Evidence |
|---|---|---|
| Built-in registration names | implemented | `internal/mcp.Builtins()` tests require 8 servers. |
| Environment command, URL, tool, and autostart metadata | implemented | `RECOMPHAMR_MCP_<NAME>_COMMAND`, `_URL`, `_TOOLS`, and `RECOMPHAMR_MCP_AUTOSTART` tests. |
| JSON-RPC 2.0 request/response/notification types | implemented | `protocol.go` constructors, validation, and decode tests. |
| MCP initialize/tools payload types | implemented | Initialize, tools/list, tools/call, schema map, and text result tests. |
| Stdio transport foundation | implemented | Fake stream round-trip, notification, mismatch, read/write, close, and cancellation tests. |
| Streamable HTTP request/response foundation | implemented | Fake HTTP server tests for `/mcp`, headers, notification, HTTP errors, malformed JSON, read failures, and ID mismatch. |

## Plan Phase 11 Implemented

| Behavior | Status | Evidence |
|---|---|---|
| Manager registration and status | implemented | `Manager`, `ServerStatus`, sorted status tests, and status formatting tests. |
| Connector-driven lifecycle | implemented | Fake connector success/failure tests and `UnsupportedConnector` tests. |
| Streamable HTTP connector | implemented | `HTTPConnector` fake server tests for initialize, initialized notification, tools/list, and tools/call. |
| Tool allowlist mutation | implemented | `SetToolEnabled` tests for single tools and `*` enable/disable. |
| Skill-gated tool visibility | implemented | `ToolsForSkills` tests for `ghidra-mcp` to `ghidra` mapping and ungated servers. |
| Full MCP tool calls | implemented | `CallTool` tests for `server.tool`, disabled tools, unknown servers, disconnected servers, and fake client calls. |
| `/mcp` lifecycle command dispatch | implemented | Command tests for manager-backed connect, disconnect, tools, enable, disable, and blocked failures. |

## Phase 29 Implemented

| Behavior | Status | Evidence |
|---|---|---|
| Live agent MCP tool exposure | implemented | `internal/app` tests expose connected, enabled, skill-visible `server.tool` schemas to prompt turns. |
| Live agent MCP tool dispatch | implemented | `internal/app` tests route `server.tool` calls through `Manager.CallTool` and return text output. |
| MCP tool failure formatting | implemented | `internal/app` tests convert MCP tool-level errors into tool-failure text. |
| Explicit autostart | implemented | `internal/app` tests connect only autostart configs and cover persistent config loading; `internal/mcp` tests cover HTTP and stdio connectors. |
| Stdio process spawning | implemented | `StdioConnector` helper-process tests cover spawn, initialize, tool list, tool call, startup failures, pipe failures, and initialization failures. |
| Persistent user MCP config | implemented | `LoadConfigFile`, `MergeConfigs`, and app runtime tests cover strict `.rehamr/mcp.json` parsing, unknown-field blockers, built-in overrides, custom servers, and false boolean overrides. |

## Verified Remaining MCP Limits

Real external MCP behavior depends on the user's configured local command or
HTTP endpoint. Stdio servers execute with the user's local permissions and must
be configured explicitly in `.rehamr/mcp.json` or built-in metadata. Tool
visibility remains limited to connected, enabled, skill-visible tools.

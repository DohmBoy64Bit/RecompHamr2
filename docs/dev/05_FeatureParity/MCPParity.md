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

## Verified Remaining MCP Limits

Stdio process spawning, app startup autoconnect, persistent user MCP config
files, and agent-loop MCP tool exposure remain `unsupported`. The supported
surface is the documented manager lifecycle, streamable HTTP connector,
skill-gated visibility, full `server.tool` calls, and manager-backed `/mcp`
command dispatch.

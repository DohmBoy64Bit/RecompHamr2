# MCP

MCP protocol and manager foundations are implemented. Bare startup wires an MCP
manager into the local command environment but does not autoconnect MCP servers.

Current implemented foundation:

- JSON-RPC 2.0 request, response, notification, and error payloads.
- MCP initialize, initialized notification, tools/list, and tools/call payloads.
- Stdio transport over an injected stream for tests and future process wiring.
- Streamable HTTP request/response posts to `<base_url>/mcp`.
- Manager state for registered servers, connect, disconnect, enabled tools,
  skill-gated visibility, status rendering, and `server.tool` calls.
- `/mcp connect`, `/mcp disconnect`, `/mcp tools`, `/mcp enable`, and
  `/mcp disable` when command execution receives an MCP manager.

Current runtime boundary:

- `/mcp` can list built-in registrations through the composed runtime manager.
- Streamable HTTP MCP endpoints can be initialized by manager wiring.
- Stdio MCP process spawning, app autostart, persistent user MCP config files,
  and agent-loop exposure remain unsupported.

All MCP servers must stay off by default unless skill-gated or explicitly
enabled. Server command and URL overrides require docs, examples, security
notes, and tests.

Environment override names:

- `RECOMPHAMR_MCP_<NAME>_COMMAND`
- `RECOMPHAMR_MCP_<NAME>_URL`
- `RECOMPHAMR_MCP_<NAME>_TOOLS`
- `RECOMPHAMR_MCP_AUTOSTART`

Examples:

```text
/mcp
/mcp connect ghidra
/mcp tools ghidra
/mcp disable ghidra decompile
/mcp enable ghidra decompile
/mcp disconnect ghidra
```

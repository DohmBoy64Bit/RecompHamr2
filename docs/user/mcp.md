# MCP

MCP protocol, manager, slash-command, persistent config, stdio process spawning,
and live agent-loop foundations are implemented. Bare startup wires an MCP
manager into the local command environment. Servers with explicit autostart
metadata may connect during runtime composition through either streamable HTTP
or stdio process transport.

Current implemented foundation:

- JSON-RPC 2.0 request, response, notification, and error payloads.
- MCP initialize, initialized notification, tools/list, and tools/call payloads.
- Stdio transport over spawned MCP server processes.
- Streamable HTTP request/response posts to `<base_url>/mcp`.
- Persistent `.rehamr/mcp.json` server overrides and custom server entries.
- Manager state for registered servers, connect, disconnect, enabled tools,
  skill-gated visibility, status rendering, and `server.tool` calls.
- `/mcp connect`, `/mcp disconnect`, `/mcp tools`, `/mcp enable`, and
  `/mcp disable` when command execution receives an MCP manager.
- Live agent turns expose connected, enabled MCP tools as `server.tool` function
  schemas only when the matching skill is active.
- MCP tool calls from the model dispatch through `internal/mcp.Manager.CallTool`
  and return text content to the agent loop.

Current runtime boundary:

- `/mcp` can list built-in registrations through the composed runtime manager.
- Streamable HTTP MCP endpoints can be initialized by manager wiring.
- `.rehamr/mcp.json` is strict JSON; unknown keys are blocked.
- `RECOMPHAMR_MCP_AUTOSTART=1` and persistent `autostart` metadata can
  autoconnect configured HTTP or stdio servers.

All MCP servers must stay off by default unless skill-gated or explicitly
enabled. Server command and URL overrides require docs, examples, security
notes, and tests.

Environment override names:

- `RECOMPHAMR_MCP_<NAME>_COMMAND`
- `RECOMPHAMR_MCP_<NAME>_URL`
- `RECOMPHAMR_MCP_<NAME>_TOOLS`
- `RECOMPHAMR_MCP_AUTOSTART`

Persistent config example:

```json
{
  "servers": {
    "ghidra": {
      "command": "ghidra-mcp",
      "args": ["--stdio"],
      "allowed_tools": ["decompile"],
      "autostart": false,
      "require_skill": true
    },
    "local-debug": {
      "url": "http://127.0.0.1:8765",
      "allowed_tools": ["ping"],
      "autostart": true,
      "require_skill": false
    }
  }
}
```

Create or edit `.rehamr/mcp.json`, then run:

```text
recomphamr --summary
/mcp
```

The summary reports the composed server count. `/mcp` reports connection state
and any blocked autostart errors.

Examples:

```text
/mcp
/mcp connect ghidra
/mcp tools ghidra
/mcp disable ghidra decompile
/mcp enable ghidra decompile
/mcp disconnect ghidra
```

To make a connected Ghidra MCP tool visible to model turns, activate a matching
skill first:

```text
/skill ghidra-mcp
/mcp connect ghidra
```

The model then sees enabled tools such as `ghidra.decompile` during prompt
turns. Disabled tools and tools behind inactive skills are not exposed.

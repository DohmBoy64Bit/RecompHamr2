# MCP Architecture

MCP servers are disabled unless skill-gated or explicitly enabled. Command and URL overrides require documentation, tests, and security review.

## Phase 10 Protocol Foundation

`internal/mcp` owns MCP JSON-RPC 2.0 protocol shapes and transport foundations:

- `Request`, `Response`, `RPCError`, and `Notification` represent JSON-RPC messages.
- `InitializeParams`, `InitializeResult`, `ListToolsResult`, `ToolDef`, `InputSchema`, `CallToolParams`, and `CallToolResult` represent MCP initialize, tool listing, and tool call payloads.
- `StdioTransport` sends line-delimited JSON over an injected `io.ReadWriteCloser`; it does not spawn processes.
- `HTTPTransport` posts JSON-RPC to `<base_url>/mcp` with `Content-Type: application/json` and `Accept: application/json, text/event-stream`.

Context cancellation is mandatory. Stdio cancellation closes the injected stream to unblock reads. HTTP cancellation uses `http.NewRequestWithContext`.

## Boundaries

Phase 10 does not connect real built-in servers, mutate `/mcp` lifecycle state, or expose MCP tools to the agent loop. Manager lifecycle, process spawning, user config merging, tool enable/disable, and `/mcp` runtime mutation belong to the next MCP manager phase.

Unsupported or unverified behavior must remain explicit in command output and docs until those manager paths are implemented and tested.

## Plan Phase 11 Manager

`internal/mcp.Manager` owns registered server state, connector-driven lifecycle,
enabled-tool allowlists, skill-gated tool visibility, status rendering, and
`server.tool` calls. It uses an injected `Connector`, so tests and app wiring can
choose fake clients, HTTP clients, or later stdio process clients without
changing manager state rules.

`HTTPConnector` is implemented for streamable HTTP endpoints. It initializes a
`ProtocolClient`, sends `notifications/initialized`, lists tools, and then uses
the protocol transport for `tools/call`.

Stdio process spawning remains explicitly unsupported in this phase. The
manager can hold stdio command metadata, but no code starts local MCP server
processes yet.

`/mcp` command behavior:

- without manager wiring: built-in listing remains available and lifecycle
  mutation returns `unsupported:`.
- with manager wiring: `connect`, `disconnect`, `tools`, `enable`, and `disable`
  call the manager and return `blocked:` on manager errors.

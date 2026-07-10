# MCP Architecture

MCP servers are disabled unless skill-gated or explicitly enabled. Command and URL overrides require documentation, tests, and security review.

## Phase 10 Protocol Foundation

`internal/mcp` owns MCP JSON-RPC 2.0 protocol shapes and transport foundations:

- `Request`, `Response`, `RPCError`, and `Notification` represent JSON-RPC messages.
- `InitializeParams`, `InitializeResult`, `ListToolsResult`, `ToolDef`, `InputSchema`, `CallToolParams`, and `CallToolResult` represent MCP initialize, tool listing, and tool call payloads.
- `StdioTransport` sends line-delimited JSON over an injected `io.ReadWriteCloser`.
- `HTTPTransport` posts JSON-RPC to `<base_url>/mcp` with `Content-Type: application/json` and `Accept: application/json, text/event-stream`.
- `StdioConnector` spawns configured stdio MCP commands and binds their stdin
  and stdout to the protocol client.
- `AutoConnector` chooses HTTP when `url` is set and stdio otherwise.

Context cancellation is mandatory. Stdio cancellation closes the injected stream to unblock reads. HTTP cancellation uses `http.NewRequestWithContext`.

## Boundaries

Phase 10 did not connect real built-in servers, mutate `/mcp` lifecycle state,
or expose MCP tools to the agent loop. Later MCP manager and corrective runtime
phases now own lifecycle mutation, stdio process spawning, persistent user config
merging, tool enable/disable, and live agent exposure.

Unsupported or unverified behavior must remain explicit in command output and docs until those manager paths are implemented and tested.

## Plan Phase 11 Manager

`internal/mcp.Manager` owns registered server state, connector-driven lifecycle,
enabled-tool allowlists, skill-gated tool visibility, status rendering, and
`server.tool` calls. It uses an injected `Connector`, so tests and app wiring can
choose fake clients, HTTP clients, or later stdio process clients without
changing manager state rules.

`HTTPConnector` is implemented for streamable HTTP endpoints. `StdioConnector`
is implemented for configured local process commands. Both initialize a
`ProtocolClient`, send `notifications/initialized`, list tools, and then use the
protocol transport for `tools/call`.

`/mcp` command behavior:

- without manager wiring: built-in listing remains available and lifecycle
  mutation returns `unsupported:`.
- with manager wiring: `connect`, `disconnect`, `tools`, `enable`, and `disable`
  call the manager and return `blocked:` on manager errors.

## Phase 29 Live Agent Integration

`internal/app` exposes MCP tools to the model only through
`internal/mcp.Manager.ToolsForSkills`. A tool is visible during a live prompt
turn only when all of these are true:

- the MCP server is connected,
- the tool is enabled by the manager allowlist,
- the server's skill gate is satisfied by the active TUI skill list.

Live MCP tool names use the `server.tool` form already accepted by
`Manager.CallTool`. `internal/app` dispatches those calls to the manager and
returns text content to the agent loop. MCP tool-level errors are formatted as
tool failures so the agent loop can nudge repeated failures instead of treating
them as success.

Autostart remains explicit and bounded. During runtime composition, app wiring
attempts autoconnect only for configs with `Autostart=true`. HTTP configs use
their URL. Stdio configs spawn only the configured command and arguments.

Persistent user config lives at `.rehamr/mcp.json`. The schema is strict JSON:
`servers`, server `name`, `command`, `args`, `url`, `allowed_tools`,
`autostart`, and `require_skill`. Server names must match their map key when a
`name` field is present. Unknown fields block runtime composition instead of
being ignored.

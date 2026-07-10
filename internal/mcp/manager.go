package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var skillServerMap = map[string]string{
	"ghidra-mcp": "ghidra",
	"n64-decomp": "n64-debug-mcp",
	"pcrecomp":   "pcrecomp",
	"mcp-pine":   "mcp-pine",
	"objdiff":    "objdiff",
	"pcsx2":      "pcsx2",
	"bizhawk":    "bizhawk",
	"sega2asm":   "sega2asm",
}

// ServerState describes the manager lifecycle state for one MCP server.
type ServerState string

const (
	// StateDisconnected means no client is attached.
	StateDisconnected ServerState = "disconnected"
	// StateConnecting means a connection attempt is in progress.
	StateConnecting ServerState = "connecting"
	// StateConnected means a client is available for tool calls.
	StateConnected ServerState = "connected"
	// StateError means the most recent lifecycle action failed.
	StateError ServerState = "error"
)

// ServerStatus describes one MCP server for status output.
type ServerStatus struct {
	// Name is the server name.
	Name string
	// State is the lifecycle state.
	State ServerState
	// Tools is the count of currently enabled tools.
	Tools int
	// Err is the latest lifecycle error, if any.
	Err string
}

// Client is an established MCP client session.
type Client interface {
	// Tools returns the server-local tool definitions.
	Tools() []ToolDef
	// CallTool invokes a server-local tool.
	CallTool(context.Context, string, map[string]interface{}) (CallToolResult, error)
	// Close releases client resources.
	Close() error
}

// Connector establishes an MCP client for a server config.
type Connector interface {
	// Connect creates a client session for config.
	Connect(context.Context, ServerConfig) (Client, error)
}

// ConnectorFunc adapts a function to Connector.
type ConnectorFunc func(context.Context, ServerConfig) (Client, error)

// Connect creates a client session for config.
func (fn ConnectorFunc) Connect(ctx context.Context, config ServerConfig) (Client, error) {
	return fn(ctx, config)
}

// Manager owns MCP server state, enabled tools, and skill-gated visibility.
type Manager struct {
	mu        sync.Mutex
	connector Connector
	entries   map[string]*managerEntry
}

type managerEntry struct {
	config  ServerConfig
	client  Client
	state   ServerState
	err     string
	enabled map[string]bool
}

// NewManager creates a manager with registered configs.
func NewManager(configs []ServerConfig, connector Connector) *Manager {
	if connector == nil {
		connector = UnsupportedConnector{}
	}
	manager := &Manager{connector: connector, entries: map[string]*managerEntry{}}
	for _, config := range configs {
		manager.Register(config)
	}
	return manager
}

// Register adds or replaces one server configuration.
func (m *Manager) Register(config ServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[config.Name] = &managerEntry{config: config, state: StateDisconnected, enabled: allowlist(config.AllowedTools)}
}

// Connect establishes a client for name.
func (m *Manager) Connect(ctx context.Context, name string) error {
	m.mu.Lock()
	entry, ok := m.entries[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("mcp: unknown server %q", name)
	}
	if entry.state == StateConnected {
		m.mu.Unlock()
		return nil
	}
	entry.state = StateConnecting
	entry.err = ""
	config := entry.config
	m.mu.Unlock()

	client, err := m.connector.Connect(ctx, config)

	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		entry.state = StateError
		entry.err = err.Error()
		entry.client = nil
		return err
	}
	entry.client = client
	entry.state = StateConnected
	entry.err = ""
	if entry.enabled == nil && len(config.AllowedTools) > 0 {
		entry.enabled = allowlist(config.AllowedTools)
	}
	return nil
}

// Disconnect closes the client for name.
func (m *Manager) Disconnect(name string) error {
	m.mu.Lock()
	entry, ok := m.entries[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("mcp: unknown server %q", name)
	}
	client := entry.client
	entry.client = nil
	entry.state = StateDisconnected
	entry.err = ""
	m.mu.Unlock()
	if client == nil {
		return nil
	}
	if err := client.Close(); err != nil {
		m.mu.Lock()
		entry.state = StateError
		entry.err = err.Error()
		m.mu.Unlock()
		return err
	}
	return nil
}

// Status returns the current status for name.
func (m *Manager) Status(name string) ServerStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[name]
	if !ok {
		return ServerStatus{Name: name, State: StateDisconnected}
	}
	return ServerStatus{Name: entry.config.Name, State: entry.state, Tools: len(entry.tools()), Err: entry.err}
}

// AllStatus returns statuses sorted by server name.
func (m *Manager) AllStatus() []ServerStatus {
	m.mu.Lock()
	names := make([]string, 0, len(m.entries))
	for name := range m.entries {
		names = append(names, name)
	}
	m.mu.Unlock()
	sort.Strings(names)
	statuses := make([]ServerStatus, 0, len(names))
	for _, name := range names {
		statuses = append(statuses, m.Status(name))
	}
	return statuses
}

// Tools returns enabled tools for one connected server.
func (m *Manager) Tools(name string) ([]ToolDef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[name]
	if !ok {
		return nil, fmt.Errorf("mcp: unknown server %q", name)
	}
	if entry.state != StateConnected || entry.client == nil {
		return nil, fmt.Errorf("mcp: server %q is not connected", name)
	}
	return entry.tools(), nil
}

// ToolsForSkills returns visible tools for active skills.
func (m *Manager) ToolsForSkills(activeSkills []string) []ToolDef {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []ToolDef
	for _, entry := range m.entries {
		if entry.state != StateConnected || entry.client == nil {
			continue
		}
		if entry.config.RequireSkill && !skillsAllowServer(activeSkills, entry.config.Name) {
			continue
		}
		for _, tool := range entry.tools() {
			tool.Name = entry.config.Name + "." + tool.Name
			out = append(out, tool)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SetToolEnabled enables or disables one tool for a server.
func (m *Manager) SetToolEnabled(serverName string, toolName string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[serverName]
	if !ok {
		return fmt.Errorf("mcp: unknown server %q", serverName)
	}
	if toolName == "*" {
		if enabled {
			entry.enabled = nil
		} else {
			entry.enabled = map[string]bool{}
		}
		return nil
	}
	if enabled {
		if entry.enabled == nil {
			entry.enabled = map[string]bool{}
			for _, tool := range entry.rawTools() {
				entry.enabled[tool.Name] = true
			}
		}
		entry.enabled[toolName] = true
		return nil
	}
	if entry.enabled == nil {
		entry.enabled = map[string]bool{}
		for _, tool := range entry.rawTools() {
			entry.enabled[tool.Name] = true
		}
	}
	delete(entry.enabled, toolName)
	return nil
}

// CallTool invokes a full server.tool name.
func (m *Manager) CallTool(ctx context.Context, fullName string, args map[string]interface{}) (CallToolResult, error) {
	serverName, toolName, ok := strings.Cut(fullName, ".")
	if !ok || serverName == "" || toolName == "" {
		return CallToolResult{}, fmt.Errorf("mcp: invalid tool name %q", fullName)
	}
	m.mu.Lock()
	entry, ok := m.entries[serverName]
	if !ok {
		m.mu.Unlock()
		return CallToolResult{}, fmt.Errorf("mcp: unknown server %q", serverName)
	}
	if entry.state != StateConnected || entry.client == nil {
		m.mu.Unlock()
		return CallToolResult{}, fmt.Errorf("mcp: server %q is not connected", serverName)
	}
	if !entry.toolEnabled(toolName) {
		m.mu.Unlock()
		return CallToolResult{}, fmt.Errorf("mcp: tool %q is disabled", fullName)
	}
	client := entry.client
	m.mu.Unlock()
	return client.CallTool(ctx, toolName, args)
}

// FormatStatus renders a compact status table.
func (m *Manager) FormatStatus() string {
	var b strings.Builder
	b.WriteString("mcp servers:")
	for _, status := range m.AllStatus() {
		line := fmt.Sprintf("\n%s %s tools:%d", status.Name, status.State, status.Tools)
		if status.Err != "" {
			line += " error:" + status.Err
		}
		b.WriteString(line)
	}
	return b.String()
}

// FormatTools renders tools for one server.
func (m *Manager) FormatTools(name string) string {
	tools, err := m.Tools(name)
	if err != nil {
		return "blocked: " + err.Error()
	}
	if len(tools) == 0 {
		return "mcp " + name + ": no enabled tools"
	}
	parts := make([]string, 0, len(tools))
	for _, tool := range tools {
		parts = append(parts, tool.Name)
	}
	sort.Strings(parts)
	return "mcp " + name + " tools: " + strings.Join(parts, ", ")
}

func (e *managerEntry) rawTools() []ToolDef {
	if e.client == nil {
		return nil
	}
	tools := e.client.Tools()
	return append([]ToolDef(nil), tools...)
}

func (e *managerEntry) tools() []ToolDef {
	raw := e.rawTools()
	out := make([]ToolDef, 0, len(raw))
	for _, tool := range raw {
		if e.toolEnabled(tool.Name) {
			out = append(out, tool)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (e *managerEntry) toolEnabled(name string) bool {
	return e.enabled == nil || e.enabled[name]
}

func allowlist(tools []string) map[string]bool {
	if len(tools) == 0 {
		return nil
	}
	out := make(map[string]bool, len(tools))
	for _, tool := range tools {
		tool = strings.TrimSpace(tool)
		if tool != "" {
			out[tool] = true
		}
	}
	return out
}

func skillsAllowServer(activeSkills []string, server string) bool {
	for _, skill := range activeSkills {
		if SkillAllowsServer(skill, server) {
			return true
		}
	}
	return false
}

// UnsupportedConnector reports unsupported lifecycle behavior without pretending success.
type UnsupportedConnector struct{}

// Connect always returns an unsupported lifecycle error.
func (UnsupportedConnector) Connect(context.Context, ServerConfig) (Client, error) {
	return nil, fmt.Errorf("unsupported: MCP connector is not configured")
}

// HTTPConnector creates MCP clients for streamable HTTP servers.
type HTTPConnector struct {
	// HTTPClient sends HTTP requests; nil uses http.DefaultClient.
	HTTPClient *http.Client
	// ClientName is reported during initialize.
	ClientName string
	// ClientVersion is reported during initialize.
	ClientVersion string
}

// StdioConnector creates MCP clients by spawning stdio MCP server processes.
type StdioConnector struct {
	// ClientName is reported during initialize.
	ClientName string
	// ClientVersion is reported during initialize.
	ClientVersion string
}

// AutoConnector chooses HTTP or stdio MCP connection based on server config.
type AutoConnector struct {
	// HTTPClient sends HTTP requests for URL-backed configs.
	HTTPClient *http.Client
	// ClientName is reported during initialize.
	ClientName string
	// ClientVersion is reported during initialize.
	ClientVersion string
}

// Connect initializes an HTTP MCP client.
func (c HTTPConnector) Connect(ctx context.Context, config ServerConfig) (Client, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("unsupported: stdio MCP process spawning is not wired")
	}
	clientName := c.ClientName
	if clientName == "" {
		clientName = "recomphamr"
	}
	clientVersion := c.ClientVersion
	if clientVersion == "" {
		clientVersion = "2.0"
	}
	session := NewProtocolClient(config.Name, clientName, clientVersion, NewHTTPTransport(config.Name, config.URL, c.HTTPClient))
	if err := session.Initialize(ctx); err != nil {
		return nil, err
	}
	return session, nil
}

// Connect starts a stdio MCP process and initializes a protocol client.
func (c StdioConnector) Connect(ctx context.Context, config ServerConfig) (Client, error) {
	command := strings.TrimSpace(config.Command)
	if command == "" {
		return nil, fmt.Errorf("mcp: command is empty for %s", config.Name)
	}
	clientName := c.ClientName
	if clientName == "" {
		clientName = "recomphamr"
	}
	clientVersion := c.ClientVersion
	if clientVersion == "" {
		clientVersion = "2.0"
	}
	stream, err := startStdioProcess(ctx, command, config.Args)
	if err != nil {
		return nil, err
	}
	session := NewProtocolClient(config.Name, clientName, clientVersion, NewStdioTransport(config.Name, stream))
	if err := session.Initialize(ctx); err != nil {
		_ = session.Close()
		return nil, err
	}
	return session, nil
}

// Connect chooses streamable HTTP when URL is present, otherwise stdio command spawning.
func (c AutoConnector) Connect(ctx context.Context, config ServerConfig) (Client, error) {
	if strings.TrimSpace(config.URL) != "" {
		return HTTPConnector{HTTPClient: c.HTTPClient, ClientName: c.ClientName, ClientVersion: c.ClientVersion}.Connect(ctx, config)
	}
	return StdioConnector{ClientName: c.ClientName, ClientVersion: c.ClientVersion}.Connect(ctx, config)
}

// ProtocolClient is an MCP client backed by a Transport.
type ProtocolClient struct {
	name          string
	clientName    string
	clientVersion string
	transport     Transport
	nextID        int64
	tools         []ToolDef
}

// NewProtocolClient creates a transport-backed MCP client.
func NewProtocolClient(name string, clientName string, clientVersion string, transport Transport) *ProtocolClient {
	return &ProtocolClient{name: name, clientName: clientName, clientVersion: clientVersion, transport: transport}
}

// Initialize performs initialize, initialized notification, and tools/list.
func (c *ProtocolClient) Initialize(ctx context.Context) error {
	initReq, _ := InitializeRequest(c.next(), c.clientName, c.clientVersion)
	initResp, err := c.transport.RoundTrip(ctx, initReq)
	if err != nil {
		return err
	}
	var initResult InitializeResult
	if err := initResp.ResultAs(&initResult); err != nil {
		return err
	}
	notification, _ := NewNotification(MethodInitialized, nil)
	if err := c.transport.Notify(ctx, notification); err != nil {
		return err
	}
	toolsReq, _ := ToolsListRequest(c.next())
	toolsResp, err := c.transport.RoundTrip(ctx, toolsReq)
	if err != nil {
		return err
	}
	var list ListToolsResult
	if err := toolsResp.ResultAs(&list); err != nil {
		return err
	}
	c.tools = append([]ToolDef(nil), list.Tools...)
	return nil
}

// Tools returns the initialized tools.
func (c *ProtocolClient) Tools() []ToolDef {
	return append([]ToolDef(nil), c.tools...)
}

// CallTool calls a server-local tool.
func (c *ProtocolClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (CallToolResult, error) {
	req, err := ToolsCallRequest(c.next(), name, args)
	if err != nil {
		return CallToolResult{}, err
	}
	resp, err := c.transport.RoundTrip(ctx, req)
	if err != nil {
		return CallToolResult{}, err
	}
	var result CallToolResult
	if err := resp.ResultAs(&result); err != nil {
		return CallToolResult{}, err
	}
	return result, nil
}

// Close releases the client transport.
func (c *ProtocolClient) Close() error {
	return c.transport.Close()
}

func (c *ProtocolClient) next() int64 {
	return atomic.AddInt64(&c.nextID, 1)
}

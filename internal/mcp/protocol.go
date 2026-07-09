package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONRPCVersion is the JSON-RPC protocol version used by MCP.
const JSONRPCVersion = "2.0"

// ProtocolVersion is the MCP protocol version requested by RecompHamr.
const ProtocolVersion = "2024-11-05"

const (
	// MethodInitialize starts an MCP session.
	MethodInitialize = "initialize"
	// MethodInitialized confirms the client completed initialization.
	MethodInitialized = "notifications/initialized"
	// MethodToolsList lists server tools.
	MethodToolsList = "tools/list"
	// MethodToolsCall invokes one server tool.
	MethodToolsCall = "tools/call"
)

// Request is a JSON-RPC request sent to an MCP server.
type Request struct {
	// JSONRPC is always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// ID correlates the request with a response.
	ID int64 `json:"id"`
	// Method is the MCP method name.
	Method string `json:"method"`
	// Params is the optional method payload.
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC response returned by an MCP server.
type Response struct {
	// JSONRPC is always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// ID matches the request ID.
	ID int64 `json:"id"`
	// Result is the successful response payload.
	Result json.RawMessage `json:"result,omitempty"`
	// Error is the JSON-RPC error payload.
	Error *RPCError `json:"error,omitempty"`
}

// RPCError describes a JSON-RPC error response.
type RPCError struct {
	// Code is the JSON-RPC or server-defined error code.
	Code int `json:"code"`
	// Message is the user-visible error text.
	Message string `json:"message"`
}

// Error formats an RPCError for Go error returns.
func (e RPCError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// Notification is a JSON-RPC notification with no response ID.
type Notification struct {
	// JSONRPC is always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// Method is the MCP notification method.
	Method string `json:"method"`
	// Params is the optional notification payload.
	Params json.RawMessage `json:"params,omitempty"`
}

// InitializeParams is sent with the MCP initialize request.
type InitializeParams struct {
	// ProtocolVersion is the requested MCP protocol version.
	ProtocolVersion string `json:"protocolVersion"`
	// Capabilities describes RecompHamr client capabilities.
	Capabilities ClientCapabilities `json:"capabilities"`
	// ClientInfo identifies RecompHamr to the server.
	ClientInfo ClientInfo `json:"clientInfo"`
}

// ClientCapabilities describes MCP client capabilities.
type ClientCapabilities struct{}

// ClientInfo identifies the MCP client.
type ClientInfo struct {
	// Name is the client name.
	Name string `json:"name"`
	// Version is the client version.
	Version string `json:"version"`
}

// InitializeResult is returned by a successful initialize request.
type InitializeResult struct {
	// ProtocolVersion is the server-selected MCP protocol version.
	ProtocolVersion string `json:"protocolVersion"`
	// Capabilities describes server capabilities.
	Capabilities ServerCapabilities `json:"capabilities"`
	// ServerInfo identifies the MCP server.
	ServerInfo ServerInfo `json:"serverInfo"`
}

// ServerCapabilities describes MCP server capabilities.
type ServerCapabilities struct {
	// Tools is present when the server supports tools.
	Tools *struct{} `json:"tools,omitempty"`
}

// ServerInfo identifies an MCP server.
type ServerInfo struct {
	// Name is the server name.
	Name string `json:"name"`
	// Version is the server version.
	Version string `json:"version"`
}

// ListToolsResult is returned by tools/list.
type ListToolsResult struct {
	// Tools is the list of tool definitions.
	Tools []ToolDef `json:"tools"`
}

// ToolDef describes one MCP tool.
type ToolDef struct {
	// Name is the server-local tool name.
	Name string `json:"name"`
	// Description explains the tool behavior.
	Description string `json:"description"`
	// InputSchema documents accepted tool input.
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema is the JSON schema object for an MCP tool.
type InputSchema struct {
	// Type is the schema type, normally "object".
	Type string `json:"type"`
	// Properties documents schema fields.
	Properties map[string]interface{} `json:"properties,omitempty"`
	// Required lists required field names.
	Required []string `json:"required,omitempty"`
}

// Map returns a complete map form suitable for LLM tool schemas.
func (s InputSchema) Map() map[string]interface{} {
	m := map[string]interface{}{"type": s.Type, "properties": map[string]interface{}{}, "required": []string{}}
	if s.Properties != nil {
		m["properties"] = s.Properties
	}
	if s.Required != nil {
		m["required"] = s.Required
	}
	return m
}

// CallToolParams is sent with tools/call.
type CallToolParams struct {
	// Name is the server-local tool name.
	Name string `json:"name"`
	// Arguments is the JSON object passed to the tool.
	Arguments map[string]interface{} `json:"arguments"`
}

// CallToolResult is returned by tools/call.
type CallToolResult struct {
	// Content is the ordered tool output content.
	Content []ContentItem `json:"content"`
	// IsError is true when the tool result represents a tool-level failure.
	IsError bool `json:"isError,omitempty"`
}

// Text returns the first text content item.
func (r CallToolResult) Text() string {
	for _, content := range r.Content {
		if content.Type == "text" {
			return content.Text
		}
	}
	return ""
}

// ContentItem is one MCP tool result content item.
type ContentItem struct {
	// Type is the content kind, such as "text".
	Type string `json:"type"`
	// Text is present for text content.
	Text string `json:"text"`
}

// NewRequest builds a validated JSON-RPC request.
func NewRequest(id int64, method string, params interface{}) (Request, error) {
	method = strings.TrimSpace(method)
	if id <= 0 {
		return Request{}, fmt.Errorf("mcp request id must be positive")
	}
	if method == "" {
		return Request{}, fmt.Errorf("mcp request method is empty")
	}
	raw, err := marshalPayload(params)
	if err != nil {
		return Request{}, err
	}
	return Request{JSONRPC: JSONRPCVersion, ID: id, Method: method, Params: raw}, nil
}

// NewNotification builds a validated JSON-RPC notification.
func NewNotification(method string, params interface{}) (Notification, error) {
	method = strings.TrimSpace(method)
	if method == "" {
		return Notification{}, fmt.Errorf("mcp notification method is empty")
	}
	raw, err := marshalPayload(params)
	if err != nil {
		return Notification{}, err
	}
	return Notification{JSONRPC: JSONRPCVersion, Method: method, Params: raw}, nil
}

// NewResponse builds a successful JSON-RPC response.
func NewResponse(id int64, result interface{}) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("mcp response id must be positive")
	}
	raw, err := marshalPayload(result)
	if err != nil {
		return Response{}, err
	}
	if len(raw) == 0 {
		return Response{}, fmt.Errorf("mcp response result is empty")
	}
	return Response{JSONRPC: JSONRPCVersion, ID: id, Result: raw}, nil
}

// NewErrorResponse builds a JSON-RPC error response.
func NewErrorResponse(id int64, code int, message string) Response {
	return Response{JSONRPC: JSONRPCVersion, ID: id, Error: &RPCError{Code: code, Message: message}}
}

// ResultAs decodes a successful response result into target.
func (r Response) ResultAs(target interface{}) error {
	if r.Error != nil {
		return *r.Error
	}
	if len(r.Result) == 0 {
		return fmt.Errorf("mcp response %d has no result", r.ID)
	}
	return json.Unmarshal(r.Result, target)
}

// DecodeResponse parses and validates a JSON-RPC response.
func DecodeResponse(data []byte) (Response, error) {
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return Response{}, err
	}
	if response.JSONRPC != JSONRPCVersion {
		return Response{}, fmt.Errorf("mcp response has unsupported jsonrpc %q", response.JSONRPC)
	}
	if response.ID <= 0 {
		return Response{}, fmt.Errorf("mcp response id must be positive")
	}
	if len(response.Result) == 0 && response.Error == nil {
		return Response{}, fmt.Errorf("mcp response has neither result nor error")
	}
	return response, nil
}

// InitializeRequest builds the standard initialize request.
func InitializeRequest(id int64, clientName string, clientVersion string) (Request, error) {
	return NewRequest(id, MethodInitialize, InitializeParams{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    ClientCapabilities{},
		ClientInfo:      ClientInfo{Name: clientName, Version: clientVersion},
	})
}

// ToolsListRequest builds the standard tools/list request.
func ToolsListRequest(id int64) (Request, error) {
	return NewRequest(id, MethodToolsList, nil)
}

// ToolsCallRequest builds the standard tools/call request.
func ToolsCallRequest(id int64, name string, args map[string]interface{}) (Request, error) {
	return NewRequest(id, MethodToolsCall, CallToolParams{Name: name, Arguments: args})
}

func marshalPayload(payload interface{}) (json.RawMessage, error) {
	if payload == nil {
		return nil, nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

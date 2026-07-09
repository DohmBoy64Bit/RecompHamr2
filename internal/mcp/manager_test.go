package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestManagerLifecycleAndTools(t *testing.T) {
	tool := ToolDef{Name: "decompile", Description: "Decompile a function.", InputSchema: InputSchema{Type: "object"}}
	client := &fakeMCPClient{tools: []ToolDef{tool}}
	manager := NewManager([]ServerConfig{{Name: "ghidra", RequireSkill: true}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return client, nil
	}))
	if got := manager.Status("missing"); got.State != StateDisconnected || got.Name != "missing" {
		t.Fatalf("Status(missing) = %+v", got)
	}
	if err := manager.Connect(context.Background(), "missing"); err == nil {
		t.Fatal("Connect() accepted unknown server")
	}
	if err := manager.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := manager.Connect(context.Background(), "ghidra"); err != nil {
		t.Fatalf("Connect() already connected error = %v", err)
	}
	status := manager.Status("ghidra")
	if status.State != StateConnected || status.Tools != 1 || status.Err != "" {
		t.Fatalf("Status(ghidra) = %+v", status)
	}
	all := manager.AllStatus()
	if len(all) != 1 || all[0].Name != "ghidra" {
		t.Fatalf("AllStatus() = %+v", all)
	}
	tools, err := manager.Tools("ghidra")
	if err != nil || len(tools) != 1 || tools[0].Name != "decompile" {
		t.Fatalf("Tools() = %+v, %v", tools, err)
	}
	tools[0].Name = "mutated"
	again, _ := manager.Tools("ghidra")
	if again[0].Name != "decompile" {
		t.Fatal("Tools() returned shared slice")
	}
	if visible := manager.ToolsForSkills([]string{"ghidra-mcp"}); len(visible) != 1 || visible[0].Name != "ghidra.decompile" {
		t.Fatalf("ToolsForSkills() = %+v", visible)
	}
	if visible := manager.ToolsForSkills([]string{"other"}); len(visible) != 0 {
		t.Fatalf("ToolsForSkills(other) = %+v", visible)
	}
	if err := manager.SetToolEnabled("ghidra", "decompile", false); err != nil {
		t.Fatalf("SetToolEnabled(false) error = %v", err)
	}
	if got := manager.FormatTools("ghidra"); !strings.Contains(got, "no enabled tools") {
		t.Fatalf("FormatTools(disabled) = %q", got)
	}
	if _, err := manager.CallTool(context.Background(), "ghidra.decompile", nil); err == nil {
		t.Fatal("CallTool() accepted disabled tool")
	}
	if err := manager.SetToolEnabled("ghidra", "decompile", true); err != nil {
		t.Fatalf("SetToolEnabled(true) error = %v", err)
	}
	result, err := manager.CallTool(context.Background(), "ghidra.decompile", map[string]interface{}{"addr": "0x1000"})
	if err != nil || result.Text() != "called decompile" || client.called != "decompile" {
		t.Fatalf("CallTool() = %+v, %v client=%q", result, err, client.called)
	}
	if got := manager.FormatStatus(); !strings.Contains(got, "ghidra connected tools:1") {
		t.Fatalf("FormatStatus() = %q", got)
	}
	if got := manager.FormatTools("ghidra"); !strings.Contains(got, "decompile") {
		t.Fatalf("FormatTools() = %q", got)
	}
	if err := manager.Disconnect("ghidra"); err != nil || !client.closed {
		t.Fatalf("Disconnect() = %v closed=%v", err, client.closed)
	}
	if err := manager.Disconnect("ghidra"); err != nil {
		t.Fatalf("Disconnect() second error = %v", err)
	}
	if err := manager.Disconnect("missing"); err == nil {
		t.Fatal("Disconnect() accepted unknown server")
	}
}

func TestManagerErrorsAndMutation(t *testing.T) {
	manager := NewManager(nil, nil)
	manager.Register(ServerConfig{Name: "x"})
	if err := manager.Connect(context.Background(), "x"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("Connect() with unsupported connector = %v", err)
	}
	if status := manager.Status("x"); status.State != StateError || status.Err == "" {
		t.Fatalf("Status(error) = %+v", status)
	}
	if _, err := manager.Tools("x"); err == nil {
		t.Fatal("Tools() accepted disconnected server")
	}
	if _, err := manager.Tools("missing"); err == nil {
		t.Fatal("Tools() accepted unknown server")
	}
	if visible := manager.ToolsForSkills([]string{"x"}); len(visible) != 0 {
		t.Fatalf("ToolsForSkills(disconnected) = %+v", visible)
	}
	if got := manager.FormatTools("x"); !strings.Contains(got, "blocked:") {
		t.Fatalf("FormatTools(disconnected) = %q", got)
	}
	if got := manager.FormatStatus(); !strings.Contains(got, "error:") {
		t.Fatalf("FormatStatus(error) = %q", got)
	}
	if err := manager.SetToolEnabled("missing", "x", true); err == nil {
		t.Fatal("SetToolEnabled() accepted unknown server")
	}
	if _, err := manager.CallTool(context.Background(), "badname", nil); err == nil {
		t.Fatal("CallTool() accepted bad full name")
	}
	if _, err := manager.CallTool(context.Background(), "missing.tool", nil); err == nil {
		t.Fatal("CallTool() accepted unknown server")
	}
	if _, err := manager.CallTool(context.Background(), "x.tool", nil); err == nil {
		t.Fatal("CallTool() accepted disconnected server")
	}

	closeClient := &fakeMCPClient{closeErr: errors.New("close failed")}
	closeManager := NewManager([]ServerConfig{{Name: "x"}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return closeClient, nil
	}))
	if err := closeManager.Connect(context.Background(), "x"); err != nil {
		t.Fatalf("Connect(close manager) error = %v", err)
	}
	if err := closeManager.Disconnect("x"); err == nil {
		t.Fatal("Disconnect() accepted close failure")
	}
	if status := closeManager.Status("x"); status.State != StateError || status.Err == "" {
		t.Fatalf("Status(close error) = %+v", status)
	}

	failManager := NewManager([]ServerConfig{{Name: "x"}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return nil, errors.New("connect failed")
	}))
	if err := failManager.Connect(context.Background(), "x"); err == nil {
		t.Fatal("Connect() accepted connector failure")
	}
}

func TestManagerAllowlistAndStarMutation(t *testing.T) {
	client := &fakeMCPClient{tools: []ToolDef{{Name: "a"}, {Name: "b"}}}
	manager := NewManager([]ServerConfig{{Name: "open", AllowedTools: []string{"b"}}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return client, nil
	}))
	if err := manager.Connect(context.Background(), "open"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if tools, _ := manager.Tools("open"); len(tools) != 1 || tools[0].Name != "b" {
		t.Fatalf("Tools(allowlist) = %+v", tools)
	}
	if err := manager.SetToolEnabled("open", "*", true); err != nil {
		t.Fatalf("SetToolEnabled(* true) error = %v", err)
	}
	if tools, _ := manager.Tools("open"); len(tools) != 2 {
		t.Fatalf("Tools(* true) = %+v", tools)
	}
	if err := manager.SetToolEnabled("open", "*", false); err != nil {
		t.Fatalf("SetToolEnabled(* false) error = %v", err)
	}
	if tools, _ := manager.Tools("open"); len(tools) != 0 {
		t.Fatalf("Tools(* false) = %+v", tools)
	}
	if visible := manager.ToolsForSkills(nil); len(visible) != 0 {
		t.Fatalf("ToolsForSkills(no tools) = %+v", visible)
	}

	convertClient := &fakeMCPClient{tools: []ToolDef{{Name: "a"}, {Name: "b"}}}
	convert := NewManager([]ServerConfig{{Name: "convert", AllowedTools: []string{"a"}}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return convertClient, nil
	}))
	if err := convert.SetToolEnabled("convert", "*", true); err != nil {
		t.Fatalf("SetToolEnabled(* true before connect) error = %v", err)
	}
	if err := convert.Connect(context.Background(), "convert"); err != nil {
		t.Fatalf("Connect(convert) error = %v", err)
	}
	if tools, _ := convert.Tools("convert"); len(tools) != 1 || tools[0].Name != "a" {
		t.Fatalf("Tools(reconnect allowlist) = %+v", tools)
	}
	if err := convert.SetToolEnabled("convert", "b", true); err != nil {
		t.Fatalf("SetToolEnabled(convert b true) error = %v", err)
	}
	if tools, _ := convert.Tools("convert"); len(tools) != 2 {
		t.Fatalf("Tools(convert b true) = %+v", tools)
	}

	openClient := &fakeMCPClient{tools: []ToolDef{{Name: "b"}, {Name: "a"}}}
	open := NewManager([]ServerConfig{{Name: "open2"}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return openClient, nil
	}))
	if err := open.Connect(context.Background(), "open2"); err != nil {
		t.Fatalf("Connect(open2) error = %v", err)
	}
	if visible := open.ToolsForSkills(nil); len(visible) != 2 || visible[0].Name != "open2.a" {
		t.Fatalf("ToolsForSkills(open2) = %+v", visible)
	}
	if err := open.SetToolEnabled("open2", "a", true); err != nil {
		t.Fatalf("SetToolEnabled(open2 a true) error = %v", err)
	}
	if tools, _ := open.Tools("open2"); len(tools) != 2 {
		t.Fatalf("Tools(open2 a true) = %+v", tools)
	}
}

func TestHTTPConnectorAndProtocolClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		switch req.Method {
		case MethodInitialize:
			response, _ := NewResponse(req.ID, InitializeResult{ProtocolVersion: ProtocolVersion, ServerInfo: ServerInfo{Name: "fake", Version: "1"}})
			_ = json.NewEncoder(w).Encode(response)
		case MethodInitialized:
			w.WriteHeader(http.StatusNoContent)
		case MethodToolsList:
			response, _ := NewResponse(req.ID, ListToolsResult{Tools: []ToolDef{{Name: "ping", Description: "Ping."}}})
			_ = json.NewEncoder(w).Encode(response)
		case MethodToolsCall:
			response, _ := NewResponse(req.ID, CallToolResult{Content: []ContentItem{{Type: "text", Text: "pong"}}})
			_ = json.NewEncoder(w).Encode(response)
		default:
			t.Fatalf("unexpected method %q", req.Method)
		}
	}))
	defer server.Close()
	connector := HTTPConnector{HTTPClient: server.Client(), ClientName: "test", ClientVersion: "v"}
	client, err := connector.Connect(context.Background(), ServerConfig{Name: "http", URL: server.URL})
	if err != nil {
		t.Fatalf("HTTPConnector.Connect() error = %v", err)
	}
	if tools := client.Tools(); len(tools) != 1 || tools[0].Name != "ping" {
		t.Fatalf("HTTP client tools = %+v", tools)
	}
	result, err := client.CallTool(context.Background(), "ping", nil)
	if err != nil || result.Text() != "pong" {
		t.Fatalf("CallTool() = %+v, %v", result, err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := connector.Connect(context.Background(), ServerConfig{Name: "stdio", Command: "x"}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("HTTPConnector stdio error = %v", err)
	}
	defaultConnector := HTTPConnector{HTTPClient: server.Client()}
	defaultClient, err := defaultConnector.Connect(context.Background(), ServerConfig{Name: "http", URL: server.URL})
	if err != nil {
		t.Fatalf("HTTPConnector default Connect() error = %v", err)
	}
	if err := defaultClient.Close(); err != nil {
		t.Fatalf("default client Close() error = %v", err)
	}
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer failServer.Close()
	if _, err := (HTTPConnector{HTTPClient: failServer.Client()}).Connect(context.Background(), ServerConfig{Name: "bad", URL: failServer.URL}); err == nil {
		t.Fatal("HTTPConnector accepted initialize failure")
	}
}

func TestProtocolClientErrors(t *testing.T) {
	cases := []struct {
		name      string
		transport *scriptTransport
		call      bool
		args      map[string]interface{}
	}{
		{"init roundtrip", &scriptTransport{roundErr: errors.New("init failed")}, false, nil},
		{"init result", &scriptTransport{responses: []Response{NewErrorResponse(1, -1, "bad init")}}, false, nil},
		{"notify", &scriptTransport{responses: []Response{mustResponse(1, InitializeResult{})}, notifyErr: errors.New("notify failed")}, false, nil},
		{"tools roundtrip", &scriptTransport{responses: []Response{mustResponse(1, InitializeResult{})}, roundErrAt: 2, roundErr: errors.New("tools failed")}, false, nil},
		{"tools result", &scriptTransport{responses: []Response{mustResponse(1, InitializeResult{}), NewErrorResponse(2, -2, "bad tools")}}, false, nil},
		{"call request", &scriptTransport{}, true, map[string]interface{}{"bad": make(chan int)}},
		{"call roundtrip", &scriptTransport{roundErr: errors.New("call failed")}, true, nil},
		{"call result", &scriptTransport{responses: []Response{NewErrorResponse(1, -3, "bad call")}}, true, nil},
	}
	for _, tc := range cases {
		client := NewProtocolClient("x", "client", "v", tc.transport)
		var err error
		if tc.call {
			_, err = client.CallTool(context.Background(), "tool", tc.args)
		} else {
			err = client.Initialize(context.Background())
		}
		if err == nil {
			t.Fatalf("%s accepted failure", tc.name)
		}
	}
}

func TestSkillAllowsServerAndVisibleTools(t *testing.T) {
	if !SkillAllowsServer("ghidra-mcp", "ghidra") || !SkillAllowsServer("ghidra", "ghidra") {
		t.Fatal("SkillAllowsServer() rejected valid skill")
	}
	if SkillAllowsServer("", "ghidra") || SkillAllowsServer("other", "ghidra") {
		t.Fatal("SkillAllowsServer() accepted invalid skill")
	}
	visible := VisibleTools(ServerConfig{Name: "ghidra"}, []string{"ghidra-mcp"}, []string{"decompile"})
	if len(visible) != 1 || visible[0] != "decompile" {
		t.Fatalf("VisibleTools() = %v", visible)
	}
}

type fakeMCPClient struct {
	tools    []ToolDef
	called   string
	closed   bool
	closeErr error
}

type scriptTransport struct {
	responses  []Response
	calls      int
	roundErr   error
	roundErrAt int
	notifyErr  error
}

func (t *scriptTransport) RoundTrip(context.Context, Request) (Response, error) {
	t.calls++
	if t.roundErr != nil && (t.roundErrAt == 0 || t.roundErrAt == t.calls) {
		return Response{}, t.roundErr
	}
	if len(t.responses) == 0 {
		return Response{}, errors.New("no scripted response")
	}
	response := t.responses[0]
	t.responses = t.responses[1:]
	return response, nil
}

func (t *scriptTransport) Notify(context.Context, Notification) error {
	return t.notifyErr
}

func (t *scriptTransport) Close() error {
	return nil
}

func mustResponse(id int64, result interface{}) Response {
	response, err := NewResponse(id, result)
	if err != nil {
		panic(err)
	}
	return response
}

func (c *fakeMCPClient) Tools() []ToolDef {
	return append([]ToolDef(nil), c.tools...)
}

func (c *fakeMCPClient) CallTool(_ context.Context, name string, _ map[string]interface{}) (CallToolResult, error) {
	c.called = name
	return CallToolResult{Content: []ContentItem{{Type: "text", Text: "called " + name}}}, nil
}

func (c *fakeMCPClient) Close() error {
	c.closed = true
	return c.closeErr
}

func BenchmarkManagerToolsForSkills(b *testing.B) {
	tools := make([]ToolDef, 0, 128)
	for i := 0; i < 128; i++ {
		tools = append(tools, ToolDef{Name: "tool" + string(rune('a'+i%26))})
	}
	manager := NewManager([]ServerConfig{{Name: "ghidra", RequireSkill: true}}, ConnectorFunc(func(context.Context, ServerConfig) (Client, error) {
		return &fakeMCPClient{tools: tools}, nil
	}))
	if err := manager.Connect(context.Background(), "ghidra"); err != nil {
		b.Fatalf("Connect() error = %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if visible := manager.ToolsForSkills([]string{"ghidra-mcp"}); len(visible) != len(tools) {
			b.Fatalf("ToolsForSkills() len = %d, want %d", len(visible), len(tools))
		}
	}
}

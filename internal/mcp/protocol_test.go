package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProtocolConstructors(t *testing.T) {
	req, err := NewRequest(7, " tools/list ", map[string]string{"scope": "all"})
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if req.JSONRPC != JSONRPCVersion || req.ID != 7 || req.Method != "tools/list" || !strings.Contains(string(req.Params), "scope") {
		t.Fatalf("NewRequest() = %+v", req)
	}
	notification, err := NewNotification(MethodInitialized, nil)
	if err != nil {
		t.Fatalf("NewNotification() error = %v", err)
	}
	if notification.JSONRPC != JSONRPCVersion || notification.Method != MethodInitialized || notification.Params != nil {
		t.Fatalf("NewNotification() = %+v", notification)
	}
	response, err := NewResponse(7, map[string]string{"ok": "yes"})
	if err != nil {
		t.Fatalf("NewResponse() error = %v", err)
	}
	var decoded map[string]string
	if err := response.ResultAs(&decoded); err != nil || decoded["ok"] != "yes" {
		t.Fatalf("ResultAs() = %v, %v", decoded, err)
	}
	errResponse := NewErrorResponse(7, -32000, "boom")
	if err := errResponse.ResultAs(&decoded); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("ResultAs(error) = %v, want boom", err)
	}
}

func TestProtocolConstructorErrors(t *testing.T) {
	badPayload := make(chan int)
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"request id", requestErr(NewRequest(0, "x", nil))},
		{"request method", requestErr(NewRequest(1, " ", nil))},
		{"request marshal", requestErr(NewRequest(1, "x", badPayload))},
		{"notification method", notificationErr(NewNotification(" ", nil))},
		{"notification marshal", notificationErr(NewNotification("x", badPayload))},
		{"response id", responseErr(NewResponse(0, map[string]string{"x": "y"}))},
		{"response marshal", responseErr(NewResponse(1, badPayload))},
		{"response nil", responseErr(NewResponse(1, nil))},
	} {
		if tc.err == nil {
			t.Fatalf("%s accepted invalid input", tc.name)
		}
	}
}

func TestDecodeResponseAndResultAs(t *testing.T) {
	ok, err := DecodeResponse([]byte(`{"jsonrpc":"2.0","id":2,"result":{"name":"server"}}`))
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}
	var server ServerInfo
	if err := ok.ResultAs(&server); err != nil || server.Name != "server" {
		t.Fatalf("ResultAs(server) = %+v, %v", server, err)
	}
	for _, data := range [][]byte{
		[]byte(`{`),
		[]byte(`{"jsonrpc":"1.0","id":2,"result":{}}`),
		[]byte(`{"jsonrpc":"2.0","id":0,"result":{}}`),
		[]byte(`{"jsonrpc":"2.0","id":2}`),
	} {
		if _, err := DecodeResponse(data); err == nil {
			t.Fatalf("DecodeResponse(%s) accepted invalid response", string(data))
		}
	}
	if err := (Response{JSONRPC: JSONRPCVersion, ID: 3}).ResultAs(&server); err == nil {
		t.Fatal("ResultAs() accepted missing result")
	}
	if err := (Response{JSONRPC: JSONRPCVersion, ID: 3, Result: json.RawMessage(`"bad"`)}).ResultAs(&server); err == nil {
		t.Fatal("ResultAs() accepted wrong result shape")
	}
}

func TestMCPConvenienceTypes(t *testing.T) {
	initReq, err := InitializeRequest(1, "recomphamr", "0.2.0")
	if err != nil || initReq.Method != MethodInitialize || !strings.Contains(string(initReq.Params), ProtocolVersion) {
		t.Fatalf("InitializeRequest() = %+v, %v", initReq, err)
	}
	listReq, err := ToolsListRequest(2)
	if err != nil || listReq.Method != MethodToolsList || listReq.Params != nil {
		t.Fatalf("ToolsListRequest() = %+v, %v", listReq, err)
	}
	callReq, err := ToolsCallRequest(3, "decompile", map[string]interface{}{"addr": "0x1000"})
	if err != nil || callReq.Method != MethodToolsCall || !strings.Contains(string(callReq.Params), "decompile") {
		t.Fatalf("ToolsCallRequest() = %+v, %v", callReq, err)
	}
	schema := InputSchema{Type: "object", Properties: map[string]interface{}{"addr": map[string]string{"type": "string"}}, Required: []string{"addr"}}
	if got := schema.Map(); got["type"] != "object" || len(got["required"].([]string)) != 1 {
		t.Fatalf("InputSchema.Map() = %#v", got)
	}
	if got := (InputSchema{Type: "object"}).Map(); len(got["properties"].(map[string]interface{})) != 0 || len(got["required"].([]string)) != 0 {
		t.Fatalf("InputSchema.Map(nil) = %#v", got)
	}
	result := CallToolResult{Content: []ContentItem{{Type: "image", Text: "skip"}, {Type: "text", Text: "hello"}}, IsError: true}
	if result.Text() != "hello" || (CallToolResult{}).Text() != "" {
		t.Fatalf("CallToolResult.Text() mismatch")
	}
	if !strings.Contains((RPCError{Code: 1, Message: "bad"}).Error(), "bad") {
		t.Fatal("RPCError.Error() missing message")
	}
}

func requestErr(_ Request, err error) error {
	return err
}

func notificationErr(_ Notification, err error) error {
	return err
}

func responseErr(_ Response, err error) error {
	return err
}

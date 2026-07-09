package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestStdioTransportRoundTripAndNotify(t *testing.T) {
	response, err := json.Marshal(NewErrorResponse(99, -1, "unused"))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	success, err := NewResponse(1, map[string]string{"ok": "yes"})
	if err != nil {
		t.Fatalf("NewResponse() error = %v", err)
	}
	successData, err := json.Marshal(success)
	if err != nil {
		t.Fatalf("Marshal(success) error = %v", err)
	}
	stream := newScriptedStream(string(successData) + "\n" + string(response) + "\n")
	transport := NewStdioTransport("fake", stream)
	req, err := ToolsListRequest(1)
	if err != nil {
		t.Fatalf("ToolsListRequest() error = %v", err)
	}
	got, err := transport.RoundTrip(context.Background(), req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if got.ID != 1 || !strings.Contains(stream.Written(), MethodToolsList) {
		t.Fatalf("RoundTrip() got=%+v written=%q", got, stream.Written())
	}
	notif, err := NewNotification(MethodInitialized, nil)
	if err != nil {
		t.Fatalf("NewNotification() error = %v", err)
	}
	if err := transport.Notify(context.Background(), notif); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if !strings.Contains(stream.Written(), MethodInitialized) {
		t.Fatalf("Notify() did not write notification: %q", stream.Written())
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := transport.Notify(context.Background(), notif); err == nil {
		t.Fatal("Notify() accepted closed transport")
	}
}

func TestStdioTransportErrors(t *testing.T) {
	req, err := ToolsListRequest(1)
	if err != nil {
		t.Fatalf("ToolsListRequest() error = %v", err)
	}
	mismatched, _ := NewResponse(2, map[string]string{"ok": "yes"})
	mismatchedData, _ := json.Marshal(mismatched)
	if _, err := NewStdioTransport("fake", newScriptedStream(string(mismatchedData)+"\n")).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted mismatched response id")
	}
	if _, err := NewStdioTransport("fake", newScriptedStream(`{"jsonrpc":"1.0","id":1,"result":{}}`+"\n")).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted wrong jsonrpc")
	}
	if _, err := NewStdioTransport("fake", newScriptedStream(`{"jsonrpc":"2.0","id":0,"result":{}}`+"\n")).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted invalid response id")
	}
	writeFail := newScriptedStream("")
	writeFail.writeErr = errors.New("write failed")
	if _, err := NewStdioTransport("fake", writeFail).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted write failure")
	}
	readFail := newScriptedStream("")
	readFail.readErr = errors.New("read failed")
	if _, err := NewStdioTransport("fake", readFail).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted read failure")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewStdioTransport("fake", newScriptedStream("")).RoundTrip(ctx, req); err == nil {
		t.Fatal("RoundTrip() accepted canceled context before write")
	}
	invalidNotification := Notification{JSONRPC: JSONRPCVersion, Method: "bad", Params: json.RawMessage("{")}
	if err := NewStdioTransport("fake", newScriptedStream("")).Notify(context.Background(), invalidNotification); err == nil {
		t.Fatal("Notify() accepted invalid JSON payload")
	}
	blocking := newBlockingStream()
	timeout, stop := context.WithTimeout(context.Background(), time.Millisecond)
	defer stop()
	if _, err := NewStdioTransport("fake", blocking).RoundTrip(timeout, req); err == nil {
		t.Fatal("RoundTrip() accepted timeout")
	}
	closeFail := newScriptedStream("")
	closeFail.closeErr = errors.New("close failed")
	if err := NewStdioTransport("fake", closeFail).Close(); err == nil {
		t.Fatal("Close() accepted close failure")
	}
}

func TestHTTPTransportRoundTripAndNotify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp" || r.Header.Get("Content-Type") != "application/json" || !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			t.Fatalf("bad request path or headers: %s %v", r.URL.Path, r.Header)
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if req.ID == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		response, err := NewResponse(req.ID, map[string]string{"ok": req.Method})
		if err != nil {
			t.Fatalf("NewResponse() error = %v", err)
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}))
	defer server.Close()
	transport := NewHTTPTransport("fake", server.URL+"/", nil)
	req, err := ToolsListRequest(4)
	if err != nil {
		t.Fatalf("ToolsListRequest() error = %v", err)
	}
	response, err := transport.RoundTrip(context.Background(), req)
	if err != nil || response.ID != 4 {
		t.Fatalf("RoundTrip() = %+v, %v", response, err)
	}
	notification, err := NewNotification(MethodInitialized, nil)
	if err != nil {
		t.Fatalf("NewNotification() error = %v", err)
	}
	if err := transport.Notify(context.Background(), notification); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestHTTPTransportErrors(t *testing.T) {
	req, err := ToolsListRequest(1)
	if err != nil {
		t.Fatalf("ToolsListRequest() error = %v", err)
	}
	if _, err := NewHTTPTransport("fake", "", http.DefaultClient).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted empty base URL")
	}
	if _, err := NewHTTPTransport("fake", "http://[::1", http.DefaultClient).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted invalid URL")
	}
	badReq := req
	badReq.Params = json.RawMessage("{")
	if _, err := NewHTTPTransport("fake", "https://example.com", http.DefaultClient).RoundTrip(context.Background(), badReq); err == nil {
		t.Fatal("RoundTrip() accepted invalid JSON payload")
	}
	doFail := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("do failed")
	})}
	if _, err := NewHTTPTransport("fake", "https://example.com", doFail).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted HTTP client failure")
	}
	httpFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer httpFail.Close()
	if _, err := NewHTTPTransport("fake", httpFail.URL, httpFail.Client()).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted HTTP error")
	}
	readFail := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: errorReadCloser{}}, nil
	})}
	if _, err := NewHTTPTransport("fake", "https://example.com", readFail).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted body read error")
	}
	badJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer badJSON.Close()
	if _, err := NewHTTPTransport("fake", badJSON.URL, badJSON.Client()).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted malformed response")
	}
	mismatch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response, _ := NewResponse(2, map[string]string{"ok": "yes"})
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer mismatch.Close()
	if _, err := NewHTTPTransport("fake", mismatch.URL, mismatch.Client()).RoundTrip(context.Background(), req); err == nil {
		t.Fatal("RoundTrip() accepted mismatched response id")
	}
}

type scriptedStream struct {
	mu       sync.Mutex
	read     strings.Reader
	written  strings.Builder
	readErr  error
	writeErr error
	closeErr error
	closed   bool
}

func newScriptedStream(data string) *scriptedStream {
	return &scriptedStream{read: *strings.NewReader(data)}
}

func (s *scriptedStream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.readErr != nil {
		return 0, s.readErr
	}
	return s.read.Read(p)
}

func (s *scriptedStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	return s.written.Write(p)
}

func (s *scriptedStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return s.closeErr
}

func (s *scriptedStream) Written() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.written.String()
}

type blockingStream struct {
	closed chan struct{}
	once   sync.Once
}

func newBlockingStream() *blockingStream {
	return &blockingStream{closed: make(chan struct{})}
}

func (s *blockingStream) Read([]byte) (int, error) {
	<-s.closed
	return 0, io.ErrClosedPipe
}

func (s *blockingStream) Write(p []byte) (int, error) {
	return len(p), nil
}

func (s *blockingStream) Close() error {
	s.once.Do(func() { close(s.closed) })
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errorReadCloser) Close() error {
	return nil
}

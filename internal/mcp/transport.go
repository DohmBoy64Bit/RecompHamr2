package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

var execCommandContext = exec.CommandContext

// Transport exchanges JSON-RPC messages with one MCP server.
type Transport interface {
	// RoundTrip sends a request and waits for the matching response.
	RoundTrip(context.Context, Request) (Response, error)
	// Notify sends a notification without waiting for a response.
	Notify(context.Context, Notification) error
	// Close releases transport resources.
	Close() error
}

// StdioTransport implements line-delimited JSON-RPC over an injected stream.
type StdioTransport struct {
	name   string
	stream io.ReadWriteCloser
	dec    *json.Decoder
	mu     sync.Mutex
	closed bool
}

type stdioProcessStream struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	wait   func() error
	once   sync.Once
	err    error
}

// NewStdioTransport creates a stdio transport over stream.
func NewStdioTransport(name string, stream io.ReadWriteCloser) *StdioTransport {
	return &StdioTransport{name: name, stream: stream, dec: json.NewDecoder(stream)}
}

func startStdioProcess(ctx context.Context, command string, args []string) (io.ReadWriteCloser, error) {
	cmd := execCommandContext(ctx, command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp stdio stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("mcp stdio stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("mcp stdio start %s: %w", command, err)
	}
	return &stdioProcessStream{stdin: stdin, stdout: stdout, wait: cmd.Wait}, nil
}

// RoundTrip sends a line-delimited request and reads a response.
func (t *StdioTransport) RoundTrip(ctx context.Context, request Request) (Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.write(ctx, request); err != nil {
		return Response{}, err
	}
	response, err := t.read(ctx)
	if err != nil {
		return Response{}, err
	}
	if response.ID != request.ID {
		return Response{}, fmt.Errorf("mcp %s: response id %d did not match request id %d", t.name, response.ID, request.ID)
	}
	return response, nil
}

// Notify sends a line-delimited notification.
func (t *StdioTransport) Notify(ctx context.Context, notification Notification) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.write(ctx, notification)
}

// Close closes the underlying stdio stream.
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return t.stream.Close()
}

// Read reads bytes from the spawned MCP process stdout stream.
func (s *stdioProcessStream) Read(p []byte) (int, error) {
	return s.stdout.Read(p)
}

// Write writes bytes to the spawned MCP process stdin stream.
func (s *stdioProcessStream) Write(p []byte) (int, error) {
	return s.stdin.Write(p)
}

// Close closes process pipes and waits for the spawned MCP process to exit.
func (s *stdioProcessStream) Close() error {
	s.once.Do(func() {
		inErr := s.stdin.Close()
		outErr := s.stdout.Close()
		waitErr := s.wait()
		switch {
		case inErr != nil:
			s.err = inErr
		case outErr != nil:
			s.err = outErr
		default:
			s.err = waitErr
		}
	})
	return s.err
}

func (t *StdioTransport) write(ctx context.Context, value interface{}) error {
	if t.closed {
		return fmt.Errorf("mcp %s: transport closed", t.name)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if _, err := t.stream.Write(data); err != nil {
		return fmt.Errorf("mcp %s: stdio write: %w", t.name, err)
	}
	return nil
}

func (t *StdioTransport) read(ctx context.Context) (Response, error) {
	type result struct {
		response Response
		err      error
	}
	done := make(chan result, 1)
	go func() {
		var response Response
		err := t.dec.Decode(&response)
		if err == nil {
			if response.JSONRPC != JSONRPCVersion {
				err = fmt.Errorf("mcp %s: unsupported jsonrpc %q", t.name, response.JSONRPC)
			} else if response.ID <= 0 {
				err = fmt.Errorf("mcp %s: response id must be positive", t.name)
			}
		}
		done <- result{response: response, err: err}
	}()
	select {
	case got := <-done:
		if got.err != nil {
			return Response{}, got.err
		}
		return got.response, nil
	case <-ctx.Done():
		t.closed = true
		_ = t.stream.Close()
		return Response{}, ctx.Err()
	}
}

// HTTPTransport implements MCP streamable HTTP request/response foundations.
type HTTPTransport struct {
	name    string
	baseURL string
	client  *http.Client
}

// NewHTTPTransport creates an HTTP transport that posts JSON-RPC to baseURL.
func NewHTTPTransport(name string, baseURL string, client *http.Client) *HTTPTransport {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPTransport{name: name, baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

// RoundTrip posts a request to the server `/mcp` endpoint.
func (t *HTTPTransport) RoundTrip(ctx context.Context, request Request) (Response, error) {
	response, err := t.post(ctx, request)
	if err != nil {
		return Response{}, err
	}
	if response.ID != request.ID {
		return Response{}, fmt.Errorf("mcp %s: response id %d did not match request id %d", t.name, response.ID, request.ID)
	}
	return response, nil
}

// Notify posts a notification to the server `/mcp` endpoint.
func (t *HTTPTransport) Notify(ctx context.Context, notification Notification) error {
	_, err := t.post(ctx, notification)
	return err
}

// Close releases HTTP transport resources.
func (t *HTTPTransport) Close() error {
	return nil
}

func (t *HTTPTransport) post(ctx context.Context, value interface{}) (Response, error) {
	if t.baseURL == "" {
		return Response{}, fmt.Errorf("mcp %s: HTTP base URL is empty", t.name)
	}
	body, err := json.Marshal(value)
	if err != nil {
		return Response{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/mcp", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := t.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Response{}, fmt.Errorf("mcp %s: HTTP %d: %s", t.name, resp.StatusCode, string(data))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return Response{}, nil
	}
	return DecodeResponse(data)
}

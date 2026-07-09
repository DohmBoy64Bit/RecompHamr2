package mcp

import "testing"

func TestBuiltinsAndEnv(t *testing.T) {
	t.Setenv("RECOMPHAMR_MCP_GHIDRA_COMMAND", "custom-ghidra")
	t.Setenv("RECOMPHAMR_MCP_GHIDRA_URL", "http://ghidra")
	t.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", "a,b")
	t.Setenv("RECOMPHAMR_MCP_AUTOSTART", "1")
	servers := Builtins()
	if len(servers) != 8 {
		t.Fatalf("len(Builtins()) = %d, want 8", len(servers))
	}
	if servers[0].Command != "custom-ghidra" || servers[0].URL != "http://ghidra" || len(servers[0].AllowedTools) != 2 || !servers[0].Autostart {
		t.Fatalf("env overrides not applied: %+v", servers[0])
	}
}

func TestVisibleTools(t *testing.T) {
	server := ServerConfig{Name: "ghidra"}
	if got := VisibleTools(server, nil, []string{"x"}); len(got) != 0 {
		t.Fatalf("VisibleTools() without skill = %v, want none", got)
	}
	got := VisibleTools(server, []string{"ghidra"}, []string{"x"})
	if len(got) != 1 || got[0] != "x" {
		t.Fatalf("VisibleTools() with skill = %v, want x", got)
	}
	got[0] = "mutated"
	again := VisibleTools(server, []string{"ghidra"}, []string{"x"})
	if again[0] != "x" {
		t.Fatal("VisibleTools() returned shared slice")
	}
}

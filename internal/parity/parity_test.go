package parity

import "testing"

func TestCurrentInventoryCopiesReferenceFacts(t *testing.T) {
	inv := CurrentInventory()
	if inv.ReferenceURL != ReferenceURL || inv.ReferenceCommit != ReferenceCommit {
		t.Fatalf("reference mismatch: %+v", inv)
	}
	if len(inv.SlashCommands) != 11 {
		t.Fatalf("SlashCommands count = %d, want 11", len(inv.SlashCommands))
	}
	if len(inv.BuiltInTools) != 6 {
		t.Fatalf("BuiltInTools count = %d, want 6", len(inv.BuiltInTools))
	}
	if len(inv.MCPServers) != 8 {
		t.Fatalf("MCPServers count = %d, want 8", len(inv.MCPServers))
	}
	if len(inv.SkillNames) != 28 {
		t.Fatalf("SkillNames count = %d, want 28", len(inv.SkillNames))
	}

	inv.SlashCommands[0] = "mutated"
	if SlashCommands[0] == "mutated" {
		t.Fatal("CurrentInventory returned shared slices")
	}
}

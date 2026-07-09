package parity

// ReferenceURL is the authoritative RecompHamr 1.x source repository.
const ReferenceURL = "https://github.com/DohmBoy64Bit/RecompHamr"

// ReferenceCommit is the inspected RecompHamr 1.x commit.
const ReferenceCommit = "259a450e93af48437ee23663e5ca66cdc1ab8569"

// SlashCommands lists the RecompHamr 1.x slash commands.
var SlashCommands = []string{"/help", "/clear", "/models", "/skills", "/skill", "/skill-audit", "/skill-new", "/init-re", "/status-re", "/doctor", "/mcp"}

// BuiltInTools lists the RecompHamr 1.x built-in LLM tools.
var BuiltInTools = []string{"bash", "read_file", "write_file", "edit_file", "repomixr", "recomp_reference"}

// MCPServers lists the RecompHamr 1.x built-in MCP server configs.
var MCPServers = []string{"ghidra", "n64-debug-mcp", "pcrecomp", "mcp-pine", "objdiff", "pcsx2", "bizhawk", "sega2asm"}

// SkillNames lists the RecompHamr 1.x embedded skills.
var SkillNames = []string{
	"bizhawk", "build-fix-loop", "cdb-debug", "core-re", "evidence-mode", "file-format-reversing",
	"function-discovery", "gb-recomp", "gc-decomp", "gen-decomp", "ghidra-mcp", "imhex",
	"mcp-pine", "n64-debug-mcp", "n64-decomp", "objdiff", "pcrecomp", "pcsx2",
	"project-handoff", "ps2recomp", "ps3recomp", "recomp-foundations", "sega2asm", "snesrecomp",
	"vb-decomp", "windows-game-decomp", "xbox360-decomp", "xboxrecomp",
}

// Inventory summarizes the reference facts.
type Inventory struct {
	ReferenceURL    string
	ReferenceCommit string
	SlashCommands   []string
	BuiltInTools    []string
	MCPServers      []string
	SkillNames      []string
}

// CurrentInventory returns a copy of the inspected RecompHamr 1.x inventory.
func CurrentInventory() Inventory {
	return Inventory{
		ReferenceURL:    ReferenceURL,
		ReferenceCommit: ReferenceCommit,
		SlashCommands:   append([]string(nil), SlashCommands...),
		BuiltInTools:    append([]string(nil), BuiltInTools...),
		MCPServers:      append([]string(nil), MCPServers...),
		SkillNames:      append([]string(nil), SkillNames...),
	}
}

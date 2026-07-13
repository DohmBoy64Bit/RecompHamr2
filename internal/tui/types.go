package tui

import (
	"github.com/charmbracelet/colorprofile"

	"recomphamr2/internal/commands"
)

const (
	// DefaultWidth is the deterministic wide render width.
	DefaultWidth = 120
	// DefaultHeight is the deterministic render height.
	DefaultHeight = 32
	// CompactWidth is the standard-to-narrow breakpoint.
	CompactWidth = 80
	// MinimumWidth is the smallest supported interactive width.
	MinimumWidth = 60
	// MinimumHeight is the smallest supported interactive height.
	MinimumHeight = 18
)

// Snapshot is immutable app-owned state rendered by the TUI.
type Snapshot struct {
	// Env provides read-only command and picker metadata.
	Env commands.Environment
	// Mode is the verified runtime mode.
	Mode string
	// ActiveModel is the configured model profile.
	ActiveModel string
	// ActiveSkill is the active skill name.
	ActiveSkill string
	// MCPStatus is the verified MCP summary.
	MCPStatus string
	// ContextStatus is the verified context summary.
	ContextStatus string
	// PendingTool is the active tool name or none.
	PendingTool string
	// MemoryStatus is the verified project-memory summary.
	MemoryStatus string
	// Status is transient app-owned feedback.
	Status string
	// Secrets are redacted before transcript rendering.
	Secrets []string
}

// TranscriptKind identifies a visible transcript block.
type TranscriptKind string

const (
	// TranscriptUser is submitted user input.
	TranscriptUser TranscriptKind = "user"
	// TranscriptAssistant is model output.
	TranscriptAssistant TranscriptKind = "assistant"
	// TranscriptTool is built-in tool output.
	TranscriptTool TranscriptKind = "tool"
	// TranscriptMCP is MCP tool output.
	TranscriptMCP TranscriptKind = "mcp"
	// TranscriptVerified is verification evidence.
	TranscriptVerified TranscriptKind = "verified"
	// TranscriptWarning is actionable warning output.
	TranscriptWarning TranscriptKind = "warning"
	// TranscriptBlocked is a blocked operation.
	TranscriptBlocked TranscriptKind = "blocked"
	// TranscriptUnsupported is an unsupported operation.
	TranscriptUnsupported TranscriptKind = "unsupported"
	// TranscriptAttachment is attachment evidence.
	TranscriptAttachment TranscriptKind = "attachment"
	// TranscriptNote is neutral informational output.
	TranscriptNote TranscriptKind = "note"
)

// TranscriptEntry is one semantic transcript block.
type TranscriptEntry struct {
	// Kind determines the visible label and semantic style.
	Kind TranscriptKind
	// Text is the block body without a repeated semantic prefix.
	Text string
}

// IntentKind identifies an app-owned effect requested by the TUI.
type IntentKind string

const (
	// IntentSubmit requests one agent prompt.
	IntentSubmit IntentKind = "submit"
	// IntentCommand requests one slash command.
	IntentCommand IntentKind = "command"
	// IntentCancel requests cancellation.
	IntentCancel IntentKind = "cancel"
	// IntentQuit requests clean process exit.
	IntentQuit IntentKind = "quit"
	// IntentModel requests a model selection.
	IntentModel IntentKind = "model"
	// IntentSkill requests a skill selection.
	IntentSkill IntentKind = "skill"
	// IntentMCP requests an MCP selection.
	IntentMCP IntentKind = "mcp"
)

// IntentMsg carries exactly one TUI request to internal/app.
type IntentMsg struct {
	// Kind identifies the requested effect.
	Kind IntentKind
	// Value carries submitted text or a selected identifier.
	Value string
}

// SnapshotMsg replaces the immutable app-owned render snapshot.
type SnapshotMsg struct {
	// Snapshot is the latest verified runtime state.
	Snapshot Snapshot
}

// TranscriptMsg appends semantic transcript entries.
type TranscriptMsg struct {
	// Entries are appended in order.
	Entries []TranscriptEntry
}

// ClearTranscriptMsg removes all visible transcript entries.
type ClearTranscriptMsg struct{}

// ColorProfileMsg selects a deterministic render profile in tests.
type ColorProfileMsg struct {
	// Profile is the terminal color capability.
	Profile colorprofile.Profile
}

type overlayKind string

const (
	overlayNone     overlayKind = ""
	overlayCommands overlayKind = "commands"
	overlayModels   overlayKind = "models"
	overlaySkills   overlayKind = "skills"
	overlayMCP      overlayKind = "mcp"
	overlayHelp     overlayKind = "help"
)

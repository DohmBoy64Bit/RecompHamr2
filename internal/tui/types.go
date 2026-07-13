package tui

import (
	"github.com/charmbracelet/colorprofile"

	"recomphamr2/internal/commands"
)

// Model is the testable terminal shell state.
type Model struct {
	// Transcript is the visible conversation and command output.
	Transcript []string
	// Env is command execution state owned by internal/commands.
	Env commands.Environment
	// Layout stores current render metadata.
	Layout Layout
	// Composer is the multiline prompt buffer.
	Composer string
	// History stores submitted prompt text newest-last.
	History []string
	// HistoryIndex is the active prompt-history cursor, or len(History).
	HistoryIndex int
	// Attachments stores large paste chips until the next submission.
	Attachments []Attachment
	// Status is the footer status text.
	Status string
	// PaletteIndex is the selected row in the active command or modal overlay.
	PaletteIndex int
	// DebugEnabled controls whether redacted debug lines render.
	DebugEnabled bool
	// DebugLog stores redacted debug entries.
	DebugLog []string
	// DebugSecrets are values removed from debug output.
	DebugSecrets []string
	// QuitArmed records the first idle Ctrl+C in the double-press quit flow.
	QuitArmed bool
	// TranscriptOffset counts transcript entries below the visible window.
	TranscriptOffset int
	// NewOutput reports output received while transcript follow mode is paused.
	NewOutput bool
}

// Layout contains the visible TUI state that can be rendered without Bubble Tea.
type Layout struct {
	// ColorProfile is the terminal color capability used for deterministic styling.
	ColorProfile colorprofile.Profile
	// Width is the terminal width in cells.
	Width int
	// Height is the terminal height in cells.
	Height int
	// Mode is the current UI mode label.
	Mode string
	// ActiveModel is the selected model profile label.
	ActiveModel string
	// ActiveSkill is the active skill indicator.
	ActiveSkill string
	// MCPStatus is the MCP gate/status indicator.
	MCPStatus string
	// ContextStatus is the context-budget evidence indicator.
	ContextStatus string
	// PendingTool is the currently visible tool status.
	PendingTool string
	// MemoryStatus is the memory freshness indicator.
	MemoryStatus string
}

// Event is one testable TUI update message.
type Event struct {
	// Key is a symbolic key constant such as KeyEnter.
	Key string
	// Text is inserted into the composer.
	Text string
	// Paste is pasted text, converted to a chip when large or multiline.
	Paste string
	// Width updates Layout.Width when positive.
	Width int
	// Height updates Layout.Height when positive.
	Height int
}

// Action is the side effect requested by Update.
type Action string

// IntentKind identifies an app-owned effect requested by the terminal UI.
type IntentKind string

// Intent carries one side-effect request from the TUI to internal/app.
type Intent struct {
	// Kind identifies the requested effect.
	Kind IntentKind
	// Value carries submitted text or a selected command/profile/server name.
	Value string
}

// Attachment describes one large paste chip held outside the composer text.
type Attachment struct {
	// Name is the visible chip identifier.
	Name string
	// Content is the pasted text associated with the chip.
	Content string
}

// BubbleModel adapts Model to the Bubble Tea runtime without owning core logic.
type BubbleModel struct {
	// State is the pure TUI shell state rendered by View.
	State Model
	// LastAction records the latest side effect requested by Update.
	LastAction Action
	// LastIntent is the typed effect request produced by the latest update.
	LastIntent Intent
	components bubbleComponents
}

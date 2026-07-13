package tui

const (
	// DefaultWidth is the canonical golden-render width for tests and docs.
	DefaultWidth = 120
	// DefaultHeight is the canonical terminal height used by renderer tests.
	DefaultHeight = 32
	// CompactWidth is the threshold below which the evidence panel collapses.
	CompactWidth = 96
	// LargePasteThreshold is the byte count that turns paste text into a chip.
	LargePasteThreshold = 1024
	// MinimumWidth is the narrowest supported interactive terminal width.
	MinimumWidth = 60
	// MinimumHeight is the shortest supported interactive terminal height.
	MinimumHeight = 18
)

const (
	brandWide    = "RECOMP HAMR"
	brandCompact = "RecompHamr"
)

const (
	// KeyEnter submits the current composer text.
	KeyEnter = "enter"
	// KeyBackspace deletes the last composer rune.
	KeyBackspace = "backspace"
	// KeyUp recalls the previous prompt history entry.
	KeyUp = "up"
	// KeyDown recalls the next prompt history entry.
	KeyDown = "down"
	// KeyTab completes the current slash command candidate.
	KeyTab = "tab"
	// KeyCtrlC cancels active work or arms quit when idle.
	KeyCtrlC = "ctrl+c"
	// KeyCtrlD quits immediately.
	KeyCtrlD = "ctrl+d"
	// KeyEsc clears transient palette and quit state.
	KeyEsc = "esc"
	// KeyPageUp scrolls the transcript toward older output.
	KeyPageUp = "pgup"
	// KeyPageDown scrolls the transcript toward newer output.
	KeyPageDown = "pgdown"
)

const (
	// ActionNone means an update changed only local UI state.
	ActionNone Action = "none"
	// ActionSubmit means the composer submitted user or slash-command text.
	ActionSubmit Action = "submit"
	// ActionCancel means the UI requested cancellation of active work.
	ActionCancel Action = "cancel"
	// ActionQuit means the UI requested process exit.
	ActionQuit Action = "quit"
)

const (
	// IntentNone means no app-owned effect was requested.
	IntentNone IntentKind = "none"
	// IntentSubmit requests a prompt or slash-command submission.
	IntentSubmit IntentKind = "submit"
	// IntentCancel requests cancellation of active work.
	IntentCancel IntentKind = "cancel"
	// IntentQuit requests a clean application exit.
	IntentQuit IntentKind = "quit"
	// IntentCommand requests dispatch of a canonical slash command.
	IntentCommand IntentKind = "command"
	// IntentModel requests selection of a model profile through app ownership.
	IntentModel IntentKind = "model"
	// IntentSkill requests selection of a skill through app ownership.
	IntentSkill IntentKind = "skill"
	// IntentMCP requests an MCP action through app ownership.
	IntentMCP IntentKind = "mcp"
)

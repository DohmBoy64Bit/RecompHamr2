package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"recomphamr2/internal/commands"
)

// NewBubble returns a Bubble Tea adapter around a fresh Model.
func NewBubble(env commands.Environment) BubbleModel {
	b := BubbleModel{State: New(env), LastAction: ActionNone, LastIntent: Intent{Kind: IntentNone}, components: newBubbleComponents()}
	b.components.syncFromState(b.State)
	return b
}

// Init satisfies tea.Model without launching background work.
func (b BubbleModel) Init() tea.Cmd {
	return nil
}

// Update translates Bubble Tea messages into pure Model events.
func (b BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	b.components.syncFromState(b.State)
	submitted := b.State.Composer
	if updated, cmd, handled := b.updateComponentMessage(msg); handled {
		return updated, cmd
	}
	event, ok := bubbleEvent(msg)
	if !ok {
		return b, nil
	}
	state, action := b.State.Update(event)
	b.State = state
	b.LastAction = action
	b.LastIntent = intentFromAction(action, submitted)
	b.components.syncFromState(b.State)
	return b, nil
}

func (b BubbleModel) updateComponentMessage(msg tea.Msg) (BubbleModel, tea.Cmd, bool) {
	switch typed := msg.(type) {
	case tea.ColorProfileMsg:
		b.State.Layout.ColorProfile = typed.Profile
		return b, nil, true
	case tea.FocusMsg:
		return b, b.components.composer.Focus(), true
	case tea.BlurMsg:
		b.components.composer.Blur()
		return b, nil, true
	case tea.MouseWheelMsg:
		if len(b.State.Transcript) == 0 {
			return b, nil, true
		}
		var cmd tea.Cmd
		b.components.transcript, cmd = b.components.transcript.Update(msg)
		delta := -3
		if typed.Mouse().Button == tea.MouseWheelUp {
			delta = 3
		}
		b.State = b.State.scrollTranscript(delta)
		return b, cmd, true
	case tea.KeyPressMsg:
		keyName := strings.ToLower(typed.String())
		if keyName == "enter" {
			if intent, ok := b.State.selectedOverlayIntent(); ok {
				b.LastAction = ActionNone
				b.LastIntent = intent
				b.State.Composer = ""
				b.State.PaletteIndex = 0
				b.components.syncFromState(b.State)
				return b, nil, true
			}
		}
		if len(b.State.overlayRows()) > 0 && (typed.Key().Text == "j" || typed.Key().Text == "k") {
			key := KeyDown
			if typed.Key().Text == "k" {
				key = KeyUp
			}
			b.State, b.LastAction = b.State.Update(Event{Key: key})
			b.LastIntent = Intent{Kind: IntentNone}
			return b, nil, true
		}
		if keyName == "pgup" || keyName == "pgdown" {
			if len(b.State.Transcript) == 0 {
				return b, nil, true
			}
			var cmd tea.Cmd
			b.components.transcript, cmd = b.components.transcript.Update(msg)
			delta := -(b.State.Layout.Height / 3)
			if keyName == "pgup" {
				delta = b.State.Layout.Height / 3
			}
			if delta == 0 {
				delta = -1
				if keyName == "pgup" {
					delta = 1
				}
			}
			b.State = b.State.scrollTranscript(delta)
			return b, cmd, true
		}
		if keyName == "shift+enter" || keyName == "ctrl+j" {
			b.components.composer.InsertString("\n")
			b.State.Composer = b.components.composer.Value()
			b.LastAction = ActionNone
			b.LastIntent = Intent{Kind: IntentNone}
			return b, nil, true
		}
		if b.composerOwnsKey(typed) {
			var cmd tea.Cmd
			b.components.composer, cmd = b.components.composer.Update(msg)
			b.State.Composer = b.components.composer.Value()
			b.State.QuitArmed = false
			b.State.PaletteIndex = 0
			b.LastAction = ActionNone
			b.LastIntent = Intent{Kind: IntentNone}
			return b, cmd, true
		}
	}
	return b, nil, false
}

func (m Model) selectedOverlayIntent() (Intent, bool) {
	kind := m.overlayKind()
	if kind == "" || kind == "commands" || kind == "help" {
		return Intent{}, false
	}
	rows := m.overlayRows()
	name, ok := selectedRowName(rows, m.PaletteIndex)
	if !ok {
		return Intent{}, false
	}
	return Intent{Kind: intentKindForOverlay(kind), Value: name}, true
}

func selectedRowName(rows []string, index int) (string, bool) {
	if len(rows) == 0 {
		return "", false
	}
	if index < 0 || index >= len(rows) {
		index = 0
	}
	fields := strings.Fields(strings.TrimLeft(rows[index], "*! "))
	if len(fields) == 0 || fields[0] == "blocked" {
		return "", false
	}
	return fields[0], true
}

func intentKindForOverlay(kind string) IntentKind {
	if kind == "skills" {
		return IntentSkill
	}
	if kind == "mcp" {
		return IntentMCP
	}
	return IntentModel
}

func (b BubbleModel) composerOwnsKey(msg tea.KeyPressMsg) bool {
	name := strings.ToLower(msg.String())
	if msg.Key().Text != "" {
		return true
	}
	switch name {
	case "backspace", "delete", "left", "right", "home", "end", "ctrl+a", "ctrl+e", "ctrl+k", "ctrl+u", "ctrl+w":
		return true
	case "up", "down":
		return len(b.State.overlayRows()) == 0 && b.components.composer.LineCount() > 1
	default:
		return false
	}
}

// View renders the current Bubble Tea model.
func (b BubbleModel) View() tea.View {
	b.components.syncFromState(b.State)
	screen := b.State.styledScreen(b.State.Layout)
	view := tea.NewView(screen.content)
	view.AltScreen = true
	if len(b.State.Transcript) > 0 {
		view.MouseMode = tea.MouseModeCellMotion
	} else {
		view.MouseMode = tea.MouseModeNone
	}
	view.ReportFocus = true
	view.WindowTitle = "RecompHamr"
	if !screen.hideCursor {
		view.Cursor = tea.NewCursor(screen.cursorX, screen.cursorY)
		view.Cursor.Shape = tea.CursorBar
		view.Cursor.Blink = true
	}
	return view
}

func bubbleEvent(msg tea.Msg) (Event, bool) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		return Event{Width: typed.Width, Height: typed.Height}, true
	case tea.KeyPressMsg:
		key := typed.Key()
		if key.Text != "" {
			return Event{Text: key.Text}, true
		}
		return Event{Key: bubbleKey(typed.String())}, true
	case tea.PasteMsg:
		return Event{Paste: typed.Content}, true
	default:
		return Event{}, false
	}
}

func bubbleKey(key string) string {
	switch strings.ToLower(key) {
	case KeyEnter, KeyBackspace, KeyUp, KeyDown, KeyCtrlC, KeyCtrlD, KeyEsc:
		return key
	case KeyTab:
		return KeyTab
	default:
		return key
	}
}

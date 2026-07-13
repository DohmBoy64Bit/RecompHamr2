package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
)

type bubbleComponents struct {
	initialized bool
	composer    textarea.Model
	transcript  viewport.Model
	help        help.Model
	keys        keyMap
}

type keyMap struct {
	Commands key.Binding
	Complete key.Binding
	Help     key.Binding
	Cancel   key.Binding
	Quit     key.Binding
}

func newBubbleComponents() bubbleComponents {
	composer := textarea.New()
	composer.Placeholder = `Ask RecompHamr... "map this function"`
	composer.Prompt = ""
	composer.ShowLineNumbers = false
	composer.DynamicHeight = true
	composer.MinHeight = 1
	composer.MaxHeight = 7
	composer.MaxContentHeight = 200
	composer.SetWidth(DefaultWidth - 20)
	composer.SetHeight(1)
	_ = composer.Focus()

	transcript := viewport.New(viewport.WithWidth(DefaultWidth), viewport.WithHeight(DefaultHeight-6))
	transcript.SoftWrap = true
	transcript.FillHeight = true
	transcript.MouseWheelEnabled = true

	helpModel := help.New()
	helpModel.SetWidth(DefaultWidth)
	return bubbleComponents{
		initialized: true,
		composer:    composer,
		transcript:  transcript,
		help:        helpModel,
		keys:        defaultKeyMap(),
	}
}

func defaultKeyMap() keyMap {
	return keyMap{
		Commands: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "commands")),
		Complete: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Cancel:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "cancel")),
		Quit:     key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "exit")),
	}
}

// ShortHelp returns the context-neutral bindings shown in the compact footer.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Commands, k.Complete, k.Help}
}

// FullHelp returns all documented TUI bindings grouped for the help overlay.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Commands, k.Complete, k.Help}, {k.Cancel, k.Quit}}
}

func (c *bubbleComponents) syncFromState(state Model) {
	if !c.initialized {
		*c = newBubbleComponents()
	}
	if c.composer.Value() != state.Composer {
		c.composer.SetValue(state.Composer)
	}
	width, height := bubbleSize(state.Layout)
	composerWidth := launcherPanelWidth(width) - 5
	if len(state.Transcript) > 0 {
		composerWidth = width - 5
	}
	c.composer.SetWidth(composerWidth)
	c.transcript.SetWidth(width)
	transcriptHeight := height - 6
	if transcriptHeight < 3 {
		transcriptHeight = 3
	}
	c.transcript.SetHeight(transcriptHeight)
	c.transcript.SetContent(transcriptText(state))
	c.help.SetWidth(width)
}

func intentFromAction(action Action, value string) Intent {
	switch action {
	case ActionSubmit:
		return Intent{Kind: IntentSubmit, Value: value}
	case ActionCancel:
		return Intent{Kind: IntentCancel}
	case ActionQuit:
		return Intent{Kind: IntentQuit}
	default:
		return Intent{Kind: IntentNone}
	}
}

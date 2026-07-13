package tui

import (
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

type pickerItem struct {
	name        string
	description string
	detail      string
	kind        IntentKind
	blocked     bool
}

// Title returns the picker row identifier required by the default delegate.
func (i pickerItem) Title() string { return i.name }

// Description returns the picker row summary required by the default delegate.
func (i pickerItem) Description() string { return i.description }

// FilterValue returns the searchable picker text.
func (i pickerItem) FilterValue() string { return i.name + " " + i.description }

type keyMap struct {
	commands key.Binding
	complete key.Binding
	help     key.Binding
	cancel   key.Binding
	quit     key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		commands: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "commands")),
		complete: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		cancel:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "cancel")),
		quit:     key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "exit")),
	}
}

// ShortHelp returns context-neutral footer bindings.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.commands, k.complete, k.help}
}

// FullHelp returns all documented binding groups.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.commands, k.complete, k.help}, {k.cancel, k.quit}}
}

func newComposer() textarea.Model {
	composer := textarea.New()
	composer.Placeholder = `Ask RecompHamr... "map this function"`
	composer.Prompt = ""
	composer.ShowLineNumbers = false
	composer.DynamicHeight = true
	composer.MinHeight = 1
	composer.MaxHeight = 7
	composer.MaxContentHeight = 200
	composer.SetWidth(72)
	composer.SetHeight(1)
	_ = composer.Focus()
	return composer
}

func newViewport() viewport.Model {
	v := viewport.New(viewport.WithWidth(DefaultWidth), viewport.WithHeight(DefaultHeight-8))
	v.SoftWrap = true
	v.FillHeight = true
	v.MouseWheelEnabled = true
	return v
}

func newPicker(profile colorprofile.Profile) list.Model {
	delegate := pickerDelegate{profile: profile}
	picker := list.New(nil, delegate, 72, 16)
	picker.Title = "Commands"
	picker.SetShowHelp(false)
	picker.SetShowStatusBar(true)
	picker.SetShowPagination(true)
	picker.SetFilteringEnabled(true)
	return picker
}

func styleComponents(composer *textarea.Model, picker *list.Model, helpModel *help.Model, profile colorprofile.Profile) {
	t := makeTheme(profile)

	textareaStyles := textarea.DefaultDarkStyles()
	textareaStyles.Focused.Base = t.body
	textareaStyles.Focused.CursorLine = t.composer
	textareaStyles.Focused.CursorLineNumber = t.muted
	textareaStyles.Focused.EndOfBuffer = t.muted
	textareaStyles.Focused.LineNumber = t.muted
	textareaStyles.Focused.Placeholder = t.muted
	textareaStyles.Focused.Prompt = t.accent
	textareaStyles.Focused.Text = t.body
	textareaStyles.Blurred = textareaStyles.Focused
	textareaStyles.Cursor.Color = t.accent.GetForeground()
	composer.SetStyles(textareaStyles)

	picker.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	picker.Styles.Title = t.brand.Bold(true).Padding(0, 1)
	picker.Styles.Spinner = t.accent
	picker.Styles.Filter.Focused.Text = t.body
	picker.Styles.Filter.Focused.Placeholder = t.muted
	picker.Styles.Filter.Focused.Suggestion = t.muted
	picker.Styles.Filter.Focused.Prompt = t.accent
	picker.Styles.Filter.Blurred = picker.Styles.Filter.Focused
	picker.Styles.Filter.Cursor.Color = t.accent.GetForeground()
	picker.Styles.DefaultFilterCharacterMatch = t.accent.Underline(true)
	picker.Styles.StatusBar = t.muted.Padding(0, 0, 1, 2)
	picker.Styles.StatusEmpty = t.warning
	picker.Styles.StatusBarActiveFilter = t.body
	picker.Styles.StatusBarFilterCount = t.muted
	picker.Styles.NoItems = t.warning
	picker.Styles.PaginationStyle = t.muted.PaddingLeft(2)
	picker.Styles.HelpStyle = t.muted.Padding(1, 0, 0, 2)
	picker.Styles.ActivePaginationDot = t.accent.SetString("•")
	picker.Styles.InactivePaginationDot = t.muted.SetString("•")
	picker.Styles.ArabicPagination = t.muted
	picker.Styles.DividerDot = t.muted.SetString(" • ")
	picker.SetDelegate(pickerDelegate{profile: profile})

	helpModel.Styles.ShortKey = t.accent
	helpModel.Styles.ShortDesc = t.muted
	helpModel.Styles.ShortSeparator = t.muted
	helpModel.Styles.Ellipsis = t.muted
	helpModel.Styles.FullKey = t.accent
	helpModel.Styles.FullDesc = t.muted
	helpModel.Styles.FullSeparator = t.muted
}

type pickerDelegate struct {
	profile colorprofile.Profile
}

// Height returns the one-row picker item height.
func (d pickerDelegate) Height() int { return 1 }

// Spacing returns the gap between picker rows.
func (d pickerDelegate) Spacing() int { return 0 }

// Update leaves interaction state under the parent Bubbles list.
func (d pickerDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

// Render draws one semantic picker row without owning selection state.
func (d pickerDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	row, ok := item.(pickerItem)
	if !ok {
		return
	}
	t := makeTheme(d.profile)
	selected := index == model.Index()
	marker := "  "
	if selected {
		marker = "> "
	}
	available := max(16, model.Width()-4)
	nameWidth := min(22, max(12, available/3))
	name := padRight(truncate(row.name, nameWidth), nameWidth)
	description := row.description
	if row.blocked {
		description += "  [blocked]"
	}
	line := marker + name + truncate(description, max(1, available-nameWidth-2))
	style := t.body
	if row.blocked {
		style = t.blocked
	}
	if selected {
		style = t.accent.Bold(true).Reverse(true)
	}
	_, _ = io.WriteString(writer, style.Render(lipgloss.NewStyle().MaxWidth(available+2).Render(line)))
}

func padRight(text string, width int) string {
	return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
}

func commandItems() []list.Item {
	registry := commandsRegistry()
	items := make([]list.Item, 0, len(registry))
	for _, command := range registry {
		items = append(items, pickerItem{
			name:        command.Name,
			description: command.Summary,
			detail:      command.Usage + "  " + command.SideEffects,
			kind:        IntentCommand,
		})
	}
	return items
}

func pickerQuery(value string) string {
	return strings.TrimPrefix(strings.TrimSpace(value), "/")
}

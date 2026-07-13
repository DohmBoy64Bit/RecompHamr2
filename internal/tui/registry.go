package tui

import "recomphamr2/internal/commands"

func commandsRegistry() []commands.Command {
	return commands.Registry()
}

func pickerItems(kind overlayKind, snapshot Snapshot) []pickerItem {
	var rows []commands.PickerRow
	var intent IntentKind
	switch kind {
	case overlayModels:
		rows, intent = commands.ModelPickerRows(snapshot.Env), IntentModel
	case overlaySkills:
		rows, intent = commands.SkillPickerRows(snapshot.Env), IntentSkill
	case overlayMCP:
		rows, intent = commands.MCPPickerRows(snapshot.Env), IntentMCP
	case overlayHelp:
		intent = IntentCommand
		for _, command := range commands.Registry() {
			rows = append(rows, commands.PickerRow{Name: command.Name, Summary: command.Summary, Detail: command.Usage})
		}
	default:
		return nil
	}
	items := make([]pickerItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, pickerItem{name: row.Name, description: row.Summary, detail: row.Detail, kind: intent, blocked: row.Blocked})
	}
	return items
}

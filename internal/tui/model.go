package tui

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/colorprofile"

	"recomphamr2/internal/commands"
	"recomphamr2/internal/security"
)

// New returns an empty terminal shell model.
func New(env commands.Environment) Model {
	return Model{Env: env, Layout: DefaultLayout(), HistoryIndex: 0}
}

// DefaultLayout returns RecompHamr's evidence-first terminal layout defaults.
func DefaultLayout() Layout {
	layout := Layout{
		ColorProfile:  colorprofile.ANSI256,
		Width:         DefaultWidth,
		Height:        DefaultHeight,
		Mode:          "plan",
		ActiveModel:   "unverified",
		ActiveSkill:   "none",
		MCPStatus:     "gated",
		ContextStatus: "local budget pending",
		PendingTool:   "none",
		MemoryStatus:  "refreshed",
	}
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		layout.ColorProfile = colorprofile.ASCII
	}
	return layout
}

// Update applies one terminal event and returns the requested side effect.
func (m Model) Update(event Event) (Model, Action) {
	if event.Width > 0 {
		m.Layout.Width = event.Width
	}
	if event.Height > 0 {
		m.Layout.Height = event.Height
	}
	if event.Paste != "" {
		m = m.Paste(event.Paste)
	}
	if event.Text != "" {
		m.Composer += event.Text
		m.QuitArmed = false
		m.PaletteIndex = 0
	}
	if event.Key == "" {
		return m, ActionNone
	}
	return m.handleKey(event.Key)
}

// Submit dispatches slash commands or appends plain user text.
func (m Model) Submit(text string) Model {
	text = strings.TrimSpace(text)
	if text == "" && len(m.Attachments) == 0 {
		return m
	}
	m.History = append(m.History, text)
	m.HistoryIndex = len(m.History)
	if strings.HasPrefix(text, "/") {
		out, env := commands.Execute(m.Env, text)
		m.Env = env
		m = m.appendTranscript(out)
		m.Attachments = nil
		m.Composer = ""
		return m
	}
	m = m.appendTranscript("user: " + submissionText(text, m.Attachments))
	m.Attachments = nil
	m.Composer = ""
	return m
}

// Paste inserts small single-line text or creates a large-paste chip.
func (m Model) Paste(text string) Model {
	if isLargePaste(text) {
		name := fmt.Sprintf("paste-%d", len(m.Attachments)+1)
		m.Attachments = append(m.Attachments, Attachment{Name: name, Content: text})
		m = m.appendTranscript(fmt.Sprintf("paste: %s (%d bytes)", name, len(text)))
		return m
	}
	m.Composer += text
	return m
}

// Debug records a redacted debug line when debug mode is enabled.
func (m Model) Debug(text string) Model {
	if !m.DebugEnabled {
		return m
	}
	m.DebugLog = append(m.DebugLog, redact(text, m.DebugSecrets))
	return m
}

// Improvements documents intentional differences from OpenCode-style terminal agents.
func Improvements() []string {
	return []string{
		"evidence rail keeps memory, skill, MCP, and tool state visible for reverse-engineering work",
		"right-side evidence deck separates verified context from chat transcript to reduce claim drift",
		"compact mode collapses panels into status bands so narrow terminals remain usable",
		"RecompHamr-owned visual tokens keep the UI distinct from OpenCode while preserving terminal polish",
	}
}

func (m Model) handleKey(key string) (Model, Action) {
	switch key {
	case KeyEnter:
		before := len(m.Transcript)
		m = m.Submit(m.Composer)
		if len(m.Transcript) == before {
			return m, ActionNone
		}
		return m, ActionSubmit
	case KeyBackspace:
		m.Composer = trimLastRune(m.Composer)
		return m, ActionNone
	case KeyUp:
		if rows := m.overlayRows(); len(rows) > 0 {
			m.PaletteIndex--
			if m.PaletteIndex < 0 {
				m.PaletteIndex = 0
			}
			return m, ActionNone
		}
		m = m.recall(-1)
		return m, ActionNone
	case KeyDown:
		if rows := m.overlayRows(); len(rows) > 0 {
			m.PaletteIndex++
			if m.PaletteIndex >= len(rows) {
				m.PaletteIndex = len(rows) - 1
			}
			return m, ActionNone
		}
		m = m.recall(1)
		return m, ActionNone
	case KeyTab:
		m = m.completeComposer()
		return m, ActionNone
	case KeyCtrlC:
		return m.ctrlC()
	case KeyCtrlD:
		m.Status = "quit"
		return m, ActionQuit
	case KeyEsc:
		m.QuitArmed = false
		m.Status = ""
		return m, ActionNone
	default:
		m.Status = "unsupported key: " + key
		return m, ActionNone
	}
}

func (m Model) completeComposer() Model {
	text := strings.TrimSpace(m.Composer)
	if !strings.HasPrefix(text, "/") {
		return m
	}
	fields := strings.Fields(text)
	prefix := text
	suffix := ""
	if len(fields) > 0 {
		prefix = fields[0]
		if len(fields) > 1 {
			suffix = " " + strings.Join(fields[1:], " ")
		}
	}
	matches := CompleteCommand(prefix)
	if len(matches) == 0 {
		m.Status = "unverified: no command matches " + prefix
		return m
	}
	index := m.PaletteIndex
	if index < 0 || index >= len(matches) {
		index = 0
	}
	m.Composer = matches[index] + suffix
	if suffix == "" {
		m.Composer += " "
	}
	m.Status = "completed command: " + matches[index]
	return m
}

func (m Model) ctrlC() (Model, Action) {
	if m.Layout.PendingTool != "none" || m.Layout.Mode == "thinking" || m.Layout.Mode == "streaming" {
		m.Layout.PendingTool = "none"
		m.Layout.Mode = "idle"
		m.Status = "cancelled"
		m = m.appendTranscript("status: cancelled")
		m.QuitArmed = false
		return m, ActionCancel
	}
	if m.QuitArmed {
		m.Status = "quit"
		return m, ActionQuit
	}
	m.QuitArmed = true
	m.Status = "press Ctrl+C again to quit"
	return m, ActionNone
}

func (m Model) recall(delta int) Model {
	if len(m.History) == 0 {
		return m
	}
	next := m.HistoryIndex + delta
	if next < 0 {
		next = 0
	}
	if next > len(m.History) {
		next = len(m.History)
	}
	m.HistoryIndex = next
	if next == len(m.History) {
		m.Composer = ""
		return m
	}
	m.Composer = m.History[next]
	return m
}

func trimLastRune(text string) string {
	if text == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(text)
	return text[:len(text)-size]
}

func (m Model) appendTranscript(lines ...string) Model {
	if len(lines) == 0 {
		return m
	}
	if m.TranscriptOffset > 0 {
		m.TranscriptOffset += len(lines)
		m.NewOutput = true
	}
	m.Transcript = append(m.Transcript, lines...)
	return m
}

// AppendRuntimeTranscript appends app-owned output while preserving transcript follow state.
func (m Model) AppendRuntimeTranscript(lines ...string) Model {
	return m.appendTranscript(lines...)
}

func (m Model) scrollTranscript(delta int) Model {
	m.TranscriptOffset += delta
	if m.TranscriptOffset < 0 {
		m.TranscriptOffset = 0
	}
	maximum := len(m.Transcript) - 1
	if maximum < 0 {
		maximum = 0
	}
	if m.TranscriptOffset > maximum {
		m.TranscriptOffset = maximum
	}
	if m.TranscriptOffset == 0 {
		m.NewOutput = false
	}
	return m
}

func redact(text string, secrets []string) string {
	out := text
	for _, secret := range secrets {
		out = security.RedactSecret(out, secret)
	}
	return out
}

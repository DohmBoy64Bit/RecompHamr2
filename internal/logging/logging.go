package logging

import (
	"fmt"
	"strings"

	"recomphamr2/internal/security"
)

// Level identifies the severity of a diagnostic log entry.
type Level string

const (
	// Debug records verbose diagnostic detail.
	Debug Level = "debug"
	// Info records normal progress.
	Info Level = "info"
	// Warn records recoverable risk.
	Warn Level = "warn"
	// Error records failed operations.
	Error Level = "error"
)

// Entry is one redacted diagnostic log record.
type Entry struct {
	Level   Level
	Message string
	Fields  map[string]string
}

// Redactor applies stable secret redaction to log output.
type Redactor struct {
	Secrets []string
}

// NewRedactor returns a redactor configured with secret values to remove.
func NewRedactor(secrets ...string) Redactor {
	return Redactor{Secrets: append([]string(nil), secrets...)}
}

// Redact removes all configured secrets from text.
func (r Redactor) Redact(text string) string {
	out := text
	for _, secret := range r.Secrets {
		out = security.RedactSecret(out, secret)
	}
	return out
}

// Format renders an entry as one deterministic log line.
func (r Redactor) Format(entry Entry) string {
	level := entry.Level
	if level == "" {
		level = Info
	}
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s", level, r.Redact(entry.Message))
	if len(entry.Fields) > 0 {
		keys := sortedKeys(entry.Fields)
		for _, key := range keys {
			fmt.Fprintf(&b, " %s=%s", key, r.Redact(entry.Fields[key]))
		}
	}
	return b.String()
}

func sortedKeys(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

package testharness

import (
	"strings"
)

// NormalizeLines converts platform-specific line endings and trims final space.
func NormalizeLines(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// GoldenMismatch returns a readable mismatch message or an empty string when equal.
func GoldenMismatch(name string, got string, want string) string {
	got = NormalizeLines(got)
	want = NormalizeLines(want)
	if got == want {
		return ""
	}
	return name + " mismatch\n--- got ---\n" + got + "\n--- want ---\n" + want
}

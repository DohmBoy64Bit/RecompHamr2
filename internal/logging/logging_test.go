package logging

import (
	"strings"
	"testing"
)

func TestFormatRedactsAndSortsFields(t *testing.T) {
	redactor := NewRedactor("secret")
	got := redactor.Format(Entry{
		Level:   Warn,
		Message: "token secret",
		Fields:  map[string]string{"z": "last", "a": "secret"},
	})
	want := "[warn] token [REDACTED] a=[REDACTED] z=last"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatDefaultsToInfoAndCopiesSecrets(t *testing.T) {
	secrets := []string{"one"}
	redactor := NewRedactor(secrets...)
	secrets[0] = "two"
	got := redactor.Format(Entry{Message: "one two"})
	if got != "[info] [REDACTED] two" {
		t.Fatalf("Format() = %q, want copied secret redaction", got)
	}
}

func TestRedactNoSecretsAndLevels(t *testing.T) {
	redactor := NewRedactor()
	if got := redactor.Redact("plain"); got != "plain" {
		t.Fatalf("Redact() = %q, want plain", got)
	}
	for _, level := range []Level{Debug, Info, Warn, Error} {
		if !strings.Contains(redactor.Format(Entry{Level: level, Message: "m"}), string(level)) {
			t.Fatalf("level %q missing from formatted output", level)
		}
	}
}

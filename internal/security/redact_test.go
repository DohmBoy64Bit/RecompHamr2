package security

import "testing"

func TestRedactSecret(t *testing.T) {
	got := RedactSecret("token=abc123", "abc123")
	want := "token=[REDACTED]"
	if got != want {
		t.Fatalf("RedactSecret() = %q, want %q", got, want)
	}
}

func TestRedactSecretEmptySecret(t *testing.T) {
	got := RedactSecret("token=abc123", "")
	want := "token=abc123"
	if got != want {
		t.Fatalf("RedactSecret() = %q, want %q", got, want)
	}
}

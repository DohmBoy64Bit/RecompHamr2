package testharness

import (
	"strings"
	"testing"
)

func TestNormalizeLines(t *testing.T) {
	got := NormalizeLines("a \r\nb\t\r\n")
	want := "a\nb"
	if got != want {
		t.Fatalf("NormalizeLines() = %q, want %q", got, want)
	}
}

func TestGoldenMismatch(t *testing.T) {
	if got := GoldenMismatch("x", "a\r\n", "a\n"); got != "" {
		t.Fatalf("GoldenMismatch() = %q, want empty", got)
	}
	got := GoldenMismatch("x", "a", "b")
	for _, want := range []string{"x mismatch", "--- got ---", "a", "--- want ---", "b"} {
		if !strings.Contains(got, want) {
			t.Fatalf("GoldenMismatch() missing %q: %q", want, got)
		}
	}
}

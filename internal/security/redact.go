package security

import "strings"

// RedactSecret replaces a known secret value in text with a stable redaction marker.
func RedactSecret(text string, secret string) string {
	if secret == "" {
		return text
	}
	return strings.ReplaceAll(text, secret, "[REDACTED]")
}

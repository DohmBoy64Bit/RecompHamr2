// Package tools implements RecompHamr built-in LLM tools.
//
// The primary shell tool is powershell because RecompHamr 2.0 is Windows-first.
// The bash name remains as a RecompHamr 1.x compatibility alias and maps to
// PowerShell on Windows. File and network tools return explicit tool-style
// failure strings so callers can nudge the model without inventing success.
package tools

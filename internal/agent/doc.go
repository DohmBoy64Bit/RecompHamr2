// Package agent implements a deterministic model-tool loop independent from the TUI.
//
// The loop owns assistant/tool transcript pairing, cancellation checks,
// repeated-failure nudges, runaway-tool nudges, empty-reply retries, and the
// verification nudge used before a substantial clean finish. User interfaces
// render the resulting transcript; they do not duplicate this turn policy.
package agent

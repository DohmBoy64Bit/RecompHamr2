// Package commands implements the RecompHamr slash command registry.
//
// The registry is the source of truth for command help, examples, side effects,
// and expected error classes. Execution stays independent from the TUI so
// command behavior can be tested without launching a terminal.
package commands

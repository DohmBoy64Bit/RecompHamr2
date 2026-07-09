// Package app wires the RecompHamr process entrypoint to supported runtime modes.
//
// The default runtime composes local config, optional project memory, command
// environment, MCP manager state, and pure TUI state without making network
// model calls or autoconnecting MCP servers.
package app

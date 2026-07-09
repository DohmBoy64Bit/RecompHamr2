# Documentation Coverage

Documentation coverage is 100% only when every user-visible and developer-visible item is linked from `.docs-index.md` and explained in the appropriate user and developer docs.

Coverage includes commands, flags, environment variables, config keys, generated files, tools, tool parameters, MCP servers, skill-loading rules, `.rehamr/` files, security boundaries, install paths, release paths, actionable errors, and exported Go symbols.

Next 20 Phase 21 upgrades `make docscheck` so it mechanically verifies:

- required durable memory docs exist and are non-empty;
- every slash command name appears in command docs;
- built-in tool names and parameters appear in tool docs;
- config keys and `RECOMPHAMR_URL` appear in config docs;
- MCP built-ins and environment override names appear in MCP docs;
- generated `.rehamr/` files appear in memory docs;
- install, release, devcontainer, CI, and checksum files appear in install or
  release docs;
- command-line help flags are documented; and
- exported Go package, type, value, and function docs are present.

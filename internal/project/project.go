package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"recomphamr2/internal/config"
)

var (
	mkdirAll  = os.MkdirAll
	lstat     = os.Lstat
	readFile  = os.ReadFile
	writeFile = os.WriteFile
	stat      = os.Stat
)

const (
	// MemoryFileName is the persistent workspace memory file.
	MemoryFileName = "REPHAMR_STATE.md"
	// DefaultMemoryMaxBytes is the default safe read cap for project memory.
	DefaultMemoryMaxBytes = 24 * 1024
)

var (
	// ErrWorkspaceMissing reports that .rehamr has not been initialized.
	ErrWorkspaceMissing = errors.New(".rehamr workspace is not initialized")
	// ErrMemoryMissing reports that REPHAMR_STATE.md is absent.
	ErrMemoryMissing = errors.New("REPHAMR_STATE.md is missing")
)

// Workspace describes the .rehamr project state paths.
type Workspace struct {
	// Root is the absolute .rehamr workspace directory.
	Root string
	// Config is the .rehamr/config.yaml path.
	Config string
	// State is the persistent project memory file path.
	State string
	// MCPConfig is the .rehamr/mcp.json path.
	MCPConfig string
	// Evidence is the directory for confirmed evidence notes.
	Evidence string
	// Repos is the repository cache directory.
	Repos string
	// Skills is the custom and active skill directory.
	Skills string
	// Logs is the diagnostic log directory.
	Logs string
	// Functions is the function ledger directory.
	Functions string
	// Formats is the file-format evidence directory.
	Formats string
	// Recomp is the static recompilation evidence directory.
	Recomp string
	// Decomp is the decompilation evidence directory.
	Decomp string
	// Project is the project summary file path.
	Project string
	// Blockers is the blocker ledger file path.
	Blockers string
	// Changelog is the workspace changelog file path.
	Changelog string
	// Hypotheses is the unverified hypothesis file path.
	Hypotheses string
}

// Memory contains REPHAMR_STATE.md content prepared for runtime prompt use.
type Memory struct {
	// Path is the absolute memory file path.
	Path string
	// Content is the UTF-8-safe memory file content.
	Content string
	// Truncated reports whether Content was capped before return.
	Truncated bool
	// MaxBytes is the byte cap used while loading memory.
	MaxBytes int
}

// Init creates the RecompHamr project workspace safely and idempotently.
func Init(projectDir string) (Workspace, error) {
	cfg, _, err := config.Bootstrap(projectDir)
	if err != nil {
		return Workspace{}, err
	}
	root := cfg.Dir
	ws := Workspace{
		Root:       root,
		Config:     filepath.Join(root, config.FileName),
		State:      filepath.Join(root, MemoryFileName),
		MCPConfig:  filepath.Join(root, "mcp.json"),
		Evidence:   filepath.Join(root, "evidence"),
		Repos:      filepath.Join(root, "repos"),
		Skills:     filepath.Join(root, "skills"),
		Logs:       filepath.Join(root, "logs"),
		Functions:  filepath.Join(root, "functions"),
		Formats:    filepath.Join(root, "formats"),
		Recomp:     filepath.Join(root, "recomp"),
		Decomp:     filepath.Join(root, "decomp"),
		Project:    filepath.Join(root, "PROJECT.md"),
		Blockers:   filepath.Join(root, "BLOCKERS.md"),
		Changelog:  filepath.Join(root, "CHANGELOG.md"),
		Hypotheses: filepath.Join(root, "HYPOTHESES.md"),
	}
	for _, dir := range workspaceDirs(ws) {
		if err := mkdirAll(dir, 0o700); err != nil {
			return Workspace{}, err
		}
	}
	for rel, body := range workspaceFiles(time.Now()) {
		if err := writeIfMissing(filepath.Join(root, rel), body); err != nil {
			return Workspace{}, err
		}
	}
	return ws, nil
}

// LoadMemory reads .rehamr/REPHAMR_STATE.md for model context injection.
func LoadMemory(projectDir string, maxBytes int) (Memory, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMemoryMaxBytes
	}
	root := filepath.Join(projectDir, config.DirName)
	info, err := lstat(root)
	if err != nil {
		return Memory{Path: filepath.Join(root, MemoryFileName), MaxBytes: maxBytes}, ErrWorkspaceMissing
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return Memory{Path: filepath.Join(root, MemoryFileName), MaxBytes: maxBytes}, fmt.Errorf("%s must be a real directory", config.DirName)
	}
	path := filepath.Join(root, MemoryFileName)
	if info, err := lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return Memory{Path: path, MaxBytes: maxBytes}, fmt.Errorf("%s must not be a symlink", MemoryFileName)
	}
	data, err := readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Memory{Path: path, MaxBytes: maxBytes}, ErrMemoryMissing
		}
		return Memory{Path: path, MaxBytes: maxBytes}, err
	}
	content := string(data)
	mem := Memory{Path: path, Content: content, MaxBytes: maxBytes}
	if len(content) > maxBytes {
		mem.Content = truncateUTF8(content, maxBytes)
		mem.Truncated = true
	}
	return mem, nil
}

// Status returns a concise workspace status string.
func Status(projectDir string) string {
	root := filepath.Join(projectDir, config.DirName)
	info, err := lstat(root)
	if err != nil {
		return "unsupported: .rehamr workspace is not initialized"
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "blocked: .rehamr must be a real directory"
	}
	var b strings.Builder
	b.WriteString("RecompHamr project status\n")
	b.WriteString("=========================\n")
	fmt.Fprintf(&b, "root: %s\n", root)
	for _, rel := range statusFiles() {
		full := filepath.Join(root, rel)
		data, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(&b, "\n## %s\nmissing\n", rel)
			continue
		}
		text := strings.TrimSpace(string(data))
		if len(text) > 1800 {
			text = truncateUTF8(text, 1800) + "\n...truncated..."
		}
		fmt.Fprintf(&b, "\n## %s\n%s\n", rel, text)
	}
	return b.String()
}

func writeIfMissing(path string, content string) error {
	if _, err := stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := mkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return writeFile(path, []byte(content), 0o600)
}

func workspaceDirs(ws Workspace) []string {
	return []string{
		ws.Evidence,
		ws.Repos,
		ws.Skills,
		ws.Logs,
		ws.Functions,
		filepath.Join(ws.Formats, "parsers"),
		filepath.Join(ws.Formats, "tests"),
		ws.Recomp,
		ws.Decomp,
	}
}

func workspaceFiles(now time.Time) map[string]string {
	date := now.Format("2006-01-02")
	return map[string]string{
		"PROJECT.md":                    "# Project\n\nRecord the target, source of truth, goal, and toolchain decisions with evidence links.\n",
		MemoryFileName:                  stateTemplate(date),
		"EVIDENCE.md":                   "# Evidence\n\nAdd confirmed facts only. Cite files, commands, logs, symbols, offsets, or test output.\n",
		"HYPOTHESES.md":                 "# Hypotheses\n\nTrack unconfirmed ideas here until evidence promotes or disproves them.\n",
		"BLOCKERS.md":                   "# Blockers\n\nTrack missing files, tools, build failures, runtime blockers, and user decisions.\n",
		"CHANGELOG.md":                  fmt.Sprintf("# Changelog\n\n## %s\n- Initialized RecompHamr evidence workspace.\n", date),
		"COMMANDS.md":                   "# Commands\n\nRecord commands that work and the evidence each command produced.\n",
		"TOOLCHAIN.md":                  "# Toolchain\n\nDocument compilers, SDKs, decompilers, scripts, versions, and verification commands.\n",
		"MODELS.md":                     "# Models\n\nRecord model profiles, context settings, endpoint behavior, and observed limits.\n",
		"functions/inventory.csv":       "address_or_symbol,name,status,classification,evidence_source,confidence,notes\n",
		"functions/game_logic.md":       "# Game Or Project Logic Functions\n\nRecord behavior-relevant functions with evidence.\n",
		"functions/runtime_platform.md": "# Runtime Platform Functions\n\nRecord runtime, middleware, platform, and system interface functions.\n",
		"functions/unknown.md":          "# Unknown Functions\n\nRecord unidentified functions and the evidence needed to classify them.\n",
		"formats/inventory.md":          "# File Format Inventory\n\nRecord discovered formats, samples, magic values, parsers, and evidence.\n",
		"formats/hypotheses.md":         "# File Format Hypotheses\n\nRecord unconfirmed format theories with evidence requirements.\n",
		"recomp/bridge_audit.md":        "# Bridge And Stub Audit\n\nRecord host bridge, thunk, syscall, and stub status with evidence.\n",
		"recomp/runtime_gaps.md":        "# Runtime Gaps\n\nRecord missing runtime behavior and the tests or traces that expose each gap.\n",
		"recomp/thread_trace.md":        "# Thread And Message Trace\n\nRecord threading, scheduling, and message-loop evidence.\n",
		"recomp/build_matrix.md":        "# Build Matrix\n\nRecord build commands, configurations, platforms, and verified outcomes.\n",
		"decomp/compiler_detection.md":  "# Compiler Detection\n\nRecord compiler, linker, ABI, and optimization evidence.\n",
		"decomp/matching_status.md":     "# Matching Status\n\nRecord match status, deltas, and verification commands.\n",
		"decomp/symbols.md":             "# Symbols\n\nRecord symbol names, addresses, confidence, and evidence sources.\n",
		"skills/active.md":              "# Active Skills\n\nRecord loaded skills and why each one applies.\n",
		"mcp.json":                      "{\n  \"servers\": {}\n}\n",
	}
}

func stateTemplate(date string) string {
	return `# RecompHamr Project State
> Maintained by agents and humans. Read it before project-changing work.

## Quick Rules
1. Evidence first: classify, cite, and save evidence under .rehamr/evidence/.
2. Fix source metadata, config, or runtime code before generated output.
3. Verify paths before using them.
4. Read files before acting on their contents.
5. Command output is evidence; preserve the exact command and result.

## Current Phase
- Track:
- Phase:
- Current goal:

## Project Info
- Project or target:
- Goal:
- Source of truth:
- Toolchain:

## Workspace Paths
- Project root:
- Evidence dir: .rehamr/evidence/
- Repos cache: .rehamr/repos/
- Key reference files:

## Active Commands
` + "```sh\n# Record commands that work with their purpose.\n```" + `

## Blockers
| Issue | Status | Evidence |
|---|---|---|

## Function And Symbol Ledger
| Name | Address | Classification | Confidence | Source |
|---|---|---|---|---|

## Learned Patterns
- 

## Session Log
| Date | Summary |
|---|---|
| ` + date + ` | Initialized by /init-re |
`
}

func statusFiles() []string {
	return []string{
		"PROJECT.md",
		MemoryFileName,
		"EVIDENCE.md",
		"HYPOTHESES.md",
		"BLOCKERS.md",
		"CHANGELOG.md",
		"functions/inventory.csv",
		"formats/inventory.md",
		"recomp/runtime_gaps.md",
	}
}

func truncateUTF8(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	for maxBytes > 0 && maxBytes < len(text) && !utf8.RuneStart(text[maxBytes]) {
		maxBytes--
	}
	return text[:maxBytes]
}

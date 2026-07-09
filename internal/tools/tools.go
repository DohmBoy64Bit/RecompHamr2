package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	mkdirAll       = os.MkdirAll
	writeFile      = os.WriteFile
	readFile       = os.ReadFile
	commandContext = exec.CommandContext
	runtimeGOOS    = runtime.GOOS
	filepathAbs    = filepath.Abs
	filepathRel    = filepath.Rel
)

const (
	// PowerShellName is the Windows-first shell command tool.
	PowerShellName = "powershell"
	// BashName is the 1.x compatibility shell command alias.
	BashName = "bash"
	// ReadFileName is the file read tool.
	ReadFileName = "read_file"
	// WriteFileName is the file write tool.
	WriteFileName = "write_file"
	// EditFileName is the exact replacement tool.
	EditFileName = "edit_file"
	// RepomixrName is the repository reference packing tool.
	RepomixrName = "repomixr"
	// RecompReferenceName is the webpage cache tool.
	RecompReferenceName = "recomp_reference"

	// DefaultShellTimeout is used when a shell tool call omits a timeout.
	DefaultShellTimeout = 120 * time.Second
	// MaxShellTimeout prevents runaway shell commands from exceeding the 1.x cap.
	MaxShellTimeout = time.Hour
	// MaxReadFileBytes limits read_file output so a single tool result cannot flood context.
	MaxReadFileBytes = 1 << 20
)

// Schema describes a tool and its documented parameters.
type Schema struct {
	Name        string
	Description string
	Parameters  []string
}

// Schemas returns the six primary built-in tool schemas.
func Schemas() []Schema {
	return []Schema{
		{PowerShellName, "Run a Windows PowerShell command with timeout and cancellation.", []string{"cmd", "timeout_seconds"}},
		{ReadFileName, "Read a file from disk.", []string{"path"}},
		{WriteFileName, "Write content bytes to a file, creating parents.", []string{"path", "content"}},
		{EditFileName, "Replace exactly one old_string occurrence.", []string{"path", "old_string", "new_string"}},
		{RepomixrName, "Clone a GitHub repository into .rehamr/repos.", []string{"url", "output_dir"}},
		{RecompReferenceName, "Fetch and cache a web reference for offline reading.", []string{"url", "output_dir"}},
	}
}

// CompatibilityToolNames returns supported legacy tool names that are not primary schemas.
func CompatibilityToolNames() []string {
	return []string{BashName}
}

// PowerShell runs a Windows PowerShell command with bounded timeout and cancellation.
func PowerShell(ctx context.Context, command string, timeout time.Duration) string {
	return runShell(ctx, PowerShellName, command, timeout)
}

// Bash runs the legacy shell alias for 1.x parity; on Windows it maps to PowerShell.
func Bash(ctx context.Context, command string, timeout time.Duration) string {
	return runShell(ctx, BashName, command, timeout)
}

func runShell(ctx context.Context, toolName string, command string, timeout time.Duration) string {
	if ctx.Err() != nil {
		return fmt.Sprintf("(%s: cancelled)", toolName)
	}
	if strings.TrimSpace(command) == "" {
		return fmt.Sprintf("(%s: empty command)", toolName)
	}
	timeout = normalizeShellTimeout(timeout)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var cmd *exec.Cmd
	if toolName == PowerShellName || runtimeGOOS == "windows" {
		exe := "powershell"
		if runtimeGOOS != "windows" {
			exe = "pwsh"
		}
		cmd = commandContext(ctx, exe, "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		cmd = commandContext(ctx, "/bin/sh", "-c", command)
	}
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\r\n")
	if ctx.Err() == context.DeadlineExceeded {
		return appendToolLine(text, "(timeout after "+timeout.String()+")")
	}
	if ctx.Err() == context.Canceled {
		return appendToolLine(text, "(cancelled)")
	}
	if err != nil {
		return appendToolLine(text, "(exit: "+err.Error()+")")
	}
	return string(out)
}

func normalizeShellTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return DefaultShellTimeout
	}
	if timeout > MaxShellTimeout {
		return MaxShellTimeout
	}
	return timeout
}

func appendToolLine(text string, line string) string {
	if text == "" {
		return line
	}
	return text + "\n" + line
}

// ReadFile reads path and returns content or a tool-style error string.
func ReadFile(path string) string {
	if path == "" {
		return "(read_file: empty path)"
	}
	data, err := readFile(path)
	if err != nil {
		return fmt.Sprintf("(read_file: %v)", err)
	}
	return truncateReadFile(data)
}

// WriteFile writes content to path and creates parents.
func WriteFile(path string, content string) string {
	if path == "" {
		return "(write_file: empty path)"
	}
	if err := mkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Sprintf("(write_file: mkdir: %v)", err)
	}
	if err := writeFile(path, []byte(content), 0o600); err != nil {
		return fmt.Sprintf("(write_file: %v)", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), path)
}

// EditFile replaces exactly one occurrence in path.
func EditFile(path string, oldString string, newString string) string {
	if strings.TrimSpace(path) == "" {
		return "(edit_file: empty path)"
	}
	if oldString == "" {
		return "(edit_file: old_string not found)"
	}
	if oldString == newString {
		return "(edit_file: old_string and new_string must differ)"
	}
	data, err := readFile(path)
	if err != nil {
		return fmt.Sprintf("(edit_file: %v)", err)
	}
	text := string(data)
	count := strings.Count(text, oldString)
	if count == 0 {
		return "(edit_file: old_string not found)"
	}
	if count > 1 {
		return "(edit_file: old_string is not unique)"
	}
	updated := strings.Replace(text, oldString, newString, 1)
	if err := writeFile(path, []byte(updated), 0o600); err != nil {
		return fmt.Sprintf("(edit_file: %v)", err)
	}
	return fmt.Sprintf("edited %s", path)
}

// Repomixr clones a GitHub repository into outputDir.
func Repomixr(ctx context.Context, rawURL string, outputDir string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "(repomixr: invalid URL)"
	}
	if parsed.Host != "github.com" {
		return "(repomixr: only github.com URLs are supported)"
	}
	if outputDir == "" {
		return "(repomixr: output directory not configured)"
	}
	name, ok := githubRepoName(parsed.Path)
	if !ok {
		return "(repomixr: invalid owner/repo name)"
	}
	target, err := joinInside(outputDir, name)
	if err != nil {
		return fmt.Sprintf("(repomixr: %v)", err)
	}
	if err := mkdirAll(outputDir, 0o700); err != nil {
		return fmt.Sprintf("(repomixr: mkdir: %v)", err)
	}
	cmd := commandContext(ctx, "git", "clone", "--depth=1", rawURL, target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Sprintf("(repomixr: git clone: %v: %s)", err, strings.TrimSpace(string(out)))
	}
	return fmt.Sprintf("cloned %s to %s", rawURL, target)
}

// RecompReference fetches rawURL and writes it into outputDir.
func RecompReference(ctx context.Context, client *http.Client, rawURL string, outputDir string) string {
	if client == nil {
		client = http.DefaultClient
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "(recomp_reference: invalid URL)"
	}
	if outputDir == "" {
		return "(recomp_reference: output directory not configured)"
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("(recomp_reference: fetch failed: %v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Sprintf("(recomp_reference: HTTP %d from %s)", resp.StatusCode, parsed.Host)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("(recomp_reference: read: %v)", err)
	}
	name := cacheFileName(parsed.Host + parsed.Path)
	if name == "" {
		return "(recomp_reference: invalid cache name)"
	}
	if err := mkdirAll(outputDir, 0o700); err != nil {
		return fmt.Sprintf("(recomp_reference: mkdir: %v)", err)
	}
	path, err := joinInside(outputDir, name+".txt")
	if err != nil {
		return fmt.Sprintf("(recomp_reference: %v)", err)
	}
	if err := writeFile(path, body, 0o600); err != nil {
		return fmt.Sprintf("(recomp_reference: write: %v)", err)
	}
	return fmt.Sprintf("fetched %s to %s", rawURL, path)
}

// IsFailure reports whether tool output follows RecompHamr error conventions.
func IsFailure(output string) bool {
	trimmed := strings.TrimSpace(output)
	return strings.HasPrefix(trimmed, "(") || strings.Contains(trimmed, "\n(exit: ") || strings.Contains(trimmed, "\n(timeout after ")
}

// Unsupported returns a consistent unsupported result for not-yet-wired tool names.
func Unsupported(name string) error {
	return errors.New("unsupported tool: " + name)
}

func truncateReadFile(data []byte) string {
	if len(data) <= MaxReadFileBytes {
		return string(data)
	}
	return string(data[:MaxReadFileBytes]) + fmt.Sprintf("\n(read_file: truncated after %d bytes)", MaxReadFileBytes)
}

func githubRepoName(path string) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return "", false
	}
	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSuffix(strings.TrimSpace(parts[1]), ".git")
	if !safePathSegment(owner) || !safePathSegment(repo) {
		return "", false
	}
	return owner + "-" + repo, true
}

func safePathSegment(segment string) bool {
	return segment != "" && segment != "." && segment != ".." && !strings.ContainsAny(segment, `/\:`)
}

func cacheFileName(name string) string {
	name = strings.NewReplacer("/", "-", "\\", "-", ":", "-", "..", "-").Replace(name)
	return strings.Trim(name, "-. ")
}

func joinInside(base string, elems ...string) (string, error) {
	absBase, err := filepathAbs(base)
	if err != nil {
		return "", err
	}
	path := filepath.Join(append([]string{absBase}, elems...)...)
	rel, err := filepathRel(absBase, path)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes output directory")
	}
	return path, nil
}

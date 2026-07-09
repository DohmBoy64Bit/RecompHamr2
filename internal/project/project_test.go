package project

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"recomphamr2/internal/config"
)

func TestInitCreatesWorkspace(t *testing.T) {
	ws, err := Init(t.TempDir())
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	for _, path := range []string{
		ws.Root, ws.Evidence, ws.Repos, ws.Skills, ws.Logs, ws.Functions,
		filepath.Join(ws.Formats, "parsers"), filepath.Join(ws.Formats, "tests"),
		ws.Recomp, ws.Decomp,
	} {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("directory %s missing or not dir", path)
		}
	}
	for _, path := range []string{
		ws.Config, ws.State, ws.MCPConfig, ws.Project, ws.Blockers,
		ws.Changelog, ws.Hypotheses, filepath.Join(ws.Functions, "inventory.csv"),
		filepath.Join(ws.Formats, "inventory.md"), filepath.Join(ws.Recomp, "runtime_gaps.md"),
		filepath.Join(ws.Decomp, "symbols.md"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("file %s missing: %v", path, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
			t.Fatalf("file %s perms = %v, want 0600", path, info.Mode().Perm())
		}
	}
}

func TestInitIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	ws, err := Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.WriteFile(ws.State, []byte("custom"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Init(dir); err != nil {
		t.Fatalf("Init() second error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".rehamr", "REPHAMR_STATE.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "custom" {
		t.Fatalf("state overwritten: %q", data)
	}
}

func TestStatus(t *testing.T) {
	dir := t.TempDir()
	if got := Status(dir); !strings.HasPrefix(got, "unsupported:") {
		t.Fatalf("Status() before init = %q", got)
	}
	if _, err := Init(dir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	got := Status(dir)
	for _, want := range []string{"RecompHamr project status", "PROJECT.md", "REPHAMR_STATE.md", "EVIDENCE.md", "functions/inventory.csv"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Status() missing %q:\n%s", want, got)
		}
	}
}

func TestLoadMemory(t *testing.T) {
	dir := t.TempDir()
	mem, err := LoadMemory(dir, 0)
	if !errors.Is(err, ErrWorkspaceMissing) {
		t.Fatalf("LoadMemory() before init error = %v, want ErrWorkspaceMissing", err)
	}
	if mem.MaxBytes != DefaultMemoryMaxBytes || !strings.HasSuffix(mem.Path, filepath.Join(".rehamr", MemoryFileName)) {
		t.Fatalf("LoadMemory() before init metadata = %#v", mem)
	}
	ws, err := Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	content := "# State\n\nverified fact"
	if err := os.WriteFile(ws.State, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	mem, err = LoadMemory(dir, 1024)
	if err != nil {
		t.Fatalf("LoadMemory() error = %v", err)
	}
	if mem.Path != ws.State || mem.Content != content || mem.Truncated || mem.MaxBytes != 1024 {
		t.Fatalf("LoadMemory() = %#v", mem)
	}
}

func TestLoadMemoryMissingAndTruncatesUTF8(t *testing.T) {
	dir := t.TempDir()
	ws, err := Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.Remove(ws.State); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := LoadMemory(dir, 8); !errors.Is(err, ErrMemoryMissing) {
		t.Fatalf("LoadMemory() missing error = %v, want ErrMemoryMissing", err)
	}
	text := "abcdétail"
	if err := os.WriteFile(ws.State, []byte(text), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	mem, err := LoadMemory(dir, 5)
	if err != nil {
		t.Fatalf("LoadMemory() truncate error = %v", err)
	}
	if !mem.Truncated || mem.Content != "abcd" || !utf8.ValidString(mem.Content) {
		t.Fatalf("LoadMemory() truncation = %#v", mem)
	}
}

func TestLoadMemoryReadError(t *testing.T) {
	dir := t.TempDir()
	if _, err := Init(dir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	origRead := readFile
	defer func() { readFile = origRead }()
	readFile = func(string) ([]byte, error) { return nil, errors.New("read") }
	if _, err := LoadMemory(dir, 0); err == nil || errors.Is(err, ErrMemoryMissing) {
		t.Fatalf("LoadMemory() read error = %v", err)
	}
}

func TestLoadMemoryRefusesSymlinkBoundaries(t *testing.T) {
	origLstat := lstat
	defer func() { lstat = origLstat }()
	dir := t.TempDir()
	root := filepath.Join(dir, config.DirName)
	state := filepath.Join(root, MemoryFileName)

	lstat = func(path string) (os.FileInfo, error) {
		switch path {
		case root:
			return fakeInfo{mode: os.ModeSymlink | os.ModeDir}, nil
		default:
			return origLstat(path)
		}
	}
	if _, err := LoadMemory(dir, 0); err == nil || !strings.Contains(err.Error(), "real directory") {
		t.Fatalf("LoadMemory() symlink root error = %v", err)
	}

	lstat = func(path string) (os.FileInfo, error) {
		switch path {
		case root:
			return fakeInfo{mode: os.ModeDir}, nil
		case state:
			return fakeInfo{mode: os.ModeSymlink}, nil
		default:
			return origLstat(path)
		}
	}
	if _, err := LoadMemory(dir, 0); err == nil || !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("LoadMemory() symlink memory error = %v", err)
	}
}

func TestStatusReportsMissingAndTruncatesUTF8(t *testing.T) {
	dir := t.TempDir()
	ws, err := Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.Remove(ws.Project); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	long := strings.Repeat("a", 1800) + "é-tail"
	if err := os.WriteFile(filepath.Join(ws.Root, "EVIDENCE.md"), []byte(long), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got := Status(dir)
	if !strings.Contains(got, "## PROJECT.md\nmissing") {
		t.Fatalf("Status() did not report missing project:\n%s", got)
	}
	if !strings.Contains(got, "...truncated...") {
		t.Fatalf("Status() did not truncate long evidence:\n%s", got)
	}
	if !utf8.ValidString(got) {
		t.Fatalf("Status() split UTF-8 rune:\n%s", got)
	}
}

func TestWorkspaceTemplatesHaveRequiredSections(t *testing.T) {
	files := workspaceFiles(time.Now())
	state := files["REPHAMR_STATE.md"]
	for _, want := range []string{"Quick Rules", "Current Phase", "Blockers", "Learned Patterns", "Session Log"} {
		if !strings.Contains(state, want) {
			t.Fatalf("state template missing %q:\n%s", want, state)
		}
	}
	for _, rel := range []string{"COMMANDS.md", "TOOLCHAIN.md", "MODELS.md", "skills/active.md", "mcp.json"} {
		if strings.TrimSpace(files[rel]) == "" {
			t.Fatalf("workspace file %s is empty", rel)
		}
	}
	if !strings.Contains(files["mcp.json"], "\"servers\"") {
		t.Fatalf("mcp config missing servers key: %s", files["mcp.json"])
	}
}

func TestWorkspaceDirsAndStatusFiles(t *testing.T) {
	ws := Workspace{Evidence: "e", Repos: "r", Skills: "s", Logs: "l", Functions: "f", Formats: "fmt", Recomp: "rc", Decomp: "dc"}
	if got := workspaceDirs(ws); len(got) != 9 {
		t.Fatalf("workspaceDirs() len = %d, want 9: %#v", len(got), got)
	}
	if got := statusFiles(); len(got) != 9 {
		t.Fatalf("statusFiles() len = %d, want 9: %#v", len(got), got)
	}
}

func TestTruncateUTF8ShortAndASCII(t *testing.T) {
	if got := truncateUTF8("short", 10); got != "short" {
		t.Fatalf("truncateUTF8(short) = %q", got)
	}
	if got := truncateUTF8("abcdef", 3); got != "abc" {
		t.Fatalf("truncateUTF8(ascii) = %q", got)
	}
	if got := truncateUTF8("aéz", 2); got != "a" {
		t.Fatalf("truncateUTF8(mid-rune) = %q", got)
	}
	if got := truncateUTF8("é", 0); got != "" {
		t.Fatalf("truncateUTF8(zero) = %q", got)
	}
}

func TestWriteIfMissingPropagatesStatError(t *testing.T) {
	origStat := stat
	defer func() { stat = origStat }()
	stat = func(string) (os.FileInfo, error) { return nil, errors.New("stat") }
	if err := writeIfMissing(filepath.Join(t.TempDir(), "x"), "content"); err == nil {
		t.Fatal("writeIfMissing() accepted stat error")
	}
}

func TestWriteIfMissingPropagatesMkdirError(t *testing.T) {
	origStat, origMkdir := stat, mkdirAll
	defer func() { stat, mkdirAll = origStat, origMkdir }()
	stat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	mkdirAll = func(string, os.FileMode) error { return errors.New("mkdir") }
	if err := writeIfMissing(filepath.Join(t.TempDir(), "nested", "x"), "content"); err == nil {
		t.Fatal("writeIfMissing() accepted mkdir error")
	}
}

func TestInitFailureSeams(t *testing.T) {
	origMkdir, origRead, origWrite, origStat := mkdirAll, readFile, writeFile, stat
	defer func() { mkdirAll, readFile, writeFile, stat = origMkdir, origRead, origWrite, origStat }()
	mkdirAll = func(string, os.FileMode) error { return errors.New("mkdir") }
	if _, err := Init(t.TempDir()); err == nil {
		t.Fatal("Init() accepted mkdir failure")
	}
	mkdirAll = origMkdir
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write") }
	if _, err := Init(t.TempDir()); err == nil {
		t.Fatal("Init() accepted write failure")
	}
	writeFile = origWrite
	stat = func(string) (os.FileInfo, error) { return nil, nil }
	if err := writeIfMissing(filepath.Join(t.TempDir(), "x"), "content"); err != nil {
		t.Fatalf("writeIfMissing() existing path error = %v", err)
	}
}

func TestInitSecondGeneratedWriteFailure(t *testing.T) {
	origWrite := writeFile
	defer func() { writeFile = origWrite }()
	writeCount := 0
	writeFile = func(path string, content []byte, mode os.FileMode) error {
		writeCount++
		if writeCount == 2 {
			return errors.New("second write")
		}
		return origWrite(path, content, mode)
	}
	if _, err := Init(t.TempDir()); err == nil {
		t.Fatal("Init() accepted second generated write failure")
	}
}

func TestInitPropagatesConfigBootstrapError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".rehamr"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Init(dir); err == nil {
		t.Fatal("Init() accepted config bootstrap failure")
	}
}

func TestStatusUnsupportedWhenStatFails(t *testing.T) {
	origLstat := lstat
	defer func() { lstat = origLstat }()
	lstat = func(string) (os.FileInfo, error) { return nil, errors.New("lstat") }
	if got := Status(t.TempDir()); !strings.HasPrefix(got, "unsupported:") {
		t.Fatalf("Status() after init = %q", got)
	}
}

func TestStatusBlocksSymlinkWorkspace(t *testing.T) {
	origLstat := lstat
	defer func() { lstat = origLstat }()
	lstat = func(string) (os.FileInfo, error) {
		return fakeInfo{mode: os.ModeSymlink | os.ModeDir}, nil
	}
	if got := Status(t.TempDir()); !strings.HasPrefix(got, "blocked:") {
		t.Fatalf("Status() symlink workspace = %q", got)
	}
}

type fakeInfo struct{ mode os.FileMode }

func (f fakeInfo) Name() string       { return "fake" }
func (f fakeInfo) Size() int64        { return 0 }
func (f fakeInfo) Mode() os.FileMode  { return f.mode }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return f.mode&os.ModeDir != 0 }
func (f fakeInfo) Sys() any           { return nil }

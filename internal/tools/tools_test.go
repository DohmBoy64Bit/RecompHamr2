package tools

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func helperCommand(ctx context.Context, mode string) *exec.Cmd {
	args := []string{"-test.run=TestHelperProcess", "--", mode}
	cmd := exec.CommandContext(ctx, os.Args[0], args...)
	cmd.Env = append(os.Environ(), "RECOMPHAMR_TOOL_HELPER=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("RECOMPHAMR_TOOL_HELPER") != "1" {
		return
	}
	mode := ""
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			mode = os.Args[i+1]
			break
		}
	}
	switch mode {
	case "hi":
		os.Stdout.WriteString("hi\n")
	case "cloned":
		os.Stdout.WriteString("cloned\n")
	case "fail":
		os.Stdout.WriteString("nope\n")
		os.Exit(7)
	case "sleep":
		time.Sleep(time.Second)
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

func TestSchemas(t *testing.T) {
	schemas := Schemas()
	if got := len(schemas); got != 6 {
		t.Fatalf("len(Schemas()) = %d, want 6", got)
	}
	if schemas[0].Name != PowerShellName {
		t.Fatalf("primary shell schema = %q, want %q", schemas[0].Name, PowerShellName)
	}
	compat := CompatibilityToolNames()
	if len(compat) != 1 || compat[0] != BashName {
		t.Fatalf("CompatibilityToolNames() = %#v, want bash alias", compat)
	}
}

func TestReadWriteEditFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "file.txt")
	if out := WriteFile(path, "hello world"); !strings.Contains(out, "wrote 11 bytes") {
		t.Fatalf("WriteFile() = %q", out)
	}
	if got := ReadFile(path); got != "hello world" {
		t.Fatalf("ReadFile() = %q", got)
	}
	if out := EditFile(path, "world", "there"); !strings.Contains(out, "edited") {
		t.Fatalf("EditFile() = %q", out)
	}
	if got := ReadFile(path); got != "hello there" {
		t.Fatalf("edited content = %q", got)
	}
	if !IsFailure(ReadFile("")) || !IsFailure(WriteFile("", "")) || !IsFailure(EditFile(path, "missing", "x")) {
		t.Fatal("expected failure conventions for invalid file tool calls")
	}
	if out := EditFile(path, "e", "x"); !IsFailure(out) {
		t.Fatalf("ambiguous edit output = %q, want failure", out)
	}
	if out := EditFile("", "a", "b"); !IsFailure(out) {
		t.Fatalf("empty edit path output = %q, want failure", out)
	}
	if out := EditFile(path, "", "b"); !IsFailure(out) {
		t.Fatalf("empty old_string output = %q, want failure", out)
	}
	if out := EditFile(path, "hello", "hello"); !IsFailure(out) {
		t.Fatalf("same replacement output = %q, want failure", out)
	}
	large := strings.Repeat("a", MaxReadFileBytes+1)
	if out := WriteFile(path, large); !strings.Contains(out, "wrote") {
		t.Fatalf("large WriteFile() = %q", out)
	}
	if got := ReadFile(path); !strings.Contains(got, "truncated after") {
		t.Fatalf("large ReadFile() = %q, want truncation marker", got[len(got)-80:])
	}
}

func TestPowerShell(t *testing.T) {
	if !IsFailure(PowerShell(context.Background(), "", time.Second)) {
		t.Fatal("empty PowerShell command should fail")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if out := PowerShell(ctx, "Write-Output hi", time.Second); !IsFailure(out) || !strings.Contains(out, "cancelled") {
		t.Fatalf("cancelled PowerShell output = %q", out)
	}
	origCommand := commandContext
	defer func() { commandContext = origCommand }()
	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return helperCommand(ctx, "hi")
	}
	out := PowerShell(context.Background(), "Write-Output hi", 5*time.Second)
	if !strings.Contains(out, "hi") {
		t.Fatalf("PowerShell() = %q, want hi", out)
	}
	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return helperCommand(ctx, "fail")
	}
	if !IsFailure(PowerShell(context.Background(), "exit 7", 5*time.Second)) {
		t.Fatal("failing PowerShell command should fail")
	}
}

func TestBashCompatibilityAlias(t *testing.T) {
	cmd := "printf hi"
	if runtime.GOOS == "windows" {
		cmd = "Write-Output hi"
	}
	out := Bash(context.Background(), cmd, 5*time.Second)
	if !strings.Contains(out, "hi") {
		t.Fatalf("Bash() = %q, want hi", out)
	}
	if !IsFailure(Bash(context.Background(), "", time.Second)) {
		t.Fatal("empty bash compatibility command should fail")
	}
}

func TestShellWindowsBranchWithCommandSeam(t *testing.T) {
	origCommand, origGOOS := commandContext, runtimeGOOS
	defer func() { commandContext, runtimeGOOS = origCommand, origGOOS }()
	runtimeGOOS = "windows"
	seen := ""
	seenArgs := []string{}
	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		seen = name
		seenArgs = append([]string{}, args...)
		return helperCommand(ctx, "hi")
	}
	out := PowerShell(context.Background(), "ignored", time.Second)
	if seen != "powershell" || !strings.Contains(strings.Join(seenArgs, " "), "-NonInteractive") || !strings.Contains(out, "hi") {
		t.Fatalf("windows branch seen=%q args=%v out=%q", seen, seenArgs, out)
	}
}

func TestShellLinuxAndTimeoutBranchesWithSeams(t *testing.T) {
	origCommand, origGOOS := commandContext, runtimeGOOS
	defer func() { commandContext, runtimeGOOS = origCommand, origGOOS }()
	runtimeGOOS = "linux"
	seen := ""
	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		seen = name
		return helperCommand(ctx, "hi")
	}
	out := Bash(context.Background(), "ignored", 0)
	if seen != "/bin/sh" || !strings.Contains(out, "hi") {
		t.Fatalf("linux branch seen=%q out=%q", seen, out)
	}
	out = PowerShell(context.Background(), "ignored", MaxShellTimeout+time.Minute)
	if seen != "pwsh" || !strings.Contains(out, "hi") {
		t.Fatalf("non-windows powershell branch seen=%q out=%q", seen, out)
	}

	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return helperCommand(ctx, "sleep")
	}
	if out := Bash(context.Background(), "ignored", time.Nanosecond); !IsFailure(out) || !strings.Contains(out, "timeout") {
		t.Fatalf("timeout output = %q", out)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()
	if out := PowerShell(ctx, "ignored", time.Minute); !IsFailure(out) || !strings.Contains(out, "cancelled") {
		t.Fatalf("cancelled during command output = %q", out)
	}
}

func TestShellHelpers(t *testing.T) {
	if got := normalizeShellTimeout(0); got != DefaultShellTimeout {
		t.Fatalf("normalizeShellTimeout(0) = %s, want %s", got, DefaultShellTimeout)
	}
	if got := normalizeShellTimeout(MaxShellTimeout + time.Second); got != MaxShellTimeout {
		t.Fatalf("normalizeShellTimeout(cap) = %s, want %s", got, MaxShellTimeout)
	}
	if got := appendToolLine("", "(x)"); got != "(x)" {
		t.Fatalf("appendToolLine(empty) = %q", got)
	}
	if got := appendToolLine("out", "(x)"); got != "out\n(x)" {
		t.Fatalf("appendToolLine(out) = %q", got)
	}
}

func TestRecompReference(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("reference"))
	}))
	defer server.Close()
	out := RecompReference(context.Background(), server.Client(), server.URL+"/doc", t.TempDir())
	if !strings.Contains(out, "fetched") {
		t.Fatalf("RecompReference() = %q", out)
	}
	if !IsFailure(RecompReference(context.Background(), server.Client(), "://bad", t.TempDir())) {
		t.Fatal("bad URL should fail")
	}
	if !IsFailure(RecompReference(context.Background(), server.Client(), server.URL, "")) {
		t.Fatal("empty output dir should fail")
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})}
	if !IsFailure(RecompReference(context.Background(), client, "https://.", t.TempDir())) {
		t.Fatal("empty cache name should fail")
	}
}

func TestRepomixrValidationAndUnsupported(t *testing.T) {
	if !IsFailure(Repomixr(context.Background(), "://bad", t.TempDir())) {
		t.Fatal("bad repomixr URL should fail")
	}
	if !IsFailure(Repomixr(context.Background(), "https://example.com/a/b", t.TempDir())) {
		t.Fatal("non-github repomixr URL should fail")
	}
	if !IsFailure(Repomixr(context.Background(), "https://github.com/owner/repo", "")) {
		t.Fatal("empty repomixr output dir should fail")
	}
	if !IsFailure(Repomixr(context.Background(), "https://github.com/../repo", t.TempDir())) {
		t.Fatal("invalid repomixr owner/repo should fail")
	}
	if !IsFailure(Repomixr(context.Background(), "https://github.com/owner/repo/extra", t.TempDir())) {
		t.Fatal("extra path segments should fail")
	}
	if err := Unsupported("x"); err == nil || !strings.Contains(err.Error(), "unsupported tool") {
		t.Fatalf("Unsupported() = %v", err)
	}
}

func TestPathHelpers(t *testing.T) {
	if name, ok := githubRepoName("/owner/repo.git"); !ok || name != "owner-repo" {
		t.Fatalf("githubRepoName() = %q, %v; want owner-repo true", name, ok)
	}
	for _, path := range []string{"/owner", "/../repo", "/owner/..", "/owner/repo/extra", "/owner/re:po"} {
		if name, ok := githubRepoName(path); ok {
			t.Fatalf("githubRepoName(%q) = %q, true; want false", path, name)
		}
	}
	if !safePathSegment("abc") || safePathSegment("") || safePathSegment(".") || safePathSegment("..") || safePathSegment(`a\b`) {
		t.Fatal("safePathSegment validation mismatch")
	}
	if got := cacheFileName("example.com/a:../b"); got != "example.com-a---b" {
		t.Fatalf("cacheFileName() = %q", got)
	}
	base := t.TempDir()
	if path, err := joinInside(base, "x", "y"); err != nil || !strings.HasPrefix(path, base) {
		t.Fatalf("joinInside() = %q, %v", path, err)
	}
	if _, err := joinInside(base, "..", "x"); err == nil {
		t.Fatal("joinInside escape should fail")
	}
	origAbs, origRel := filepathAbs, filepathRel
	defer func() { filepathAbs, filepathRel = origAbs, origRel }()
	filepathAbs = func(string) (string, error) { return "", errors.New("abs") }
	if _, err := joinInside(base, "x"); err == nil || err.Error() != "abs" {
		t.Fatalf("joinInside abs error = %v, want abs", err)
	}
	filepathAbs = origAbs
	filepathRel = func(string, string) (string, error) { return "", errors.New("rel") }
	if _, err := joinInside(base, "x"); err == nil || err.Error() != "rel" {
		t.Fatalf("joinInside rel error = %v, want rel", err)
	}
	filepathRel = origRel
	filepathAbs = func(string) (string, error) { return "", errors.New("abs") }
	if !IsFailure(Repomixr(context.Background(), "https://github.com/a/b", base)) {
		t.Fatal("repomixr joinInside failure should surface")
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})}
	if !IsFailure(RecompReference(context.Background(), client, "https://example.com/doc", base)) {
		t.Fatal("reference joinInside failure should surface")
	}
}

func TestToolFailureSeams(t *testing.T) {
	origMkdir, origWrite, origRead, origCommand := mkdirAll, writeFile, readFile, commandContext
	defer func() {
		mkdirAll, writeFile, readFile, commandContext = origMkdir, origWrite, origRead, origCommand
	}()

	readFile = func(string) ([]byte, error) { return nil, errors.New("read") }
	if !IsFailure(ReadFile("x")) || !IsFailure(EditFile("x", "a", "b")) {
		t.Fatal("read seam failures should surface")
	}
	readFile = origRead

	mkdirAll = func(string, os.FileMode) error { return errors.New("mkdir") }
	if !IsFailure(WriteFile(filepath.Join("x", "y"), "z")) || !IsFailure(Repomixr(context.Background(), "https://github.com/a/b", t.TempDir())) {
		t.Fatal("mkdir seam failures should surface")
	}
	mkdirAll = origMkdir

	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write") }
	file := filepath.Join(t.TempDir(), "x")
	if err := origWrite(file, []byte("a"), 0o600); err != nil {
		t.Fatalf("WriteFile() setup error = %v", err)
	}
	if !IsFailure(WriteFile(file, "z")) || !IsFailure(EditFile(file, "a", "b")) {
		t.Fatal("write seam failures should surface")
	}
	writeFile = origWrite

	commandContext = func(context.Context, string, ...string) *exec.Cmd {
		return exec.Command("definitely-missing-recomphamr-command")
	}
	if !IsFailure(Repomixr(context.Background(), "https://github.com/a/b", t.TempDir())) {
		t.Fatal("git failure should surface")
	}
	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return helperCommand(ctx, "cloned")
	}
	if out := Repomixr(context.Background(), "https://github.com/a/b", t.TempDir()); !strings.Contains(out, "cloned") {
		t.Fatalf("repomixr success seam output = %q", out)
	}
}

func TestRecompReferenceHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	defer server.Close()
	if out := RecompReference(context.Background(), server.Client(), server.URL, t.TempDir()); !IsFailure(out) {
		t.Fatalf("HTTP failure output = %q", out)
	}
	if err := os.MkdirAll(filepath.Join(t.TempDir(), "x"), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
}

func TestRecompReferenceSeams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	origMkdir, origWrite := mkdirAll, writeFile
	defer func() { mkdirAll, writeFile = origMkdir, origWrite }()
	mkdirAll = func(string, os.FileMode) error { return errors.New("mkdir") }
	if !IsFailure(RecompReference(context.Background(), server.Client(), server.URL, t.TempDir())) {
		t.Fatal("reference mkdir failure should surface")
	}
	mkdirAll = origMkdir
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write") }
	if !IsFailure(RecompReference(context.Background(), server.Client(), server.URL, t.TempDir())) {
		t.Fatal("reference write failure should surface")
	}
}

func TestRecompReferenceClientAndReadFailures(t *testing.T) {
	if !IsFailure(RecompReference(context.Background(), nil, "http://127.0.0.1:1", t.TempDir())) {
		t.Fatal("nil client fetch failure should surface")
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: failingBody{}}, nil
	})}
	if !IsFailure(RecompReference(context.Background(), client, "https://example.com/doc", t.TempDir())) {
		t.Fatal("body read failure should surface")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("read") }
func (failingBody) Close() error             { return nil }

var _ io.ReadCloser = failingBody{}

func BenchmarkSchemas(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if got := Schemas(); len(got) != 6 {
			b.Fatalf("len(Schemas()) = %d", len(got))
		}
	}
}

func BenchmarkReadFileSmall(b *testing.B) {
	path := filepath.Join(b.TempDir(), "small.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("a", 4096)), 0o600); err != nil {
		b.Fatalf("WriteFile() setup error = %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if got := ReadFile(path); len(got) == 0 {
			b.Fatal("ReadFile returned empty content")
		}
	}
}

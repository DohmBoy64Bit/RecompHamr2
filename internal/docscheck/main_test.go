package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type exitCode int

func TestMainFunction(t *testing.T) {
	originalExit := exit
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() {
		exit = originalExit
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Chdir() restore error = %v", err)
		}
	}()

	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		t.Fatalf("Chdir() repo root error = %v", err)
	}

	exit = func(code int) {
		panic(exitCode(code))
	}

	defer func() {
		recovered := recover()
		code, ok := recovered.(exitCode)
		if !ok {
			t.Fatalf("main() panic = %#v, want exitCode", recovered)
		}
		if code != 0 {
			t.Fatalf("main() exit code = %d, want 0", code)
		}
	}()

	main()
}

func TestRunSuccess(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(doc, []byte("# Doc\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(&stdout, &stderr, dir, []string{"doc.md"}, nil)

	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "docscheck: durable memory docs and coverage gates pass") {
		t.Fatalf("stdout = %q, want success message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunFailure(t *testing.T) {
	dir := t.TempDir()
	emptyDoc := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(emptyDoc, []byte("  \n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(&stdout, &stderr, dir, []string{"empty.md", "missing.md"}, nil)

	if code != 1 {
		t.Fatalf("run() exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "empty required doc") {
		t.Fatalf("stderr = %q, want empty-doc failure", stderr.String())
	}
	if !strings.Contains(stderr.String(), "missing required doc") {
		t.Fatalf("stderr = %q, want missing-doc failure", stderr.String())
	}
}

func TestCoverageItems(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "docs/a.md", "alpha beta")
	writeDoc(t, root, "docs/b.md", "gamma")

	failures := checkCoverageItems(root, []coverageItem{{Name: "sample", Docs: []string{"docs/a.md", "docs/b.md"}, Terms: []string{"alpha", "gamma"}}})
	if len(failures) != 0 {
		t.Fatalf("checkCoverageItems() failures = %v, want none", failures)
	}
	failures = checkCoverageItems(root, []coverageItem{{Name: "sample", Docs: []string{"docs/missing.md"}, Terms: []string{"delta"}}})
	if len(failures) != 2 || !strings.Contains(strings.Join(failures, "\n"), "missing coverage doc") || !strings.Contains(strings.Join(failures, "\n"), "undocumented sample term") {
		t.Fatalf("checkCoverageItems() failures = %v", failures)
	}
}

func TestExportedDocs(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "internal/good/doc.go", "// Package good is documented.\npackage good\n")
	writeDoc(t, root, "internal/good/good.go", `package good

// Thing is documented.
type Thing struct{}

// Value is documented.
const Value = "x"

// Do returns a result.
func Do() {}
`)
	failures, err := checkExportedDocs(root)
	if err != nil {
		t.Fatalf("checkExportedDocs() error = %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("checkExportedDocs() failures = %v, want none", failures)
	}
}

func TestExportedDocsFailures(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "internal/bad/bad.go", `package bad

type MissingType struct{}
const MissingValue = "x"
func MissingFunc() {}
`)
	failures, err := checkExportedDocs(root)
	if err != nil {
		t.Fatalf("checkExportedDocs() error = %v", err)
	}
	text := strings.Join(failures, "\n")
	for _, want := range []string{"package bad missing package doc", "exported type MissingType", "exported value MissingValue", "exported function MissingFunc"} {
		if !strings.Contains(text, want) {
			t.Fatalf("checkExportedDocs() missing %q in %v", want, failures)
		}
	}
}

func TestRunReportsCoverageAndExportedDocFailures(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "doc.md", "# Doc\n")
	writeDoc(t, root, "internal/bad/bad.go", "package bad\nfunc Missing() {}\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, root, []string{"doc.md"}, []coverageItem{{Name: "sample", Docs: []string{"doc.md"}, Terms: []string{"absent"}}})

	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("run() code=%d stdout=%q, want failure with empty stdout", code, stdout.String())
	}
	if !strings.Contains(stderr.String(), "undocumented sample term") || !strings.Contains(stderr.String(), "exported function Missing") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunReportsExportedDocWalkError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(&stdout, &stderr, filepath.Join(t.TempDir(), "missing"), nil, nil)
	if code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("run() missing root code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestExportedDocsSkipsNonProductFilesAndMainPackages(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, ".git/internal/skip/skip.go", "package skip\nfunc Missing() {}\n")
	writeDoc(t, root, ".reference/internal/skip/skip.go", "package skip\nfunc Missing() {}\n")
	writeDoc(t, root, "outside/out.go", "package outside\nfunc Missing() {}\n")
	writeDoc(t, root, "internal/good/good_test.go", "package good\nfunc Missing() {}\n")
	writeDoc(t, root, "cmd/tool/main.go", "package main\nfunc main() {}\n")
	failures, err := checkExportedDocs(root)
	if err != nil {
		t.Fatalf("checkExportedDocs() error = %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("checkExportedDocs() failures = %v, want none", failures)
	}
}

func TestPackageDocMapParseError(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "internal/bad/bad.go", "package ")
	if _, err := packageDocMap(root); err == nil {
		t.Fatal("packageDocMap() accepted parse error")
	}
}

func TestCheckExportedDocsInFileParseError(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "internal", "bad", "bad.go")
	writeDoc(t, root, "internal/bad/bad.go", "package ")
	if _, _, err := checkExportedDocsInFile(root, bad, "internal/bad/bad.go"); err == nil {
		t.Fatal("checkExportedDocsInFile() accepted parse error")
	}
}

func TestExportedDocsWalkAndRelErrors(t *testing.T) {
	origWalk, origRel := walkDir, relPath
	defer func() { walkDir, relPath = origWalk, origRel }()
	walkDir = func(string, fs.WalkDirFunc) error { return errors.New("walk") }
	if _, err := packageDocMap(t.TempDir()); err == nil || !strings.Contains(err.Error(), "walk") {
		t.Fatalf("packageDocMap() walk error = %v", err)
	}

	root := t.TempDir()
	writeDoc(t, root, "internal/good/doc.go", "// Package good is documented.\npackage good\n")
	walkDir = origWalk
	relPath = func(string, string) (string, error) { return "", errors.New("rel") }
	if _, err := packageDocMap(root); err == nil || !strings.Contains(err.Error(), "rel") {
		t.Fatalf("packageDocMap() rel error = %v", err)
	}
	if _, err := checkExportedDocs(root); err == nil || !strings.Contains(err.Error(), "rel") {
		t.Fatalf("checkExportedDocs() rel error = %v", err)
	}
}

func TestCheckExportedDocsSecondWalkErrors(t *testing.T) {
	origWalk, origRel := walkDir, relPath
	defer func() {
		walkDir = origWalk
		relPath = origRel
	}()
	root := t.TempDir()
	writeDoc(t, root, "internal/good/doc.go", "// Package good is documented.\npackage good\n")

	walkCount := 0
	walkDir = func(root string, fn fs.WalkDirFunc) error {
		walkCount++
		if walkCount == 2 {
			return errors.New("second walk")
		}
		return origWalk(root, fn)
	}
	if _, err := checkExportedDocs(root); err == nil || !strings.Contains(err.Error(), "second walk") {
		t.Fatalf("checkExportedDocs() second walk error = %v", err)
	}

	walkDir = origWalk
	relCount := 0
	relPath = func(root string, file string) (string, error) {
		relCount++
		if relCount > 1 {
			return "", errors.New("second rel")
		}
		return origRel(root, file)
	}
	if _, err := checkExportedDocs(root); err == nil || !strings.Contains(err.Error(), "second rel") {
		t.Fatalf("checkExportedDocs() second rel error = %v", err)
	}
}

func TestCheckExportedDocsCallbackAndSecondParseErrors(t *testing.T) {
	origWalk := walkDir
	defer func() { walkDir = origWalk }()
	root := t.TempDir()
	writeDoc(t, root, "internal/good/doc.go", "// Package good is documented.\npackage good\n")

	walkCount := 0
	walkDir = func(root string, fn fs.WalkDirFunc) error {
		walkCount++
		if walkCount == 2 {
			return fn(root, nil, errors.New("callback"))
		}
		return origWalk(root, fn)
	}
	if _, err := checkExportedDocs(root); err == nil || !strings.Contains(err.Error(), "callback") {
		t.Fatalf("checkExportedDocs() callback error = %v", err)
	}

	root = t.TempDir()
	writeDoc(t, root, "internal/good/doc.go", "// Package good is documented.\npackage good\n")
	writeDoc(t, root, "internal/bad/bad.go", "package ")
	walkCount = 0
	walkDir = func(root string, fn fs.WalkDirFunc) error {
		walkCount++
		if walkCount == 1 {
			return fn(filepath.Join(root, "internal", "good", "doc.go"), fakeDirEntry{name: "doc.go"}, nil)
		}
		return origWalk(root, fn)
	}
	if _, err := checkExportedDocs(root); err == nil {
		t.Fatal("checkExportedDocs() accepted second parse error")
	}
}

func writeDoc(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

type fakeDirEntry struct{ name string }

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return false }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

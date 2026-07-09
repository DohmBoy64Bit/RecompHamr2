package main

import (
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
	exit = func(code int) { panic(exitCode(code)) }
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

func TestCheckAllowsDocumentedImport(t *testing.T) {
	root := t.TempDir()
	writeGo(t, root, "internal/tui/tui.go", `package tui
import _ "recomphamr2/internal/commands"
`)
	writeGo(t, root, "internal/app/app.go", `package app
import (
	_ "recomphamr2/internal/agent"
	_ "recomphamr2/internal/commands"
	_ "recomphamr2/internal/config"
	_ "recomphamr2/internal/llm"
	_ "recomphamr2/internal/mcp"
	_ "recomphamr2/internal/project"
	_ "recomphamr2/internal/tui"
)
`)
	failures, err := check(root)
	if err != nil {
		t.Fatalf("check() error = %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("check() failures = %v, want none", failures)
	}
}

func TestCheckRejectsUndocumentedImport(t *testing.T) {
	root := t.TempDir()
	writeGo(t, root, "internal/tui/tui.go", `package tui
import _ "recomphamr2/internal/tools"
`)
	failures, err := check(root)
	if err != nil {
		t.Fatalf("check() error = %v", err)
	}
	if len(failures) != 1 || !strings.Contains(failures[0], "must not import internal/tools") {
		t.Fatalf("check() failures = %v", failures)
	}
}

func TestCheckRejectsUnknownInternalPackage(t *testing.T) {
	root := t.TempDir()
	writeGo(t, root, "internal/newpkg/newpkg.go", `package newpkg
import _ "recomphamr2/internal/tools"
`)
	failures, err := check(root)
	if err != nil {
		t.Fatalf("check() error = %v", err)
	}
	if len(failures) != 1 || !strings.Contains(failures[0], "undocumented package boundary") {
		t.Fatalf("check() failures = %v", failures)
	}
}

func TestCheckPropagatesParseAndWalkErrors(t *testing.T) {
	root := t.TempDir()
	writeGo(t, root, "internal/tui/bad.go", "package tui\nimport ")
	if _, err := check(root); err == nil {
		t.Fatal("check() accepted parse error")
	}
	if _, err := check(filepath.Join(root, "missing")); err == nil {
		t.Fatal("check() accepted missing root")
	}
}

func TestRunFailure(t *testing.T) {
	root := t.TempDir()
	writeGo(t, root, "internal/tui/tui.go", `package tui
import _ "recomphamr2/internal/tools"
`)
	if code := run(root); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if code := run(filepath.Join(root, "missing")); code != 1 {
		t.Fatalf("run() missing root = %d, want 1", code)
	}
}

func TestHelpers(t *testing.T) {
	if owner, ok := ownerPackage(".", filepath.Join("cmd", "recomphamr", "main.go")); !ok || owner != "cmd/recomphamr" {
		t.Fatalf("ownerPackage cmd = %q, %v", owner, ok)
	}
	if owner, ok := ownerPackage(".", filepath.Join("docs", "x.go")); ok || owner != "" {
		t.Fatalf("ownerPackage docs = %q, %v", owner, ok)
	}
	if target, ok := internalImport(`"recomphamr2/internal/tools"`); !ok || target != "internal/tools" {
		t.Fatalf("internalImport = %q, %v", target, ok)
	}
	if target, ok := internalImport(`"fmt"`); ok || target != "" {
		t.Fatalf("internalImport fmt = %q, %v", target, ok)
	}
	if target, ok := internalImport(`"recomphamr2/internal/"`); !ok || target != "internal/" {
		t.Fatalf("internalImport package root = %q, %v", target, ok)
	}
}

func TestOwnerPackageEdgeCases(t *testing.T) {
	root := t.TempDir()
	doc := filepath.Join(root, "docs", "x.go")
	if failures, err := checkFile(root, doc); err != nil || failures != nil {
		t.Fatalf("checkFile() docs = %v, %v; want nil, nil", failures, err)
	}
	writeGo(t, root, "internal/x.go", "package internal\n")
	if failures, err := check(root); err != nil {
		t.Fatalf("check() internal root error = %v", err)
	} else if len(failures) != 0 {
		t.Fatalf("check() internal root failures = %v, want none", failures)
	}
	originalRel := relPath
	defer func() { relPath = originalRel }()
	relPath = func(string, string) (string, error) { return "", os.ErrInvalid }
	if owner, ok := ownerPackage(root, filepath.Join(root, "x.go")); ok || owner != "" {
		t.Fatalf("ownerPackage rel error = %q, %v", owner, ok)
	}
}

func writeGo(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

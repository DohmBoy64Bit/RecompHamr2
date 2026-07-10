package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var exit = os.Exit
var relPath = filepath.Rel

var allowedInternalImports = map[string]map[string]bool{
	"cmd/recomphamr":       {"internal/app": true},
	"internal/agent":       {"internal/llm": true},
	"internal/commands":    {"internal/config": true, "internal/doctor": true, "internal/mcp": true, "internal/parity": true, "internal/project": true, "internal/skills": true, "internal/tools": true},
	"internal/project":     {"internal/config": true},
	"internal/tui":         {"internal/commands": true, "internal/security": true},
	"internal/app":         {"internal/agent": true, "internal/commands": true, "internal/config": true, "internal/llm": true, "internal/mcp": true, "internal/project": true, "internal/tools": true, "internal/tui": true},
	"internal/archcheck":   {},
	"internal/config":      {},
	"internal/covercheck":  {},
	"internal/docscheck":   {},
	"internal/doctor":      {"internal/config": true, "internal/mcp": true, "internal/project": true, "internal/release": true, "internal/skills": true, "internal/tools": true},
	"internal/llm":         {},
	"internal/logging":     {"internal/security": true},
	"internal/mcp":         {},
	"internal/parity":      {},
	"internal/release":     {},
	"internal/security":    {},
	"internal/skills":      {},
	"internal/testharness": {},
	"internal/tools":       {},
	"internal/update":      {"internal/release": true},
}

func main() {
	exit(run("."))
}

func run(root string) int {
	failures, err := check(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(failures) > 0 {
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, failure)
		}
		return 1
	}
	fmt.Println("archcheck: separation-of-concerns boundaries hold")
	return 0
}

func check(root string) ([]string, error) {
	var failures []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == ".reference" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		fileFailures, err := checkFile(root, path)
		if err != nil {
			return err
		}
		failures = append(failures, fileFailures...)
		return nil
	})
	sort.Strings(failures)
	return failures, err
}

func checkFile(root string, filePath string) ([]string, error) {
	owner, ok := ownerPackage(root, filePath)
	if !ok {
		return nil, nil
	}
	file, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var failures []string
	for _, spec := range file.Imports {
		target, ok := internalImport(spec.Path.Value)
		if !ok {
			continue
		}
		allowed, known := allowedInternalImports[owner]
		if !known {
			failures = append(failures, fmt.Sprintf("%s: undocumented package boundary %q", filePath, owner))
			continue
		}
		if !allowed[target] {
			failures = append(failures, fmt.Sprintf("%s: %s must not import %s", filePath, owner, target))
		}
	}
	return failures, nil
}

func ownerPackage(root string, filePath string) (string, bool) {
	rel, err := relPath(root, filePath)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	dir := path.Dir(rel)
	if strings.HasPrefix(dir, "cmd/recomphamr") {
		return "cmd/recomphamr", true
	}
	if !strings.HasPrefix(dir, "internal/") {
		return "", false
	}
	parts := strings.Split(dir, "/")
	return parts[0] + "/" + parts[1], true
}

func internalImport(quoted string) (string, bool) {
	path := strings.Trim(quoted, `"`)
	const prefix = "recomphamr2/"
	if !strings.HasPrefix(path, prefix+"internal/") {
		return "", false
	}
	rel := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rel, "/")
	return parts[0] + "/" + parts[1], true
}

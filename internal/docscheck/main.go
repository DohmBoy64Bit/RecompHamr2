package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

var requiredFiles = []string{
	"AGENTS.md",
	"README.md",
	"SECURITY.md",
	"CONTRIBUTING.md",
	"recomphamr_2_rewrite_workflow.md",
	".docs-index.md",
	"docs/dev/00_Governance/AntiHallucination.md",
	"docs/dev/00_Governance/DefinitionOfDone.md",
	"docs/dev/02_Verification/DocumentationCoverage.md",
	"docs/dev/03_Architecture/SeparationOfConcerns.md",
	"docs/dev/04_Workflows/PhaseGoals.md",
	"docs/dev/05_FeatureParity/ParityMatrix.md",
	"docs/dev/06_Testing/CoverageRequirements.md",
	"docs/dev/07_ProjectManagement/TraceabilityMatrix.md",
}

var exit = os.Exit
var walkDir = filepath.WalkDir
var relPath = filepath.Rel

func main() {
	exit(run(os.Stdout, os.Stderr, ".", requiredFiles, coverageItems))
}

func run(stdout io.Writer, stderr io.Writer, root string, files []string, items []coverageItem) int {
	failures := checkRequired(root, files)
	failures = append(failures, checkCoverageItems(root, items)...)
	symbolFailures, err := checkExportedDocs(root)
	if err != nil {
		failures = append(failures, err.Error())
	} else {
		failures = append(failures, symbolFailures...)
	}
	if len(failures) > 0 {
		slices.Sort(failures)
		for _, failure := range failures {
			fmt.Fprintln(stderr, failure)
		}
		return 1
	}

	fmt.Fprintln(stdout, "docscheck: durable memory docs and coverage gates pass")
	return 0
}

type coverageItem struct {
	Name  string
	Docs  []string
	Terms []string
}

var coverageItems = []coverageItem{
	{"slash commands", []string{"docs/user/commands.md", "docs/dev/05_FeatureParity/CommandParity.md"}, []string{"/clear", "/models", "/skills", "/skill", "/skill-audit", "/skill-new", "/init-re", "/status-re", "/doctor", "/mcp", "/help"}},
	{"tool schemas", []string{"docs/user/tools.md", "docs/dev/03_Architecture/ToolRuntime.md", "docs/dev/05_FeatureParity/ToolParity.md"}, []string{"powershell", "bash", "read_file", "write_file", "edit_file", "repomixr", "recomp_reference", "cmd", "timeout_seconds", "path", "content", "old_string", "new_string", "url", "output_dir"}},
	{"config keys and env", []string{"docs/user/configuration.md", "docs/user/model-profiles.md", "docs/dev/05_FeatureParity/ConfigParity.md"}, []string{"active", "models", "llm", "url", "key", "context_size", "logging", "RECOMPHAMR_URL"}},
	{"MCP servers and env", []string{"docs/user/mcp.md", "docs/dev/05_FeatureParity/MCPParity.md", "docs/dev/03_Architecture/MCPArchitecture.md"}, []string{"ghidra", "n64-debug-mcp", "pcrecomp", "mcp-pine", "objdiff", "pcsx2", "bizhawk", "sega2asm", "RECOMPHAMR_MCP_<NAME>_COMMAND", "RECOMPHAMR_MCP_<NAME>_URL", "RECOMPHAMR_MCP_<NAME>_TOOLS", "RECOMPHAMR_MCP_AUTOSTART"}},
	{"generated workspace files", []string{"docs/user/memory.md", "docs/dev/05_FeatureParity/MemoryParity.md"}, []string{"PROJECT.md", "REPHAMR_STATE.md", "EVIDENCE.md", "HYPOTHESES.md", "BLOCKERS.md", "CHANGELOG.md", "COMMANDS.md", "TOOLCHAIN.md", "MODELS.md", "mcp.json", "functions/inventory.csv", "formats/inventory.md", "recomp/runtime_gaps.md", "decomp/symbols.md", "skills/active.md"}},
	{"release files", []string{"docs/user/install.md", "docs/dev/05_FeatureParity/ReleaseParity.md", "docs/dev/04_Workflows/ReleaseWorkflow.md"}, []string{"scripts/install.ps1", "scripts/install.sh", ".goreleaser.yaml", ".devcontainer/devcontainer.json", ".github/workflows/verify.yml", "SHA256SUMS"}},
	{"CLI help", []string{"README.md", "docs/user/quickstart.md", "docs/user/commands.md"}, []string{"--diagnostic", "--help", "go run ./cmd/recomphamr --diagnostic"}},
}

func checkRequired(root string, files []string) []string {
	var failures []string
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			failures = append(failures, fmt.Sprintf("missing required doc: %s", name))
			continue
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			failures = append(failures, fmt.Sprintf("empty required doc: %s", name))
		}
	}
	return failures
}

func checkCoverageItems(root string, items []coverageItem) []string {
	var failures []string
	for _, item := range items {
		text, missingDocs := docsText(root, item.Docs)
		for _, doc := range missingDocs {
			failures = append(failures, fmt.Sprintf("missing coverage doc for %s: %s", item.Name, doc))
		}
		for _, term := range item.Terms {
			if !strings.Contains(text, term) {
				failures = append(failures, fmt.Sprintf("undocumented %s term: %s", item.Name, term))
			}
		}
	}
	return failures
}

func docsText(root string, docs []string) (string, []string) {
	var b strings.Builder
	var missing []string
	for _, doc := range docs {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(doc)))
		if err != nil {
			missing = append(missing, doc)
			continue
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	return b.String(), missing
}

func checkExportedDocs(root string) ([]string, error) {
	packageDocs, err := packageDocMap(root)
	if err != nil {
		return nil, err
	}
	var failures []string
	seenPackages := map[string]string{}
	err = walkDir(root, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".reference":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(filePath) != ".go" || strings.HasSuffix(filePath, "_test.go") {
			return nil
		}
		rel, err := relPath(root, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !strings.HasPrefix(rel, "cmd/") && !strings.HasPrefix(rel, "internal/") {
			return nil
		}
		fileFailures, packageName, err := checkExportedDocsInFile(root, filePath, rel)
		if err != nil {
			return err
		}
		if packageName != "main" {
			seenPackages[path.Dir(rel)] = packageName
		}
		failures = append(failures, fileFailures...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for dir, packageName := range seenPackages {
		if !packageDocs[dir] {
			failures = append(failures, fmt.Sprintf("%s: package %s missing package doc", dir, packageName))
		}
	}
	return failures, err
}

func packageDocMap(root string) (map[string]bool, error) {
	docs := map[string]bool{}
	err := walkDir(root, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".reference":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(filePath) != ".go" || strings.HasSuffix(filePath, "_test.go") {
			return nil
		}
		rel, err := relPath(root, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !strings.HasPrefix(rel, "cmd/") && !strings.HasPrefix(rel, "internal/") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		if file.Doc != nil {
			docs[path.Dir(rel)] = true
		}
		return nil
	})
	return docs, err
}

func checkExportedDocsInFile(root string, filePath string, rel string) ([]string, string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, "", err
	}
	var failures []string
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			failures = append(failures, checkGenDeclDocs(rel, d)...)
		case *ast.FuncDecl:
			if d.Name.IsExported() && !docStartsWith(d.Doc, d.Name.Name) {
				failures = append(failures, fmt.Sprintf("%s: exported function %s missing Go doc", rel, d.Name.Name))
			}
		}
	}
	return failures, file.Name.Name, nil
}

func checkGenDeclDocs(rel string, decl *ast.GenDecl) []string {
	var failures []string
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() && !docStartsWith(firstDoc(s.Doc, decl.Doc), s.Name.Name) {
				failures = append(failures, fmt.Sprintf("%s: exported type %s missing Go doc", rel, s.Name.Name))
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() && !docStartsWith(firstDoc(s.Doc, decl.Doc), name.Name) {
					failures = append(failures, fmt.Sprintf("%s: exported value %s missing Go doc", rel, name.Name))
				}
			}
		}
	}
	return failures
}

func firstDoc(primary *ast.CommentGroup, fallback *ast.CommentGroup) *ast.CommentGroup {
	if primary != nil {
		return primary
	}
	return fallback
}

func docStartsWith(doc *ast.CommentGroup, name string) bool {
	if doc == nil {
		return false
	}
	text := strings.TrimSpace(doc.Text())
	return strings.HasPrefix(text, name+" ") || strings.HasPrefix(text, name+".") || strings.HasPrefix(text, name+" is") || strings.HasPrefix(text, name+" returns")
}

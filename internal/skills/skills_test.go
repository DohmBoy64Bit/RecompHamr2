package skills

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedSkills(t *testing.T) {
	got := Embedded()
	if len(got) != 28 {
		t.Fatalf("len(Embedded()) = %d, want 28", len(got))
	}
	if got[0].Name == "" || got[0].Body == "" {
		t.Fatalf("first embedded skill incomplete: %+v", got[0])
	}
}

func TestResolveCustomPrecedenceAndSuffix(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ghidra-mcp.md"), []byte("custom"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	skill, err := Resolve("GHIDRA-MCP.md", dir)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if skill.Source != "custom" || skill.Body != "custom" {
		t.Fatalf("Resolve() = %+v, want custom override", skill)
	}
	if _, err := Resolve("", dir); err == nil {
		t.Fatal("Resolve() accepted empty query")
	}
	if _, err := Resolve("missing", dir); err == nil {
		t.Fatal("Resolve() accepted missing skill")
	}
	got, err := Get("ghidra-mcp", dir)
	if err != nil || got.Source != "custom" {
		t.Fatalf("Get() = %+v, %v; want custom skill", got, err)
	}
}

func TestLoadCustomAndAudit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.md"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() txt error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	got, err := LoadCustom(dir)
	if err != nil {
		t.Fatalf("LoadCustom() error = %v", err)
	}
	if len(got) != 1 || got[0].Name != "x" {
		t.Fatalf("LoadCustom() = %+v", got)
	}
	if Audit("n64-debug-mcp") != "runtime-integration" {
		t.Fatal("debug MCP audit classification wrong")
	}
	if Audit("xboxrecomp") != "reverse-engineering-workflow" {
		t.Fatal("recomp audit classification wrong")
	}
	if Audit("evidence-mode") != "methodology" {
		t.Fatal("methodology audit classification wrong")
	}
	names, err := Names(dir)
	if err != nil {
		t.Fatalf("Names() error = %v", err)
	}
	if !sorted(names) || !contains(names, "x") || !contains(names, "core-re") {
		t.Fatalf("Names() = %v, want sorted embedded+custom names", names)
	}
	list, err := ListMarkdown([]string{"x", "CORE-RE.md"}, dir)
	if err != nil {
		t.Fatalf("ListMarkdown() error = %v", err)
	}
	for _, want := range []string{"Built-in RE skills", "* x (custom)", "* core-re", "Load one with /skill <name>."} {
		if !strings.Contains(list, want) {
			t.Fatalf("ListMarkdown() missing %q:\n%s", want, list)
		}
	}
	if !IsEmbedded("core-re") || IsEmbedded("x") || IsEmbedded("") {
		t.Fatal("IsEmbedded() classification mismatch")
	}
	missing, err := LoadCustom(filepath.Join(t.TempDir(), "missing"))
	if err != nil || missing != nil {
		t.Fatalf("LoadCustom() missing = %+v, %v; want nil, nil", missing, err)
	}
	empty, err := LoadCustom("")
	if err != nil || empty != nil {
		t.Fatalf("LoadCustom(empty) = %+v, %v; want nil, nil", empty, err)
	}
}

func TestCustomFailureSeams(t *testing.T) {
	origReadDir, origReadFile, origMkdirAll, origWriteFile := readDir, readFile, mkdirAll, writeFile
	defer func() {
		readDir, readFile = origReadDir, origReadFile
		mkdirAll, writeFile = origMkdirAll, origWriteFile
	}()
	readDir = func(string) ([]os.DirEntry, error) { return nil, errors.New("readdir") }
	if _, err := LoadCustom(t.TempDir()); err == nil {
		t.Fatal("LoadCustom() accepted readDir failure")
	}
	readDir = origReadDir
	readFile = func(string) ([]byte, error) { return nil, errors.New("readfile") }
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.md"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := LoadCustom(dir); err == nil {
		t.Fatal("LoadCustom() accepted readFile failure")
	}
	if _, err := Resolve("x", dir); err == nil {
		t.Fatal("Resolve() accepted custom load failure")
	}
	if _, err := Names(dir); err == nil {
		t.Fatal("Names() accepted custom load failure")
	}
	if _, err := ListMarkdown(nil, dir); err == nil {
		t.Fatal("ListMarkdown() accepted custom load failure")
	}
	mkdirAll = origMkdirAll
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("writefile") }
	body := "# Rules\nUse this skill when writing one custom workflow. Stop condition: done with evidence.\n"
	if _, err := ScaffoldCustomSkill(t.TempDir(), "custom", body); err == nil {
		t.Fatal("ScaffoldCustomSkill() accepted write failure")
	}
}

func TestClassifyAndSkillNewDraft(t *testing.T) {
	full := strings.Repeat("# Workflow\nRun build and test phases with cmake.\n", 130)
	classification := Classify("full", full)
	if classification.Class != FullWorkflow || classification.Confidence < 0 || classification.Confidence > 1 || len(classification.Reasoning) == 0 {
		t.Fatalf("Classify(full) = %+v", classification)
	}
	bridge := "# Setup\nUse this skill when MCP tools connect to debug targets.\n| Operation | Tool |\n|---|---|\n"
	if got := Classify("bridge", bridge); got.Class != ToolBridge {
		t.Fatalf("Classify(bridge) = %+v, want tool bridge", got)
	}
	micro := "# Rules\nUse this skill when reviewing one file. Stop condition: one concern is resolved.\n"
	if got := Classify("micro", micro); got.Class != MicroSkill {
		t.Fatalf("Classify(micro) = %+v, want micro", got)
	}
	if got := Classify("tiny", "short"); got.Class != NoneClass || got.Confidence != 1 {
		t.Fatalf("Classify(tiny) = %+v, want none confidence 1", got)
	}
	noSignals := strings.Repeat("plain words without classifier markers ", 5)
	if got := Classify("nosignals", noSignals); got.Class != NoneClass || got.Confidence != 1 {
		t.Fatalf("Classify(no signals) = %+v, want none confidence 1", got)
	}
	weak := strings.Repeat("recomp evidence ", 10)
	if got := Classify("weak", weak); got.Class != NoneClass || got.Reasoning[0] != "content signals present but weakly separated" {
		t.Fatalf("Classify(weak) = %+v, want weak none classification", got)
	}
	class, high, second := winningClass(map[TemplateClass]int{FullWorkflow: 3, MicroSkill: 2, ToolBridge: 1})
	if class != FullWorkflow || high != 3 || second != 2 {
		t.Fatalf("winningClass() = %s, %d, %d; want full, 3, 2", class, high, second)
	}
	draft, err := NewDraft("https://example.com/path/My Skill.md", bridge)
	if err != nil {
		t.Fatalf("NewDraft() error = %v", err)
	}
	if draft.Name != "my-skill" || draft.TargetPath != ".rehamr/skills/my-skill.md" || len(draft.Instructions) != 4 || draft.Classification.Class != ToolBridge {
		t.Fatalf("NewDraft() = %+v", draft)
	}
	if _, err := NewDraft("://bad", bridge); err == nil {
		t.Fatal("NewDraft() accepted invalid URL")
	}
	if _, err := NewDraft("https://example.com/x.md", "short"); err == nil {
		t.Fatal("NewDraft() accepted short body")
	}
}

func TestScaffoldCustomSkill(t *testing.T) {
	body := "# Rules\nUse this skill when writing one custom workflow. Stop condition: done with evidence.\n"
	dir := t.TempDir()
	path, err := ScaffoldCustomSkill(dir, "My Skill.md", body)
	if err != nil {
		t.Fatalf("ScaffoldCustomSkill() error = %v", err)
	}
	if filepath.Base(path) != "my-skill.md" {
		t.Fatalf("ScaffoldCustomSkill() path = %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != body {
		t.Fatalf("scaffold read = %q, %v", string(data), err)
	}
	for _, tc := range []struct {
		dir  string
		name string
		body string
	}{
		{"", "x", body},
		{dir, "", body},
		{dir, "custom", "short"},
	} {
		if _, err := ScaffoldCustomSkill(tc.dir, tc.name, tc.body); err == nil {
			t.Fatalf("ScaffoldCustomSkill(%q, %q) accepted invalid input", tc.dir, tc.name)
		}
	}
	file := filepath.Join(t.TempDir(), "not-dir")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := ScaffoldCustomSkill(filepath.Join(file, "child"), "custom", body); err == nil {
		t.Fatal("ScaffoldCustomSkill() accepted mkdir failure")
	}
}

func TestSkillNameFallbacks(t *testing.T) {
	draft, err := NewDraft("https://example.com/", "# Rules\nUse this skill when host path has no segment. Stop condition: done with evidence.\n")
	if err != nil {
		t.Fatalf("NewDraft(host fallback) error = %v", err)
	}
	if draft.Name != "example-com" {
		t.Fatalf("host fallback name = %q", draft.Name)
	}
	shortDraft, err := NewDraft("https://example.com/!.md", "# Rules\nUse this skill when the path sanitizes away. Stop condition: done with evidence.\n")
	if err != nil {
		t.Fatalf("NewDraft(short fallback) error = %v", err)
	}
	if shortDraft.Name != "new-skill" {
		t.Fatalf("short fallback name = %q", shortDraft.Name)
	}
}

func sorted(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i-1] > values[i] {
			return false
		}
	}
	return true
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

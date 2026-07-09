package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"recomphamr2/internal/config"
	"recomphamr2/internal/project"
	"recomphamr2/internal/release"
)

func TestRunReportsUnsupportedWithoutWorkspace(t *testing.T) {
	report := Run(Options{ProjectDir: t.TempDir()})
	out := report.String()
	for _, want := range []string{
		"RecompHamr doctor",
		"[verified] runtime:",
		"[unsupported] workspace: .rehamr workspace is not initialized",
		"[unsupported] config: .rehamr/config.yaml is not initialized",
		"[unsupported] memory: .rehamr workspace is not initialized",
		"[verified] tools: primary=6 compatibility=1",
		"[blocked] install-update-release:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestRunReportsOperationalFiles(t *testing.T) {
	dir := t.TempDir()
	writeOperationalFiles(t, dir)
	out := Run(Options{ProjectDir: dir}).String()
	if !strings.Contains(out, "[verified] install-update-release: 5 operational files verified") {
		t.Fatalf("doctor operational output:\n%s", out)
	}
}

func TestRunReportsInitializedWorkspace(t *testing.T) {
	dir := t.TempDir()
	ws, err := project.Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	customDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(customDir, "custom.md"), []byte("# Custom"), 0o600); err != nil {
		t.Fatalf("WriteFile() custom error = %v", err)
	}
	if err := os.WriteFile(ws.State, []byte(strings.Repeat("abcd", 1200)+"é-tail"), 0o600); err != nil {
		t.Fatalf("WriteFile() memory error = %v", err)
	}
	t.Setenv("RECOMPHAMR_MCP_AUTOSTART", "1")
	out := Run(Options{ProjectDir: dir, CustomSkillsDir: customDir}).String()
	for _, want := range []string{
		"[verified] workspace:",
		"[verified] config: active=lmstudio-amd profiles=4",
		"[verified] memory:",
		"truncated=true",
		"[verified] skills: embedded=28 custom=1",
		"[verified] mcp: registered=8 autostart=8",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor initialized output missing %q:\n%s", want, out)
		}
	}
}

func TestRunReportsBlockedStates(t *testing.T) {
	workspaceFile := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspaceFile, config.DirName), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() workspace blocker error = %v", err)
	}
	out := Run(Options{ProjectDir: workspaceFile, CustomSkillsDir: string(rune(0))}).String()
	for _, want := range []string{
		"[blocked] workspace: .rehamr exists but is not a directory",
		"[blocked] config:",
		"[blocked] skills:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor blocked output missing %q:\n%s", want, out)
		}
	}
}

func TestRunReportsMissingMemory(t *testing.T) {
	dir := t.TempDir()
	ws, err := project.Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.Remove(ws.State); err != nil {
		t.Fatalf("Remove() memory error = %v", err)
	}
	out := Run(Options{ProjectDir: dir}).String()
	if !strings.Contains(out, "[unsupported] memory: REPHAMR_STATE.md is missing") {
		t.Fatalf("doctor missing memory output:\n%s", out)
	}
}

func TestRunReportsConfigFileStates(t *testing.T) {
	missingConfig := t.TempDir()
	if err := os.Mkdir(filepath.Join(missingConfig, config.DirName), 0o700); err != nil {
		t.Fatalf("Mkdir() .rehamr error = %v", err)
	}
	out := Run(Options{ProjectDir: missingConfig}).String()
	if !strings.Contains(out, "[unsupported] config: .rehamr/config.yaml is not initialized") {
		t.Fatalf("doctor missing config output:\n%s", out)
	}

	badConfig := t.TempDir()
	if err := os.Mkdir(filepath.Join(badConfig, config.DirName), 0o700); err != nil {
		t.Fatalf("Mkdir() bad .rehamr error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(badConfig, config.DirName, config.FileName), []byte("active: [bad"), 0o600); err != nil {
		t.Fatalf("WriteFile() bad config error = %v", err)
	}
	out = Run(Options{ProjectDir: badConfig}).String()
	if !strings.Contains(out, "[blocked] config:") {
		t.Fatalf("doctor bad config output:\n%s", out)
	}
}

func TestRunReportsInvalidProjectDir(t *testing.T) {
	out := Run(Options{ProjectDir: string(rune(0))}).String()
	for _, want := range []string{
		"[blocked] workspace:",
		"[blocked] config:",
		"[unsupported] memory: .rehamr workspace is not initialized",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor invalid project output missing %q:\n%s", want, out)
		}
	}
}

func TestRunReportsMemoryReadBlocked(t *testing.T) {
	dir := t.TempDir()
	ws, err := project.Init(dir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := os.Remove(ws.State); err != nil {
		t.Fatalf("Remove() memory error = %v", err)
	}
	if err := os.Mkdir(ws.State, 0o700); err != nil {
		t.Fatalf("Mkdir() memory blocker error = %v", err)
	}
	out := Run(Options{ProjectDir: dir}).String()
	if !strings.Contains(out, "[blocked] memory:") {
		t.Fatalf("doctor memory blocked output:\n%s", out)
	}
}

func TestRunDefaultsProjectDir(t *testing.T) {
	out := Run(Options{}).String()
	if !strings.Contains(out, "RecompHamr doctor") || strings.HasSuffix(out, "\n") {
		t.Fatalf("doctor default output = %q", out)
	}
}

func writeOperationalFiles(t *testing.T, dir string) {
	t.Helper()
	for _, file := range release.OperationalFiles() {
		path := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() operational file error = %v", err)
		}
		if err := os.WriteFile(path, []byte(strings.Join(file.Markers, "\n")), 0o600); err != nil {
			t.Fatalf("WriteFile() operational file error = %v", err)
		}
	}
}

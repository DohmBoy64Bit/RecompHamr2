package doctor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"recomphamr2/internal/config"
	"recomphamr2/internal/mcp"
	"recomphamr2/internal/project"
	"recomphamr2/internal/release"
	"recomphamr2/internal/skills"
	"recomphamr2/internal/tools"
)

// Status identifies the evidence level for one diagnostic check.
type Status string

const (
	// StatusVerified means the check was proven from local runtime or files.
	StatusVerified Status = "verified"
	// StatusUnsupported means the check is intentionally outside this phase.
	StatusUnsupported Status = "unsupported"
	// StatusBlocked means local state prevented a definitive check result.
	StatusBlocked Status = "blocked"
)

// Check records one doctor diagnostic result.
type Check struct {
	// Name is the stable diagnostic label.
	Name string
	// Status is the evidence level for Detail.
	Status Status
	// Detail explains the local evidence or missing capability.
	Detail string
}

// Options controls doctor diagnostics.
type Options struct {
	// ProjectDir is the repository or project directory containing .rehamr.
	ProjectDir string
	// CustomSkillsDir is the optional directory for user-authored skills.
	CustomSkillsDir string
}

// Report is a complete doctor diagnostic snapshot.
type Report struct {
	// Checks is the ordered diagnostic result list.
	Checks []Check
}

// Run collects deterministic local diagnostics without network or process probes.
func Run(opts Options) Report {
	projectDir := strings.TrimSpace(opts.ProjectDir)
	if projectDir == "" {
		projectDir = "."
	}
	checks := []Check{
		{Name: "runtime", Status: StatusVerified, Detail: fmt.Sprintf("go=%s os=%s arch=%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)},
		workspaceCheck(projectDir),
		configCheck(projectDir),
		memoryCheck(projectDir),
		skillsCheck(opts.CustomSkillsDir),
		toolsCheck(),
		mcpCheck(),
		operationalCheck(projectDir),
	}
	return Report{Checks: checks}
}

// String renders the report for slash-command output.
func (r Report) String() string {
	var b strings.Builder
	b.WriteString("RecompHamr doctor\n")
	for _, check := range r.Checks {
		fmt.Fprintf(&b, "[%s] %s: %s\n", check.Status, check.Name, check.Detail)
	}
	return strings.TrimRight(b.String(), "\n")
}

func workspaceCheck(projectDir string) Check {
	root := filepath.Join(projectDir, config.DirName)
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Check{Name: "workspace", Status: StatusUnsupported, Detail: ".rehamr workspace is not initialized"}
		}
		return Check{Name: "workspace", Status: StatusBlocked, Detail: err.Error()}
	}
	if !info.IsDir() {
		return Check{Name: "workspace", Status: StatusBlocked, Detail: ".rehamr exists but is not a directory"}
	}
	return Check{Name: "workspace", Status: StatusVerified, Detail: root}
}

func configCheck(projectDir string) Check {
	root := filepath.Join(projectDir, config.DirName)
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Check{Name: "config", Status: StatusUnsupported, Detail: ".rehamr/config.yaml is not initialized"}
		}
		return Check{Name: "config", Status: StatusBlocked, Detail: err.Error()}
	}
	if !info.IsDir() {
		return Check{Name: "config", Status: StatusBlocked, Detail: ".rehamr exists but is not a directory"}
	}
	cfg, err := config.Load(filepath.Join(projectDir, config.DirName, config.FileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Check{Name: "config", Status: StatusUnsupported, Detail: ".rehamr/config.yaml is not initialized"}
		}
		return Check{Name: "config", Status: StatusBlocked, Detail: err.Error()}
	}
	return Check{Name: "config", Status: StatusVerified, Detail: fmt.Sprintf("active=%s profiles=%d", cfg.Active, len(cfg.Models))}
}

func memoryCheck(projectDir string) Check {
	mem, err := project.LoadMemory(projectDir, 4096)
	if err != nil {
		if errors.Is(err, project.ErrWorkspaceMissing) || errors.Is(err, project.ErrMemoryMissing) {
			return Check{Name: "memory", Status: StatusUnsupported, Detail: err.Error()}
		}
		return Check{Name: "memory", Status: StatusBlocked, Detail: err.Error()}
	}
	detail := fmt.Sprintf("%s bytes=%d", mem.Path, len(mem.Content))
	if mem.Truncated {
		detail += " truncated=true"
	}
	return Check{Name: "memory", Status: StatusVerified, Detail: detail}
}

func skillsCheck(customDir string) Check {
	custom, err := skills.LoadCustom(customDir)
	if err != nil {
		return Check{Name: "skills", Status: StatusBlocked, Detail: err.Error()}
	}
	return Check{Name: "skills", Status: StatusVerified, Detail: fmt.Sprintf("embedded=%d custom=%d", len(skills.Embedded()), len(custom))}
}

func toolsCheck() Check {
	return Check{Name: "tools", Status: StatusVerified, Detail: fmt.Sprintf("primary=%d compatibility=%d", len(tools.Schemas()), len(tools.CompatibilityToolNames()))}
}

func mcpCheck() Check {
	servers := mcp.Builtins()
	autostart := 0
	for _, server := range servers {
		if server.Autostart {
			autostart++
		}
	}
	return Check{Name: "mcp", Status: StatusVerified, Detail: fmt.Sprintf("registered=%d autostart=%d", len(servers), autostart)}
}

func operationalCheck(projectDir string) Check {
	results := release.ValidateOperationalFiles(projectDir)
	blocked := 0
	for _, result := range results {
		if result.Status == release.StatusBlocked {
			blocked++
		}
	}
	if blocked > 0 {
		return Check{Name: "install-update-release", Status: StatusBlocked, Detail: fmt.Sprintf("%d/%d operational files blocked", blocked, len(results))}
	}
	return Check{Name: "install-update-release", Status: StatusVerified, Detail: fmt.Sprintf("%d operational files verified", len(results))}
}

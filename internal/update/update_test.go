package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"recomphamr2/internal/release"
)

func TestPlanLocal(t *testing.T) {
	dir := t.TempDir()
	artifact := "recomphamr_windows_amd64.zip"
	if err := os.WriteFile(filepath.Join(dir, artifact), []byte("zip"), 0o600); err != nil {
		t.Fatalf("WriteFile() artifact error = %v", err)
	}
	manifest, err := release.WriteManifest(dir, []string{artifact}, "")
	if err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	plan, err := PlanLocal(PlanSpec{
		RootDir:          dir,
		ManifestPath:     manifest,
		ArtifactPath:     artifact,
		CurrentVersion:   "v1.0.0",
		CandidateVersion: "v1.1.0",
	})
	if err != nil {
		t.Fatalf("PlanLocal() error = %v", err)
	}
	if plan.ArtifactPath != artifact || !strings.Contains(plan.String(), "v1.0.0 -> v1.1.0") {
		t.Fatalf("PlanLocal() plan = %#v string=%q", plan, plan.String())
	}
}

func TestPlanLocalErrors(t *testing.T) {
	if _, err := PlanLocal(PlanSpec{}); !errors.Is(err, ErrMissingVersion) {
		t.Fatalf("PlanLocal(missing versions) error = %v", err)
	}
	if _, err := PlanLocal(PlanSpec{CurrentVersion: "v1", CandidateVersion: "v1"}); !errors.Is(err, ErrSameVersion) {
		t.Fatalf("PlanLocal(same version) error = %v", err)
	}
	dir := t.TempDir()
	if _, err := PlanLocal(PlanSpec{RootDir: dir, ManifestPath: filepath.Join(dir, "missing"), ArtifactPath: "x.zip", CurrentVersion: "v1", CandidateVersion: "v2"}); err == nil {
		t.Fatal("PlanLocal() accepted missing manifest")
	}
	if err := os.WriteFile(filepath.Join(dir, "a.zip"), []byte("a"), 0o600); err != nil {
		t.Fatalf("WriteFile() a.zip error = %v", err)
	}
	manifest, err := release.WriteManifest(dir, []string{"a.zip"}, "")
	if err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	if _, err := PlanLocal(PlanSpec{RootDir: dir, ManifestPath: manifest, ArtifactPath: "missing.zip", CurrentVersion: "v1", CandidateVersion: "v2"}); !errors.Is(err, ErrArtifactUnverified) {
		t.Fatalf("PlanLocal(missing artifact entry) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.zip"), []byte("changed"), 0o600); err != nil {
		t.Fatalf("WriteFile() changed artifact error = %v", err)
	}
	if _, err := PlanLocal(PlanSpec{RootDir: dir, ManifestPath: manifest, ArtifactPath: "a.zip", CurrentVersion: "v1", CandidateVersion: "v2"}); !errors.Is(err, ErrArtifactUnverified) {
		t.Fatalf("PlanLocal(mismatch) error = %v", err)
	}
}

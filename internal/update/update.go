package update

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"recomphamr2/internal/release"
)

// ErrMissingVersion reports that a current or candidate version was omitted.
var ErrMissingVersion = errors.New("update versions must be provided")

// ErrSameVersion reports that the candidate version matches the current version.
var ErrSameVersion = errors.New("candidate version matches current version")

// ErrArtifactUnverified reports that the requested artifact did not verify.
var ErrArtifactUnverified = errors.New("update artifact is not verified")

// PlanSpec describes a local self-update dry-run request.
type PlanSpec struct {
	// RootDir is the release directory containing artifacts.
	RootDir string
	// ManifestPath is the SHA256SUMS manifest to verify.
	ManifestPath string
	// ArtifactPath is the release-directory-relative artifact selected for update.
	ArtifactPath string
	// CurrentVersion is the currently running version string.
	CurrentVersion string
	// CandidateVersion is the version string from the selected local artifact.
	CandidateVersion string
}

// Plan describes a verified local self-update dry-run.
type Plan struct {
	// CurrentVersion is the version being replaced.
	CurrentVersion string
	// CandidateVersion is the proposed replacement version.
	CandidateVersion string
	// ArtifactPath is the verified artifact path relative to RootDir.
	ArtifactPath string
	// Detail explains the verified dry-run result.
	Detail string
}

// PlanLocal verifies a local artifact manifest entry and returns a dry-run update plan.
func PlanLocal(spec PlanSpec) (Plan, error) {
	current := strings.TrimSpace(spec.CurrentVersion)
	candidate := strings.TrimSpace(spec.CandidateVersion)
	if current == "" || candidate == "" {
		return Plan{}, ErrMissingVersion
	}
	if current == candidate {
		return Plan{}, ErrSameVersion
	}
	artifact := filepath.ToSlash(strings.TrimSpace(spec.ArtifactPath))
	report, err := release.VerifyManifest(spec.RootDir, spec.ManifestPath)
	if err != nil {
		return Plan{}, err
	}
	for _, result := range report.Results {
		if result.Entry.Path != artifact {
			continue
		}
		if result.Status != release.StatusVerified {
			return Plan{}, fmt.Errorf("%w: %s", ErrArtifactUnverified, result.Detail)
		}
		return Plan{
			CurrentVersion:   current,
			CandidateVersion: candidate,
			ArtifactPath:     artifact,
			Detail:           "verified local artifact; replacement must be performed by an installer or explicit user action",
		}, nil
	}
	return Plan{}, fmt.Errorf("%w: artifact missing from manifest", ErrArtifactUnverified)
}

// String renders a user-facing dry-run update plan.
func (p Plan) String() string {
	return fmt.Sprintf("update dry-run: %s -> %s using %s\n%s", p.CurrentVersion, p.CandidateVersion, p.ArtifactPath, p.Detail)
}

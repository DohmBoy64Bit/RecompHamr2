package release

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	readFile  = os.ReadFile
	mkdirAll  = os.MkdirAll
	osEnviron = os.Environ
	runCmd    = runBuildCommand
	statFile  = os.Stat
	writeFile = os.WriteFile
)

// Status identifies the result for one manifest entry.
type Status string

const (
	// StatusVerified means the artifact hash matched its manifest entry.
	StatusVerified Status = "verified"
	// StatusBlocked means the entry could not be checked safely.
	StatusBlocked Status = "blocked"
)

// ErrEmptyManifest reports a checksum manifest with no artifact entries.
var ErrEmptyManifest = errors.New("checksum manifest has no entries")

// ErrUnsupportedTarget reports an unsupported release OS, architecture, or archive format.
var ErrUnsupportedTarget = errors.New("unsupported release target")

// ErrArchiveExists reports that an archive path already exists.
var ErrArchiveExists = errors.New("release archive already exists")

// ErrBinaryExists reports that a build output path already exists.
var ErrBinaryExists = errors.New("release binary already exists")

const (
	// BinaryName is the executable base name used in release archives.
	BinaryName = "recomphamr"
	// ManifestName is the canonical checksum manifest filename.
	ManifestName = "SHA256SUMS"
)

// Target describes one supported release artifact target.
type Target struct {
	// OS is the Go GOOS value for the artifact.
	OS string
	// Arch is the Go GOARCH value for the artifact.
	Arch string
	// Archive is the archive extension without a leading dot.
	Archive string
}

// BuildSpec describes one local Go build request.
type BuildSpec struct {
	// Target is the supported release target to build.
	Target Target
	// Package is the Go package path to build; empty uses ./cmd/recomphamr.
	Package string
	// OutputDir is the directory that receives the built binary.
	OutputDir string
}

// ArchiveSpec describes one local archive creation request.
type ArchiveSpec struct {
	// Target is the supported release target to archive.
	Target Target
	// BinaryPath is the already-built executable to place in the archive.
	BinaryPath string
	// OutputDir is the directory that receives the archive.
	OutputDir string
}

// Entry describes one SHA-256 manifest row.
type Entry struct {
	// Path is the relative artifact path from the manifest.
	Path string
	// SHA256 is the lowercase expected SHA-256 hex digest.
	SHA256 string
}

// Result describes verification for one artifact entry.
type Result struct {
	// Entry is the manifest row that was checked.
	Entry Entry
	// Status is the verification outcome.
	Status Status
	// ActualSHA256 is populated when the artifact could be read.
	ActualSHA256 string
	// Detail explains mismatches, missing files, or unsafe paths.
	Detail string
}

// Report is the complete checksum verification result.
type Report struct {
	// ManifestPath is the checksum manifest file that was read.
	ManifestPath string
	// Results is the ordered result list for manifest entries.
	Results []Result
}

// OperationalFile describes a required Phase 12 release or install file.
type OperationalFile struct {
	// Path is the repository-relative file path.
	Path string
	// Purpose explains the operational behavior covered by the file.
	Purpose string
	// Markers are substrings that must exist in the file for local validation.
	Markers []string
}

// OperationalResult describes local validation for one operational file.
type OperationalResult struct {
	// File is the operational file that was checked.
	File OperationalFile
	// Status is the validation outcome.
	Status Status
	// Detail explains the local evidence or missing marker.
	Detail string
}

// DefaultTargets returns the supported release artifact targets.
func DefaultTargets() []Target {
	return []Target{
		{OS: "windows", Arch: "amd64", Archive: "zip"},
		{OS: "windows", Arch: "arm64", Archive: "zip"},
		{OS: "linux", Arch: "amd64", Archive: "tar.gz"},
		{OS: "linux", Arch: "arm64", Archive: "tar.gz"},
		{OS: "darwin", Arch: "amd64", Archive: "tar.gz"},
		{OS: "darwin", Arch: "arm64", Archive: "tar.gz"},
	}
}

// OperationalFiles returns the required Phase 12 install, release, devcontainer, and CI files.
func OperationalFiles() []OperationalFile {
	return []OperationalFile{
		{Path: "scripts/install.ps1", Purpose: "Windows local archive installer with SHA256SUMS verification", Markers: []string{"param(", "Get-FileHash", "Expand-Archive", "SHA256SUMS"}},
		{Path: "scripts/install.sh", Purpose: "POSIX local archive installer with SHA256SUMS verification", Markers: []string{"set -eu", "sha256sum", "tar -xzf", "SHA256SUMS"}},
		{Path: ".goreleaser.yaml", Purpose: "GoReleaser release archive and checksum configuration", Markers: []string{"builds:", "archives:", "checksum:", "SHA256SUMS"}},
		{Path: ".devcontainer/devcontainer.json", Purpose: "Go devcontainer that runs make verify after creation", Markers: []string{"devcontainers/go", "make verify"}},
		{Path: ".github/workflows/verify.yml", Purpose: "CI workflow that runs make verify on Linux, Windows, and macOS", Markers: []string{"ubuntu-latest", "windows-latest", "macos-latest", "make verify"}},
	}
}

// ValidateOperationalFiles checks required Phase 12 operational files under repoRoot.
func ValidateOperationalFiles(repoRoot string) []OperationalResult {
	files := OperationalFiles()
	results := make([]OperationalResult, 0, len(files))
	for _, file := range files {
		results = append(results, validateOperationalFile(repoRoot, file))
	}
	return results
}

// ArtifactName returns the canonical archive name for target.
func ArtifactName(target Target) (string, error) {
	target = normalizeTarget(target)
	if !supportedTarget(target) {
		return "", ErrUnsupportedTarget
	}
	return fmt.Sprintf("%s_%s_%s.%s", BinaryName, target.OS, target.Arch, target.Archive), nil
}

// BinaryFileName returns the executable filename stored inside target archives.
func BinaryFileName(target Target) (string, error) {
	target = normalizeTarget(target)
	if !supportedTarget(target) {
		return "", ErrUnsupportedTarget
	}
	if target.OS == "windows" {
		return BinaryName + ".exe", nil
	}
	return BinaryName, nil
}

// BinaryOutputName returns the deterministic local build output filename.
func BinaryOutputName(target Target) (string, error) {
	target = normalizeTarget(target)
	if !supportedTarget(target) {
		return "", ErrUnsupportedTarget
	}
	name := fmt.Sprintf("%s_%s_%s", BinaryName, target.OS, target.Arch)
	if target.OS == "windows" {
		name += ".exe"
	}
	return name, nil
}

// ManifestEntries returns empty manifest entries for all supported artifact names.
func ManifestEntries() []Entry {
	targets := DefaultTargets()
	entries := make([]Entry, 0, len(targets))
	for _, target := range targets {
		name, _ := ArtifactName(target)
		entries = append(entries, Entry{Path: name})
	}
	return entries
}

// GenerateManifest returns SHA-256 manifest entries for local artifact paths under rootDir.
func GenerateManifest(rootDir string, artifactPaths []string) ([]Entry, error) {
	entries := make([]Entry, 0, len(artifactPaths))
	for _, artifactPath := range artifactPaths {
		rel := filepath.ToSlash(strings.TrimSpace(artifactPath))
		if unsafeArtifactPath(rel) {
			return nil, fmt.Errorf("artifact path must be relative and stay inside release directory: %s", artifactPath)
		}
		data, err := readFile(filepath.Join(rootDir, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, Entry{Path: rel, SHA256: hex.EncodeToString(sum[:])})
	}
	if len(entries) == 0 {
		return nil, ErrEmptyManifest
	}
	sortEntries(entries)
	return entries, nil
}

// ManifestText renders sorted SHA256SUMS text from entries.
func ManifestText(entries []Entry) (string, error) {
	if len(entries) == 0 {
		return "", ErrEmptyManifest
	}
	sorted := append([]Entry(nil), entries...)
	sortEntries(sorted)
	var b strings.Builder
	for _, entry := range sorted {
		digest := strings.ToLower(strings.TrimSpace(entry.SHA256))
		if !validSHA256(digest) {
			return "", fmt.Errorf("sha256 for %s must be %d hex characters", entry.Path, sha256.Size*2)
		}
		path := filepath.ToSlash(strings.TrimSpace(entry.Path))
		if unsafeArtifactPath(path) {
			return "", fmt.Errorf("artifact path must be relative and stay inside release directory: %s", entry.Path)
		}
		fmt.Fprintf(&b, "%s  %s\n", digest, path)
	}
	return b.String(), nil
}

// WriteManifest writes a SHA256SUMS manifest for artifactPaths and returns the written manifest path.
func WriteManifest(rootDir string, artifactPaths []string, manifestPath string) (string, error) {
	entries, err := GenerateManifest(rootDir, artifactPaths)
	if err != nil {
		return "", err
	}
	text, _ := ManifestText(entries)
	out := strings.TrimSpace(manifestPath)
	if out == "" {
		out = filepath.Join(rootDir, ManifestName)
	}
	if err := writeFile(out, []byte(text), 0o644); err != nil {
		return "", err
	}
	return out, nil
}

// BuildBinary runs go build for a supported release target.
func BuildBinary(spec BuildSpec) (string, error) {
	target := normalizeTarget(spec.Target)
	outputName, err := BinaryOutputName(target)
	if err != nil {
		return "", err
	}
	pkg := strings.TrimSpace(spec.Package)
	if pkg == "" {
		pkg = "./cmd/recomphamr"
	}
	outputDir := strings.TrimSpace(spec.OutputDir)
	if outputDir == "" {
		outputDir = "dist"
	}
	if err := mkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	outputPath := filepath.Join(outputDir, outputName)
	if _, err := statFile(outputPath); err == nil {
		return "", ErrBinaryExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	args := []string{"build", "-trimpath", "-o", outputPath, pkg}
	env := []string{"GOOS=" + target.OS, "GOARCH=" + target.Arch, "CGO_ENABLED=0"}
	if err := runCmd("go", args, env); err != nil {
		return "", err
	}
	return outputPath, nil
}

// CreateArchive creates a release archive from an already-built binary.
func CreateArchive(spec ArchiveSpec) (string, error) {
	target := normalizeTarget(spec.Target)
	archiveName, err := ArtifactName(target)
	if err != nil {
		return "", err
	}
	binaryName, _ := BinaryFileName(target)
	data, err := readFile(spec.BinaryPath)
	if err != nil {
		return "", err
	}
	outputDir := strings.TrimSpace(spec.OutputDir)
	if outputDir == "" {
		outputDir = "."
	}
	if err := mkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	outputPath := filepath.Join(outputDir, archiveName)
	if _, err := statFile(outputPath); err == nil {
		return "", ErrArchiveExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	archive, _ := ArchiveBytes(target, binaryName, data)
	if err := writeFile(outputPath, archive, 0o644); err != nil {
		return "", err
	}
	return outputPath, nil
}

func runBuildCommand(name string, args []string, env []string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = append(osEnviron(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(out))
		if text == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, text)
	}
	return nil
}

// ArchiveBytes returns archive bytes for binaryName and binary data.
func ArchiveBytes(target Target, binaryName string, data []byte) ([]byte, error) {
	target = normalizeTarget(target)
	if !supportedTarget(target) {
		return nil, ErrUnsupportedTarget
	}
	if strings.TrimSpace(binaryName) == "" || unsafeArtifactPath(binaryName) {
		return nil, fmt.Errorf("archive binary name must be a safe relative path")
	}
	switch target.Archive {
	case "zip":
		return zipBytes(binaryName, data)
	}
	return tarGzBytes(binaryName, data)
}

// Verified reports whether every manifest entry matched.
func (r Report) Verified() bool {
	return len(r.Results) > 0 && r.Count(StatusBlocked) == 0
}

func normalizeTarget(target Target) Target {
	return Target{
		OS:      strings.ToLower(strings.TrimSpace(target.OS)),
		Arch:    strings.ToLower(strings.TrimSpace(target.Arch)),
		Archive: strings.TrimPrefix(strings.ToLower(strings.TrimSpace(target.Archive)), "."),
	}
}

func supportedTarget(target Target) bool {
	for _, supported := range DefaultTargets() {
		if target == supported {
			return true
		}
	}
	return false
}

func zipBytes(binaryName string, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create(binaryName)
	_, _ = w.Write(data)
	_ = zw.Close()
	return buf.Bytes(), nil
}

func tarGzBytes(binaryName string, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	_ = tw.WriteHeader(&tar.Header{Name: binaryName, Mode: 0o755, Size: int64(len(data))})
	_, _ = tw.Write(data)
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes(), nil
}

func sortEntries(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
}

func validSHA256(digest string) bool {
	if len(digest) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil
}

func validateOperationalFile(repoRoot string, file OperationalFile) OperationalResult {
	if unsafeArtifactPath(file.Path) {
		return OperationalResult{File: file, Status: StatusBlocked, Detail: "operational file path must be relative and stay inside repository"}
	}
	data, err := readFile(filepath.Join(repoRoot, filepath.FromSlash(file.Path)))
	if err != nil {
		return OperationalResult{File: file, Status: StatusBlocked, Detail: err.Error()}
	}
	text := string(data)
	if strings.TrimSpace(text) == "" {
		return OperationalResult{File: file, Status: StatusBlocked, Detail: "file is empty"}
	}
	for _, marker := range file.Markers {
		if !strings.Contains(text, marker) {
			return OperationalResult{File: file, Status: StatusBlocked, Detail: "missing required marker: " + marker}
		}
	}
	return OperationalResult{File: file, Status: StatusVerified, Detail: file.Purpose}
}

// Count returns the number of results with status.
func (r Report) Count(status Status) int {
	total := 0
	for _, result := range r.Results {
		if result.Status == status {
			total++
		}
	}
	return total
}

// String renders a user-facing release verification report.
func (r Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "release verification: %s\n", r.ManifestPath)
	for _, result := range r.Results {
		detail := result.Detail
		if detail == "" {
			detail = "sha256 matched"
		}
		fmt.Fprintf(&b, "[%s] %s: %s\n", result.Status, result.Entry.Path, detail)
	}
	return strings.TrimRight(b.String(), "\n")
}

// ParseManifest parses SHA-256 checksum manifest text.
func ParseManifest(text string) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("line %d: expected '<sha256> <path>'", lineNo)
		}
		digest := strings.ToLower(fields[0])
		if len(digest) != sha256.Size*2 {
			return nil, fmt.Errorf("line %d: sha256 must be %d hex characters", lineNo, sha256.Size*2)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return nil, fmt.Errorf("line %d: sha256 is not hex", lineNo)
		}
		path := strings.TrimPrefix(fields[1], "*")
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("line %d: path is empty", lineNo)
		}
		entries = append(entries, Entry{Path: filepath.ToSlash(path), SHA256: digest})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, ErrEmptyManifest
	}
	return entries, nil
}

// VerifyManifest reads manifestPath and verifies each listed artifact under rootDir.
func VerifyManifest(rootDir string, manifestPath string) (Report, error) {
	data, err := readFile(manifestPath)
	if err != nil {
		return Report{ManifestPath: manifestPath}, err
	}
	entries, err := ParseManifest(string(data))
	if err != nil {
		return Report{ManifestPath: manifestPath}, err
	}
	report := Report{ManifestPath: manifestPath, Results: make([]Result, 0, len(entries))}
	for _, entry := range entries {
		report.Results = append(report.Results, verifyEntry(rootDir, entry))
	}
	return report, nil
}

func verifyEntry(rootDir string, entry Entry) Result {
	if unsafeArtifactPath(entry.Path) {
		return Result{Entry: entry, Status: StatusBlocked, Detail: "artifact path must be relative and stay inside release directory"}
	}
	path := filepath.Join(rootDir, filepath.FromSlash(entry.Path))
	data, err := readFile(path)
	if err != nil {
		return Result{Entry: entry, Status: StatusBlocked, Detail: err.Error()}
	}
	actual := sha256.Sum256(data)
	actualHex := hex.EncodeToString(actual[:])
	if actualHex != entry.SHA256 {
		return Result{Entry: entry, Status: StatusBlocked, ActualSHA256: actualHex, Detail: "sha256 mismatch"}
	}
	return Result{Entry: entry, Status: StatusVerified, ActualSHA256: actualHex}
}

func unsafeArtifactPath(path string) bool {
	if filepath.IsAbs(path) || strings.HasPrefix(path, "/") || strings.HasPrefix(path, `\`) {
		return true
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	return clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

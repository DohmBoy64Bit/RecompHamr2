package release

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	sum := digest("artifact")
	entries, err := ParseManifest("# release\n\n" + sum + " *recomphamr_windows_amd64.zip\n")
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}
	if len(entries) != 1 || entries[0].SHA256 != sum || entries[0].Path != "recomphamr_windows_amd64.zip" {
		t.Fatalf("ParseManifest() = %#v", entries)
	}
}

func TestArtifactNames(t *testing.T) {
	want := []string{
		"recomphamr_windows_amd64.zip",
		"recomphamr_windows_arm64.zip",
		"recomphamr_linux_amd64.tar.gz",
		"recomphamr_linux_arm64.tar.gz",
		"recomphamr_darwin_amd64.tar.gz",
		"recomphamr_darwin_arm64.tar.gz",
	}
	targets := DefaultTargets()
	if len(targets) != len(want) {
		t.Fatalf("DefaultTargets() len = %d, want %d", len(targets), len(want))
	}
	for i, target := range targets {
		got, err := ArtifactName(target)
		if err != nil {
			t.Fatalf("ArtifactName(%#v) error = %v", target, err)
		}
		if got != want[i] {
			t.Fatalf("ArtifactName(%#v) = %q, want %q", target, got, want[i])
		}
	}
	got, err := ArtifactName(Target{OS: " WINDOWS ", Arch: "AMD64", Archive: ".ZIP"})
	if err != nil || got != want[0] {
		t.Fatalf("ArtifactName() normalized = %q, %v", got, err)
	}
	if _, err := ArtifactName(Target{OS: "plan9", Arch: "amd64", Archive: "zip"}); !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("ArtifactName() unsupported error = %v", err)
	}
}

func TestManifestEntries(t *testing.T) {
	entries := ManifestEntries()
	if len(entries) != len(DefaultTargets()) {
		t.Fatalf("ManifestEntries() len = %d", len(entries))
	}
	for _, entry := range entries {
		if entry.SHA256 != "" || unsafeArtifactPath(entry.Path) {
			t.Fatalf("ManifestEntries() entry = %#v", entry)
		}
	}
}

func TestOperationalFiles(t *testing.T) {
	files := OperationalFiles()
	if len(files) != 5 {
		t.Fatalf("OperationalFiles() len = %d", len(files))
	}
	for _, file := range files {
		if file.Path == "" || file.Purpose == "" || len(file.Markers) == 0 {
			t.Fatalf("OperationalFiles() incomplete file = %#v", file)
		}
	}
}

func TestValidateOperationalFiles(t *testing.T) {
	dir := t.TempDir()
	for _, file := range OperationalFiles() {
		path := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(strings.Join(file.Markers, "\n")), 0o600); err != nil {
			t.Fatalf("WriteFile() %s error = %v", file.Path, err)
		}
	}
	results := ValidateOperationalFiles(dir)
	if len(results) != len(OperationalFiles()) {
		t.Fatalf("ValidateOperationalFiles() len = %d", len(results))
	}
	for _, result := range results {
		if result.Status != StatusVerified || result.Detail == "" {
			t.Fatalf("ValidateOperationalFiles() result = %#v", result)
		}
	}
}

func TestValidateOperationalFilesBlocked(t *testing.T) {
	dir := t.TempDir()
	results := ValidateOperationalFiles(dir)
	if len(results) == 0 || results[0].Status != StatusBlocked {
		t.Fatalf("ValidateOperationalFiles(missing) = %#v", results)
	}

	file := OperationalFiles()[0]
	path := filepath.Join(dir, filepath.FromSlash(file.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("WriteFile(empty) error = %v", err)
	}
	if got := validateOperationalFile(dir, file); got.Status != StatusBlocked || !strings.Contains(got.Detail, "empty") {
		t.Fatalf("validateOperationalFile(empty) = %#v", got)
	}
	if err := os.WriteFile(path, []byte(file.Markers[0]), 0o600); err != nil {
		t.Fatalf("WriteFile(partial) error = %v", err)
	}
	if got := validateOperationalFile(dir, file); got.Status != StatusBlocked || !strings.Contains(got.Detail, "missing required marker") {
		t.Fatalf("validateOperationalFile(partial) = %#v", got)
	}
	unsafe := OperationalFile{Path: "../bad", Purpose: "bad", Markers: []string{"x"}}
	if got := validateOperationalFile(dir, unsafe); got.Status != StatusBlocked || !strings.Contains(got.Detail, "relative") {
		t.Fatalf("validateOperationalFile(unsafe) = %#v", got)
	}
}

func TestGenerateManifestAndManifestText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.zip"), []byte("bravo"), 0o600); err != nil {
		t.Fatalf("WriteFile() b.zip error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("Mkdir() nested error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "a.tar.gz"), []byte("alpha"), 0o600); err != nil {
		t.Fatalf("WriteFile() nested artifact error = %v", err)
	}

	entries, err := GenerateManifest(dir, []string{"b.zip", "nested/a.tar.gz"})
	if err != nil {
		t.Fatalf("GenerateManifest() error = %v", err)
	}
	if len(entries) != 2 || entries[0].Path != "b.zip" || entries[1].Path != "nested/a.tar.gz" {
		t.Fatalf("GenerateManifest() entries = %#v", entries)
	}
	text, err := ManifestText([]Entry{entries[1], entries[0]})
	if err != nil {
		t.Fatalf("ManifestText() error = %v", err)
	}
	wantFirst := digest("bravo") + "  b.zip"
	wantSecond := digest("alpha") + "  nested/a.tar.gz"
	if text != wantFirst+"\n"+wantSecond+"\n" {
		t.Fatalf("ManifestText() = %q", text)
	}
	roundTrip, err := ParseManifest(text)
	if err != nil {
		t.Fatalf("ParseManifest(ManifestText()) error = %v", err)
	}
	if len(roundTrip) != 2 || roundTrip[0] != entries[0] || roundTrip[1] != entries[1] {
		t.Fatalf("manifest round trip = %#v", roundTrip)
	}
}

func TestGenerateManifestErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := GenerateManifest(dir, nil); !errors.Is(err, ErrEmptyManifest) {
		t.Fatalf("GenerateManifest(empty) error = %v", err)
	}
	if _, err := GenerateManifest(dir, []string{"../escape.zip"}); err == nil {
		t.Fatal("GenerateManifest() accepted unsafe path")
	}
	if _, err := GenerateManifest(dir, []string{"missing.zip"}); err == nil {
		t.Fatal("GenerateManifest() accepted missing artifact")
	}
}

func TestManifestTextErrors(t *testing.T) {
	if _, err := ManifestText(nil); !errors.Is(err, ErrEmptyManifest) {
		t.Fatalf("ManifestText(empty) error = %v", err)
	}
	if _, err := ManifestText([]Entry{{Path: "../escape.zip", SHA256: digest("x")}}); err == nil {
		t.Fatal("ManifestText() accepted unsafe path")
	}
	if _, err := ManifestText([]Entry{{Path: "bad.zip", SHA256: "abc"}}); err == nil {
		t.Fatal("ManifestText() accepted short digest")
	}
	if _, err := ManifestText([]Entry{{Path: "bad.zip", SHA256: strings.Repeat("z", 64)}}); err == nil {
		t.Fatal("ManifestText() accepted non-hex digest")
	}
}

func TestWriteManifest(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "artifact.zip")
	if err := os.WriteFile(artifact, []byte("artifact"), 0o600); err != nil {
		t.Fatalf("WriteFile() artifact error = %v", err)
	}
	path, err := WriteManifest(dir, []string{"artifact.zip"}, "")
	if err != nil {
		t.Fatalf("WriteManifest(default) error = %v", err)
	}
	if filepath.Base(path) != ManifestName {
		t.Fatalf("WriteManifest(default) path = %q", path)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() manifest error = %v", err)
	}
	if string(body) != digest("artifact")+"  artifact.zip\n" {
		t.Fatalf("default manifest body = %q", body)
	}

	explicit := filepath.Join(dir, "explicit.SHA256SUMS")
	path, err = WriteManifest(dir, []string{"artifact.zip"}, explicit)
	if err != nil {
		t.Fatalf("WriteManifest(explicit) error = %v", err)
	}
	if path != explicit {
		t.Fatalf("WriteManifest(explicit) path = %q", path)
	}
}

func TestWriteManifestErrors(t *testing.T) {
	if _, err := WriteManifest(t.TempDir(), nil, ""); !errors.Is(err, ErrEmptyManifest) {
		t.Fatalf("WriteManifest(empty) error = %v", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "artifact.zip"), []byte("artifact"), 0o600); err != nil {
		t.Fatalf("WriteFile() artifact error = %v", err)
	}
	origWrite := writeFile
	defer func() { writeFile = origWrite }()
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write failed") }
	if _, err := WriteManifest(dir, []string{"artifact.zip"}, filepath.Join(dir, "SHA256SUMS")); err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("WriteManifest() write error = %v", err)
	}
}

func TestBinaryFileName(t *testing.T) {
	win, err := BinaryFileName(Target{OS: "windows", Arch: "amd64", Archive: "zip"})
	if err != nil || win != "recomphamr.exe" {
		t.Fatalf("BinaryFileName(windows) = %q, %v", win, err)
	}
	linux, err := BinaryFileName(Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"})
	if err != nil || linux != "recomphamr" {
		t.Fatalf("BinaryFileName(linux) = %q, %v", linux, err)
	}
	if _, err := BinaryFileName(Target{OS: "linux", Arch: "386", Archive: "tar.gz"}); !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("BinaryFileName(unsupported) error = %v", err)
	}
}

func TestBinaryOutputName(t *testing.T) {
	win, err := BinaryOutputName(Target{OS: "windows", Arch: "amd64", Archive: "zip"})
	if err != nil || win != "recomphamr_windows_amd64.exe" {
		t.Fatalf("BinaryOutputName(windows) = %q, %v", win, err)
	}
	linux, err := BinaryOutputName(Target{OS: "linux", Arch: "arm64", Archive: "tar.gz"})
	if err != nil || linux != "recomphamr_linux_arm64" {
		t.Fatalf("BinaryOutputName(linux) = %q, %v", linux, err)
	}
	if _, err := BinaryOutputName(Target{OS: "plan9", Arch: "amd64", Archive: "zip"}); !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("BinaryOutputName(unsupported) error = %v", err)
	}
}

func TestBuildBinary(t *testing.T) {
	origRun := runCmd
	defer func() { runCmd = origRun }()
	var gotName string
	var gotArgs []string
	var gotEnv []string
	runCmd = func(name string, args []string, env []string) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		gotEnv = append([]string(nil), env...)
		return nil
	}
	out, err := BuildBinary(BuildSpec{Target: Target{OS: "windows", Arch: "amd64", Archive: "zip"}, OutputDir: t.TempDir()})
	if err != nil {
		t.Fatalf("BuildBinary() error = %v", err)
	}
	if gotName != "go" || filepath.Base(out) != "recomphamr_windows_amd64.exe" {
		t.Fatalf("BuildBinary() command=%q output=%q", gotName, out)
	}
	if strings.Join(gotArgs, " ") != "build -trimpath -o "+out+" ./cmd/recomphamr" {
		t.Fatalf("BuildBinary() args = %#v", gotArgs)
	}
	for _, want := range []string{"GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0"} {
		if !contains(gotEnv, want) {
			t.Fatalf("BuildBinary() env missing %q: %#v", want, gotEnv)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defaultDir := t.TempDir()
	if err := os.Chdir(defaultDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer os.Chdir(cwd)
	out, err = BuildBinary(BuildSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}})
	if err != nil {
		t.Fatalf("BuildBinary() default output error = %v", err)
	}
	if filepath.ToSlash(out) != "dist/recomphamr_linux_amd64" {
		t.Fatalf("BuildBinary() default output = %q", out)
	}
}

func TestBuildBinaryErrors(t *testing.T) {
	if _, err := BuildBinary(BuildSpec{Target: Target{OS: "linux", Arch: "386", Archive: "tar.gz"}, OutputDir: t.TempDir()}); !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("BuildBinary() unsupported error = %v", err)
	}
	dir := t.TempDir()
	existing := filepath.Join(dir, "recomphamr_linux_amd64")
	if err := os.WriteFile(existing, []byte("x"), 0o755); err != nil {
		t.Fatalf("WriteFile() existing binary error = %v", err)
	}
	if _, err := BuildBinary(BuildSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, OutputDir: dir}); !errors.Is(err, ErrBinaryExists) {
		t.Fatalf("BuildBinary() existing error = %v", err)
	}
	outputFile := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(outputFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() output file error = %v", err)
	}
	if _, err := BuildBinary(BuildSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, OutputDir: outputFile}); err == nil {
		t.Fatal("BuildBinary() accepted file output dir")
	}

	origStat, origRun := statFile, runCmd
	defer func() { statFile, runCmd = origStat, origRun }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("stat failed") }
	if _, err := BuildBinary(BuildSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, OutputDir: filepath.Join(t.TempDir(), "stat")}); err == nil || !strings.Contains(err.Error(), "stat failed") {
		t.Fatalf("BuildBinary() stat error = %v", err)
	}
	statFile = origStat
	runCmd = func(string, []string, []string) error { return errors.New("build failed") }
	if _, err := BuildBinary(BuildSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, Package: "./cmd/recomphamr", OutputDir: filepath.Join(t.TempDir(), "run")}); err == nil || !strings.Contains(err.Error(), "build failed") {
		t.Fatalf("BuildBinary() run error = %v", err)
	}
}

func TestRunBuildCommand(t *testing.T) {
	if err := runBuildCommand("go", []string{"version"}, nil); err != nil {
		t.Fatalf("runBuildCommand(go version) error = %v", err)
	}
	if err := runBuildCommand("go", []string{"definitely-not-a-go-command"}, nil); err == nil {
		t.Fatal("runBuildCommand() accepted failing command")
	}
	if err := runBuildCommand("definitely-not-a-real-command", nil, nil); err == nil {
		t.Fatal("runBuildCommand() accepted missing executable")
	}
}

func TestArchiveBytes(t *testing.T) {
	zipData, err := ArchiveBytes(Target{OS: "windows", Arch: "amd64", Archive: "zip"}, "recomphamr.exe", []byte("exe"))
	if err != nil {
		t.Fatalf("ArchiveBytes(zip) error = %v", err)
	}
	if got := readZipMember(t, zipData, "recomphamr.exe"); got != "exe" {
		t.Fatalf("zip member = %q", got)
	}
	tarData, err := ArchiveBytes(Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, "recomphamr", []byte("bin"))
	if err != nil {
		t.Fatalf("ArchiveBytes(tar.gz) error = %v", err)
	}
	if got := readTarGzMember(t, tarData, "recomphamr"); got != "bin" {
		t.Fatalf("tar.gz member = %q", got)
	}
	if _, err := ArchiveBytes(Target{OS: "linux", Arch: "386", Archive: "tar.gz"}, "recomphamr", []byte("bin")); !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("ArchiveBytes(unsupported) error = %v", err)
	}
	if _, err := ArchiveBytes(Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, "../recomphamr", []byte("bin")); err == nil {
		t.Fatal("ArchiveBytes() accepted unsafe binary name")
	}
}

func TestCreateArchive(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "recomphamr.exe")
	if err := os.WriteFile(binary, []byte("exe"), 0o755); err != nil {
		t.Fatalf("WriteFile() binary error = %v", err)
	}
	outDir := filepath.Join(dir, "dist")
	path, err := CreateArchive(ArchiveSpec{Target: Target{OS: "windows", Arch: "amd64", Archive: "zip"}, BinaryPath: binary, OutputDir: outDir})
	if err != nil {
		t.Fatalf("CreateArchive() error = %v", err)
	}
	if filepath.Base(path) != "recomphamr_windows_amd64.zip" {
		t.Fatalf("CreateArchive() path = %q", path)
	}
	archiveData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() archive error = %v", err)
	}
	if got := readZipMember(t, archiveData, "recomphamr.exe"); got != "exe" {
		t.Fatalf("created archive member = %q", got)
	}
	if _, err := CreateArchive(ArchiveSpec{Target: Target{OS: "windows", Arch: "amd64", Archive: "zip"}, BinaryPath: binary, OutputDir: outDir}); !errors.Is(err, ErrArchiveExists) {
		t.Fatalf("CreateArchive() overwrite error = %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defaultDir := t.TempDir()
	if err := os.Chdir(defaultDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer os.Chdir(cwd)
	defaultPath, err := CreateArchive(ArchiveSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, BinaryPath: binary})
	if err != nil {
		t.Fatalf("CreateArchive() default output error = %v", err)
	}
	if filepath.Dir(defaultPath) != "." || filepath.Base(defaultPath) != "recomphamr_linux_amd64.tar.gz" {
		t.Fatalf("CreateArchive() default path = %q", defaultPath)
	}
}

func TestCreateArchiveErrors(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "recomphamr")
	if err := os.WriteFile(binary, []byte("bin"), 0o755); err != nil {
		t.Fatalf("WriteFile() binary error = %v", err)
	}
	if _, err := CreateArchive(ArchiveSpec{Target: Target{OS: "linux", Arch: "386", Archive: "tar.gz"}, BinaryPath: binary, OutputDir: dir}); !errors.Is(err, ErrUnsupportedTarget) {
		t.Fatalf("CreateArchive() unsupported target error = %v", err)
	}
	if _, err := CreateArchive(ArchiveSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, BinaryPath: filepath.Join(dir, "missing"), OutputDir: dir}); err == nil {
		t.Fatal("CreateArchive() accepted missing binary")
	}
	outputFile := filepath.Join(dir, "file")
	if err := os.WriteFile(outputFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() output blocker error = %v", err)
	}
	if _, err := CreateArchive(ArchiveSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, BinaryPath: binary, OutputDir: outputFile}); err == nil {
		t.Fatal("CreateArchive() accepted file output dir")
	}

	origStat, origWrite := statFile, writeFile
	defer func() { statFile, writeFile = origStat, origWrite }()
	statFile = func(string) (os.FileInfo, error) { return nil, errors.New("stat failed") }
	if _, err := CreateArchive(ArchiveSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, BinaryPath: binary, OutputDir: dir}); err == nil || !strings.Contains(err.Error(), "stat failed") {
		t.Fatalf("CreateArchive() stat error = %v", err)
	}
	statFile = origStat
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write failed") }
	writeDir := filepath.Join(dir, "write-error")
	if _, err := CreateArchive(ArchiveSpec{Target: Target{OS: "linux", Arch: "amd64", Archive: "tar.gz"}, BinaryPath: binary, OutputDir: writeDir}); err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("CreateArchive() write error = %v", err)
	}
}

func TestParseManifestErrors(t *testing.T) {
	cases := map[string]string{
		"empty":     "# only comments\n",
		"fields":    "abc\n",
		"length":    "abc file.zip\n",
		"hex":       strings.Repeat("z", 64) + " file.zip\n",
		"emptypath": strings.Repeat("a", 64) + " *\n",
		"scanner":   strings.Repeat("a", 70000),
	}
	for name, text := range cases {
		_, err := ParseManifest(text)
		if name == "empty" {
			if !errors.Is(err, ErrEmptyManifest) {
				t.Fatalf("%s ParseManifest() error = %v, want ErrEmptyManifest", name, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("%s ParseManifest() accepted invalid manifest", name)
		}
	}
}

func TestVerifyManifestSuccess(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "recomphamr_windows_amd64.zip")
	body := []byte("binary")
	if err := os.WriteFile(artifact, body, 0o600); err != nil {
		t.Fatalf("WriteFile() artifact error = %v", err)
	}
	manifest := filepath.Join(dir, "SHA256SUMS")
	if err := os.WriteFile(manifest, []byte(digest(string(body))+" recomphamr_windows_amd64.zip\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() manifest error = %v", err)
	}
	report, err := VerifyManifest(dir, manifest)
	if err != nil {
		t.Fatalf("VerifyManifest() error = %v", err)
	}
	if !report.Verified() || report.Count(StatusVerified) != 1 || report.Count(StatusBlocked) != 0 {
		t.Fatalf("report verification = %#v", report)
	}
	out := report.String()
	if !strings.Contains(out, "[verified] recomphamr_windows_amd64.zip: sha256 matched") || strings.HasSuffix(out, "\n") {
		t.Fatalf("report string = %q", out)
	}
}

func TestVerifyManifestBlockedResults(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.zip"), []byte("actual"), 0o600); err != nil {
		t.Fatalf("WriteFile() bad artifact error = %v", err)
	}
	manifest := filepath.Join(dir, "SHA256SUMS")
	text := strings.Join([]string{
		digest("expected") + " bad.zip",
		digest("missing") + " missing.zip",
		digest("escape") + " ../escape.zip",
	}, "\n")
	if err := os.WriteFile(manifest, []byte(text), 0o600); err != nil {
		t.Fatalf("WriteFile() manifest error = %v", err)
	}
	report, err := VerifyManifest(dir, manifest)
	if err != nil {
		t.Fatalf("VerifyManifest() error = %v", err)
	}
	if report.Verified() || report.Count(StatusBlocked) != 3 {
		t.Fatalf("blocked report = %#v", report)
	}
	out := report.String()
	for _, want := range []string{"sha256 mismatch", "missing.zip", "artifact path must be relative"} {
		if !strings.Contains(out, want) {
			t.Fatalf("report missing %q:\n%s", want, out)
		}
	}
}

func TestVerifyManifestReadAndParseErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := VerifyManifest(dir, filepath.Join(dir, "missing.SHA256SUMS")); err == nil {
		t.Fatal("VerifyManifest() accepted missing manifest")
	}
	manifest := filepath.Join(dir, "SHA256SUMS")
	if err := os.WriteFile(manifest, []byte("bad\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() manifest error = %v", err)
	}
	if _, err := VerifyManifest(dir, manifest); err == nil {
		t.Fatal("VerifyManifest() accepted malformed manifest")
	}
}

func TestUnsafeArtifactPath(t *testing.T) {
	cases := map[string]bool{
		"artifact.zip":        false,
		"nested/artifact.zip": false,
		"":                    true,
		".":                   true,
		"..":                  true,
		"../artifact.zip":     true,
		"/artifact.zip":       true,
		`\\artifact.zip`:      true,
	}
	for path, want := range cases {
		if got := unsafeArtifactPath(path); got != want {
			t.Fatalf("unsafeArtifactPath(%q) = %v, want %v", path, got, want)
		}
	}
}

func digest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func readZipMember(t *testing.T, data []byte, name string) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader(zip) error = %v", err)
	}
	for _, file := range r.File {
		if file.Name != name {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("Open(zip member) error = %v", err)
		}
		defer rc.Close()
		body, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("ReadAll(zip member) error = %v", err)
		}
		return string(body)
	}
	t.Fatalf("zip member %q missing", name)
	return ""
}

func readTarGzMember(t *testing.T, data []byte, name string) string {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("NewReader(gzip) error = %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next(tar) error = %v", err)
		}
		if header.Name != name {
			continue
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("ReadAll(tar member) error = %v", err)
		}
		return string(body)
	}
	t.Fatalf("tar member %q missing", name)
	return ""
}

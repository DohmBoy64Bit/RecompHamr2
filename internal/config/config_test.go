package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBootstrapCreatesSecureDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	cfg, created, err := Bootstrap(dir)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if !created {
		t.Fatal("Bootstrap() created = false, want true")
	}
	if cfg.Active != "lmstudio-amd" {
		t.Fatalf("active = %q", cfg.Active)
	}
	profile, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile() error = %v", err)
	}
	if profile.LLM != "qwen/qwen3.6-35b-a3b" || profile.ContextSize != DefaultContextSize {
		t.Fatalf("default profile = %+v", profile)
	}
	if _, err := os.Stat(filepath.Join(dir, DirName, FileName)); err != nil {
		t.Fatalf("config not written: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, DirName, FileName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, want := range []string{"recomphamr configuration", "host.docker.internal", "context_size"} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("config header missing %q:\n%s", want, raw)
		}
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	data := []byte("active: local\nmodels:\n  local:\n    llm: m\n    url: http://x\n    key: ''\n    context_size: 1\nextra: nope\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() accepted unknown field")
	}
}

func TestSetActivePersists(t *testing.T) {
	cfg, _, err := Bootstrap(t.TempDir())
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if err := cfg.SetActive("ollama-amd"); err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}
	loaded, err := Load(filepath.Join(cfg.Dir, FileName))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Active != "ollama-amd" {
		t.Fatalf("active = %q, want ollama-amd", loaded.Active)
	}
	if err := cfg.SetActive("missing"); err == nil {
		t.Fatal("SetActive() accepted unknown profile")
	}
}

func TestURLOverrideDoesNotPersist(t *testing.T) {
	t.Setenv(URLOverrideEnv, "http://override")
	cfg, _, err := Bootstrap(t.TempDir())
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	profile, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile() error = %v", err)
	}
	if profile.URL != "http://override" {
		t.Fatalf("URL = %q, want override", profile.URL)
	}
	t.Setenv(URLOverrideEnv, "")
	loaded, err := Load(filepath.Join(cfg.Dir, FileName))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	profile, err = loaded.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile() reload error = %v", err)
	}
	if profile.URL == "http://override" {
		t.Fatal("override persisted to disk")
	}
}

func TestBootstrapRefusesUnsafePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, DirName), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, _, err := Bootstrap(root); err == nil {
		t.Fatal("Bootstrap() accepted regular file .rehamr")
	}
}

func TestSaveRequiresDirectory(t *testing.T) {
	if err := Default().Save(); err == nil {
		t.Fatal("Save() without Dir succeeded")
	}
}

func TestBootstrapLoadsExistingAndCoercesContext(t *testing.T) {
	dir := t.TempDir()
	cdir := filepath.Join(dir, DirName)
	if err := os.Mkdir(cdir, 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	data := []byte("active: local\nmodels:\n  local:\n    llm: m\n    url: http://x\n    key: ''\n    context_size: 0\n")
	if err := os.WriteFile(filepath.Join(cdir, FileName), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, created, err := Bootstrap(dir)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if created {
		t.Fatal("Bootstrap() created existing config")
	}
	profile, err := cfg.ActiveProfile()
	if err != nil {
		t.Fatalf("ActiveProfile() error = %v", err)
	}
	if profile.ContextSize != DefaultContextSize {
		t.Fatalf("ContextSize = %d, want default", profile.ContextSize)
	}
}

func TestLoadValidationFailures(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(filepath.Join(dir, "missing.yaml")); err == nil {
		t.Fatal("Load() accepted missing file")
	}
	empty := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(empty, []byte("active: local\nmodels: {}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Load(empty); err == nil {
		t.Fatal("Load() accepted no profiles")
	}
	noModels := filepath.Join(dir, "no-models.yaml")
	if err := os.WriteFile(noModels, []byte("active: missing\nmodels: {}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Load(noModels); err == nil {
		t.Fatal("Load() accepted empty models")
	}
}

func TestLoadCoercesMissingActiveToFirstModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	data := []byte("active: missing\nmodels:\n  zulu:\n    llm: m\n    url: http://z\n    key: ''\n    context_size: 1\n  alpha:\n    llm: m\n    url: http://a\n    key: ''\n    context_size: 1\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Active != "alpha" {
		t.Fatalf("Active = %q, want alpha", cfg.Active)
	}
}

func TestActiveProfileMissing(t *testing.T) {
	cfg := Default()
	cfg.Active = "missing"
	if _, err := cfg.ActiveProfile(); err == nil {
		t.Fatal("ActiveProfile() accepted missing active profile")
	}
}

func TestSaveFailureSeams(t *testing.T) {
	cfg := Default()
	cfg.Dir = t.TempDir()
	origWrite, origChmod, origRename, origRemove, origMarshal := writeFile, chmodFile, renameFile, removeFile, marshalYAML
	defer func() {
		writeFile, chmodFile, renameFile, removeFile = origWrite, origChmod, origRename, origRemove
		marshalYAML = origMarshal
	}()

	marshalYAML = func(any) ([]byte, error) { return nil, errors.New("marshal") }
	if err := cfg.Save(); err == nil {
		t.Fatal("Save() accepted marshal failure")
	}
	marshalYAML = origMarshal
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write") }
	if err := cfg.Save(); err == nil {
		t.Fatal("Save() accepted write failure")
	}
	writeFile = origWrite
	chmodFile = func(string, os.FileMode) error { return errors.New("chmod") }
	removeFile = func(string) error { return nil }
	if err := cfg.Save(); err == nil {
		t.Fatal("Save() accepted chmod failure")
	}
	chmodFile = origChmod
	renameFile = func(string, string) error { return errors.New("rename") }
	if err := cfg.Save(); err == nil {
		t.Fatal("Save() accepted rename failure")
	}
}

func TestSetActiveRevertsOnSaveFailure(t *testing.T) {
	cfg := Default()
	cfg.Dir = ""
	if err := cfg.SetActive("ollama-amd"); err == nil {
		t.Fatal("SetActive() accepted save failure")
	}
	if cfg.Active != "lmstudio-amd" {
		t.Fatalf("Active = %q, want rollback to lmstudio-amd", cfg.Active)
	}
}

func TestModelNamesSortedAndLoggingRoundTrips(t *testing.T) {
	cfg := Default()
	if got := cfg.ModelNames(); !reflect.DeepEqual(got, []string{"llama-vulkan", "lmstudio-amd", "lmstudio-fast", "ollama-amd"}) {
		t.Fatalf("ModelNames() = %#v", got)
	}
	cfg.Dir = t.TempDir()
	cfg.Logging = true
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := Load(filepath.Join(cfg.Dir, FileName))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !loaded.Logging {
		t.Fatal("Logging did not round-trip")
	}
}

func TestBootstrapFailureSeams(t *testing.T) {
	origLstat, origStat, origMkdir, origChmod, origWrite := lstatPath, statPath, mkdirAll, chmodFile, writeFile
	defer func() {
		lstatPath, statPath, mkdirAll, chmodFile, writeFile = origLstat, origStat, origMkdir, origChmod, origWrite
	}()
	lstatPath = func(string) (os.FileInfo, error) { return nil, errors.New("lstat") }
	if _, _, err := Bootstrap(t.TempDir()); err == nil {
		t.Fatal("Bootstrap() accepted lstat failure")
	}
	lstatPath = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	mkdirAll = func(string, os.FileMode) error { return errors.New("mkdir") }
	if _, _, err := Bootstrap(t.TempDir()); err == nil {
		t.Fatal("Bootstrap() accepted mkdir failure")
	}
	mkdirAll = origMkdir
	chmodFile = func(string, os.FileMode) error { return errors.New("chmod") }
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, DirName), 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	lstatPath = origLstat
	if _, _, err := Bootstrap(dir); err == nil {
		t.Fatal("Bootstrap() accepted chmod failure")
	}
	chmodFile = origChmod
	lstatPath = func(path string) (os.FileInfo, error) {
		if filepath.Base(path) == FileName {
			return fakeInfo{mode: os.ModeSymlink}, nil
		}
		return origLstat(path)
	}
	if _, _, err := Bootstrap(dir); err == nil {
		t.Fatal("Bootstrap() accepted config symlink")
	}
	statPath = func(string) (os.FileInfo, error) { return nil, errors.New("stat") }
	lstatPath = func(string) (os.FileInfo, error) { return fakeInfo{mode: os.ModeDir}, nil }
	if _, _, err := Bootstrap(t.TempDir()); err == nil {
		t.Fatal("Bootstrap() accepted load after stat failure")
	}
	statPath = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("write") }
	if _, _, err := Bootstrap(t.TempDir()); err == nil {
		t.Fatal("Bootstrap() accepted save failure")
	}
	writeFile = origWrite
}

func TestBootstrapCreateSaveAndLoadErrors(t *testing.T) {
	origLstat, origStat, origWrite, origChmod := lstatPath, statPath, writeFile, chmodFile
	defer func() { lstatPath, statPath, writeFile, chmodFile = origLstat, origStat, origWrite, origChmod }()

	lstatPath = func(string) (os.FileInfo, error) { return fakeInfo{mode: os.ModeDir}, nil }
	chmodFile = func(string, os.FileMode) error { return nil }
	statPath = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("save") }
	if _, _, err := Bootstrap(t.TempDir()); err == nil {
		t.Fatal("Bootstrap() accepted create Save failure")
	}

	writeFile = origWrite
	statPath = func(string) (os.FileInfo, error) { return fakeInfo{}, nil }
	if _, _, err := Bootstrap(t.TempDir()); err == nil {
		t.Fatal("Bootstrap() accepted load failure")
	}
}

type fakeInfo struct{ mode os.FileMode }

func (f fakeInfo) Name() string       { return "fake" }
func (f fakeInfo) Size() int64        { return 0 }
func (f fakeInfo) Mode() os.FileMode  { return f.mode }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return f.mode&os.ModeDir != 0 }
func (f fakeInfo) Sys() any           { return nil }

package config

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

var (
	writeFile   = os.WriteFile
	chmodFile   = os.Chmod
	renameFile  = os.Rename
	removeFile  = os.Remove
	lstatPath   = os.Lstat
	statPath    = os.Stat
	mkdirAll    = os.MkdirAll
	marshalYAML = yaml.Marshal
)

const (
	// DirName is the project-local RecompHamr state directory.
	DirName = ".rehamr"
	// FileName is the YAML config file inside DirName.
	FileName = "config.yaml"
	// URLOverrideEnv overrides the active profile URL for the current process only.
	URLOverrideEnv = "RECOMPHAMR_URL"
	// DefaultContextSize is used when a local profile omits or invalidates context_size.
	DefaultContextSize = 32768
)

// Profile is one OpenAI-compatible model endpoint.
type Profile struct {
	// LLM is the model identifier sent to the OpenAI-compatible backend.
	LLM string `yaml:"llm"`
	// URL is the backend base URL for this profile.
	URL string `yaml:"url"`
	// Key is an optional API key; local profiles usually leave it empty.
	Key string `yaml:"key"`
	// ContextSize is the prompt packing budget for this profile.
	ContextSize int `yaml:"context_size,omitempty"`
}

// Config is the strict YAML schema for .rehamr/config.yaml.
type Config struct {
	// Active names the selected model profile.
	Active string `yaml:"active"`
	// Models maps profile names to OpenAI-compatible endpoint settings.
	Models map[string]Profile `yaml:"models"`
	// Logging enables redacted diagnostic session logging when app wiring supports logs.
	Logging bool `yaml:"logging,omitempty"`
	// Dir is the runtime-only directory containing config.yaml.
	Dir string `yaml:"-"`
}

// Default returns the seeded local-first profile set.
func Default() *Config {
	return &Config{
		Active: "lmstudio-amd",
		Models: map[string]Profile{
			"lmstudio-amd":  {LLM: "qwen/qwen3.6-35b-a3b", URL: "http://localhost:1234", ContextSize: DefaultContextSize},
			"lmstudio-fast": {LLM: "openai/gpt-oss-20b", URL: "http://localhost:1234", ContextSize: DefaultContextSize},
			"ollama-amd":    {LLM: "qwen3.6:27b", URL: "http://localhost:11434", ContextSize: DefaultContextSize},
			"llama-vulkan":  {LLM: "qwen3.6-35b-a3b", URL: "http://localhost:8080", ContextSize: DefaultContextSize},
		},
	}
}

// Bootstrap loads or creates .rehamr/config.yaml under projectDir.
func Bootstrap(projectDir string) (*Config, bool, error) {
	dir := filepath.Join(projectDir, DirName)
	if info, err := lstatPath(dir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, false, fmt.Errorf("%s must be a real directory", DirName)
		}
		if err := chmodFile(dir, 0o700); err != nil {
			return nil, false, err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := mkdirAll(dir, 0o700); err != nil {
			return nil, false, err
		}
	} else {
		return nil, false, err
	}

	path := filepath.Join(dir, FileName)
	if info, err := lstatPath(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, false, fmt.Errorf("%s must not be a symlink", FileName)
	}
	if _, err := statPath(path); errors.Is(err, os.ErrNotExist) {
		cfg := Default()
		cfg.Dir = dir
		if err := cfg.Save(); err != nil {
			return nil, false, err
		}
		applyURLOverride(cfg)
		return cfg, true, nil
	}
	cfg, err := Load(path)
	if err != nil {
		return nil, false, err
	}
	cfg.Dir = dir
	coerce(cfg)
	applyURLOverride(cfg)
	return cfg, false, nil
}

// Load strictly decodes a config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config.yaml: %w", err)
	}
	if cfg.Models == nil || len(cfg.Models) == 0 {
		return nil, errors.New("config.yaml: no profiles configured")
	}
	if _, ok := cfg.Models[cfg.Active]; !ok {
		cfg.Active = cfg.ModelNames()[0]
	}
	cfg.Dir = filepath.Dir(path)
	coerce(&cfg)
	applyURLOverride(&cfg)
	return &cfg, nil
}

// Save atomically rewrites config.yaml with owner-only permissions.
func (c *Config) Save() error {
	if c.Dir == "" {
		return errors.New("config directory is empty")
	}
	coerce(c)
	data, err := marshalYAML(c)
	if err != nil {
		return err
	}
	data = append(configHeader(), data...)
	path := filepath.Join(c.Dir, FileName)
	tmpName := path + ".tmp"
	if err := writeFile(tmpName, data, 0o600); err != nil {
		return err
	}
	if err := chmodFile(tmpName, 0o600); err != nil {
		removeFile(tmpName)
		return err
	}
	if err := renameFile(tmpName, path); err != nil {
		removeFile(tmpName)
		return err
	}
	return nil
}

// ActiveProfile returns the selected profile.
func (c *Config) ActiveProfile() (Profile, error) {
	profile, ok := c.Models[c.Active]
	if !ok {
		return Profile{}, fmt.Errorf("active profile %q not found", c.Active)
	}
	return profile, nil
}

// ModelNames returns profile names in deterministic sorted order.
func (c *Config) ModelNames() []string {
	return slices.Sorted(maps.Keys(c.Models))
}

// SetActive changes the active profile and persists the config.
func (c *Config) SetActive(name string) error {
	if _, ok := c.Models[name]; !ok {
		return fmt.Errorf("unknown profile %q", name)
	}
	prev := c.Active
	c.Active = name
	if err := c.Save(); err != nil {
		c.Active = prev
		return err
	}
	return nil
}

func coerce(c *Config) {
	for name, profile := range c.Models {
		if profile.ContextSize <= 0 {
			profile.ContextSize = DefaultContextSize
			c.Models[name] = profile
		}
	}
}

func configHeader() []byte {
	return []byte(`# recomphamr configuration
#
# For devcontainer or WSL2 access to a host Ollama server, replace
# http://localhost:11434 with http://host.docker.internal:11434.
# context_size is the packing budget RecompHamr uses for this profile.

`)
}

func applyURLOverride(c *Config) {
	override := os.Getenv(URLOverrideEnv)
	if override == "" {
		return
	}
	profile := c.Models[c.Active]
	profile.URL = override
	c.Models[c.Active] = profile
}

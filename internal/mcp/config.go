package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

var readConfigFile = os.ReadFile

// ConfigFile is the persistent .rehamr/mcp.json schema.
type ConfigFile struct {
	// Servers maps server names to persistent MCP overrides or custom entries.
	Servers map[string]PersistentServerConfig `json:"servers"`
}

// PersistentServerConfig is one .rehamr/mcp.json server entry.
type PersistentServerConfig struct {
	// Name is the server name and must match the map key when present.
	Name string `json:"name"`
	// Command is the stdio command used to launch the server.
	Command string `json:"command"`
	// Args are stdio command arguments.
	Args []string `json:"args"`
	// URL is the streamable HTTP endpoint for the server.
	URL string `json:"url"`
	// AllowedTools is the configured tool allowlist.
	AllowedTools []string `json:"allowed_tools"`
	// Autostart controls runtime autoconnect for this server.
	Autostart *bool `json:"autostart"`
	// RequireSkill controls whether tools are gated by a matching skill.
	RequireSkill *bool `json:"require_skill"`
}

// LoadConfigFile strictly loads a persistent MCP config file.
func LoadConfigFile(path string) (ConfigFile, error) {
	data, err := readConfigFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ConfigFile{Servers: map[string]PersistentServerConfig{}}, nil
		}
		return ConfigFile{}, err
	}
	var cfg ConfigFile
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return ConfigFile{}, fmt.Errorf("mcp.json: %w", err)
	}
	if cfg.Servers == nil {
		cfg.Servers = map[string]PersistentServerConfig{}
	}
	for name, server := range cfg.Servers {
		if strings.TrimSpace(server.Name) == "" {
			server.Name = strings.TrimSpace(name)
		}
		if server.Name != strings.TrimSpace(name) {
			return ConfigFile{}, fmt.Errorf("mcp.json: server key %q does not match name %q", name, server.Name)
		}
		cfg.Servers[name] = server
	}
	return cfg, nil
}

// MergeConfigs overlays persistent server config onto built-in configs.
func MergeConfigs(builtins []ServerConfig, cfg ConfigFile) []ServerConfig {
	out := append([]ServerConfig(nil), builtins...)
	index := map[string]int{}
	for i, server := range out {
		index[server.Name] = i
	}
	names := mapsKeys(cfg.Servers)
	sort.Strings(names)
	for _, name := range names {
		override := cfg.Servers[name]
		if i, ok := index[name]; ok {
			out[i] = mergeServer(out[i], override)
			continue
		}
		out = append(out, persistentToServer(override))
	}
	return out
}

func persistentToServer(config PersistentServerConfig) ServerConfig {
	server := ServerConfig{
		Name:         config.Name,
		Command:      config.Command,
		Args:         append([]string(nil), config.Args...),
		URL:          config.URL,
		AllowedTools: append([]string(nil), config.AllowedTools...),
	}
	if config.Autostart != nil {
		server.Autostart = *config.Autostart
	}
	if config.RequireSkill != nil {
		server.RequireSkill = *config.RequireSkill
	}
	return server
}

func mergeServer(base ServerConfig, override PersistentServerConfig) ServerConfig {
	if strings.TrimSpace(override.Name) != "" {
		base.Name = override.Name
	}
	if override.Command != "" {
		base.Command = override.Command
	}
	if override.Args != nil {
		base.Args = append([]string(nil), override.Args...)
	}
	if override.URL != "" {
		base.URL = override.URL
	}
	if override.AllowedTools != nil {
		base.AllowedTools = append([]string(nil), override.AllowedTools...)
	}
	if override.Autostart != nil {
		base.Autostart = *override.Autostart
	}
	if override.RequireSkill != nil {
		base.RequireSkill = *override.RequireSkill
	}
	return base
}

func mapsKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

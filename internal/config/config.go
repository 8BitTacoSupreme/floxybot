package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration loaded from config file and env.
type Config struct {
	AnthropicAPIKey string `toml:"anthropic_api_key"`
	VoyageAPIKey    string `toml:"voyage_api_key"`
	Model           string `toml:"model"`
	CanonDir        string `toml:"canon_dir"`
	BackendURL      string `toml:"backend_url"`
	Verbose         bool   `toml:"-"`
}

// DefaultConfig returns config with sensible defaults.
func DefaultConfig() *Config {
	dataDir := xdgDataDir()
	return &Config{
		Model:      "claude-sonnet-4-20250514",
		CanonDir:   filepath.Join(dataDir, "floxybot", "canon"),
		BackendURL: "https://floxybot.example.com",
	}
}

// Load reads config from XDG config file, then applies env overrides.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	cfgPath := configFilePath()
	if _, err := os.Stat(cfgPath); err == nil {
		if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
			return nil, err
		}
	}

	// Environment overrides always win.
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := os.Getenv("VOYAGE_API_KEY"); v != "" {
		cfg.VoyageAPIKey = v
	}
	if v := os.Getenv("FLOXYBOT_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("FLOXYBOT_CANON_DIR"); v != "" {
		cfg.CanonDir = v
	}
	if v := os.Getenv("FLOXYBOT_BACKEND_URL"); v != "" {
		cfg.BackendURL = v
	}

	return cfg, nil
}

func configFilePath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "floxybot", "config.toml")
}

func xdgDataDir() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".local", "share")
	}
	return dir
}

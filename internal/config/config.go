package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds persistent application settings.
type Config struct {
	BaseDir string `json:"base_dir"`
	Editor  string `json:"editor,omitempty"`
}

// configPath returns the path to the config file.
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "treework", "config.json")
}

// FileExists returns true if the config file exists on disk.
func FileExists() bool {
	_, err := os.Stat(configPath())
	return err == nil
}

// Load reads the config file from disk. Returns defaults if the file is missing or unreadable.
func Load() *Config {
	cfg := &Config{}
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, cfg)
	return cfg
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// DevDir returns the root development directory.
// Returns empty string if not configured — caller must handle setup.
func DevDir() string {
	if d := os.Getenv("DEV_DIR"); d != "" {
		return d
	}
	if cfg := Load(); cfg.BaseDir != "" {
		return cfg.BaseDir
	}
	return ""
}

// Editor returns the preferred editor command.
// Priority: WT_EDITOR env var > config file > "" (auto-detect).
func Editor() string {
	if e := os.Getenv("WT_EDITOR"); e != "" {
		return e
	}
	if cfg := Load(); cfg.Editor != "" {
		return cfg.Editor
	}
	return ""
}

// EditorSource returns the current editor and a human-readable source description.
func EditorSource() (editor, source string) {
	if e := os.Getenv("WT_EDITOR"); e != "" {
		return e, "WT_EDITOR env var"
	}
	if cfg := Load(); cfg.Editor != "" {
		return cfg.Editor, "config file"
	}
	return "", "auto-detect"
}

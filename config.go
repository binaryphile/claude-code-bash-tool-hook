package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the hook configuration
type Config struct {
	Enabled                 bool     `json:"enabled"`
	AdditionalSafePatterns  []string `json:"additional_safe_patterns,omitempty"`
	AdditionalEscapeMarkers []string `json:"additional_escape_markers,omitempty"`
	ForceWrapPatterns       []string `json:"force_wrap_patterns,omitempty"`
	DebugLog                bool     `json:"debug_log,omitempty"`
	LogFile                 string   `json:"log_file,omitempty"`
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "bash-hook-config.json")
}

// LoadConfig loads the configuration from the default location
func LoadConfig() (*Config, error) {
	configPath := DefaultConfigPath()
	if configPath == "" {
		// Can't determine home directory, use defaults
		return &Config{
			Enabled:  true,
			DebugLog: false,
		}, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		return &Config{
			Enabled:  true,
			DebugLog: false,
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set default log file if debug enabled but no log file specified
	if config.DebugLog && config.LogFile == "" {
		homeDir, _ := os.UserHomeDir()
		config.LogFile = filepath.Join(homeDir, ".claude", "bash-hook-debug.log")
	}

	return &config, nil
}

// SaveConfig saves the configuration to the default location
func SaveConfig(config *Config) error {
	configPath := DefaultConfigPath()
	if configPath == "" {
		return os.ErrNotExist
	}

	// Ensure .claude directory exists
	claudeDir := filepath.Dir(configPath)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write to file with restrictive permissions
	return os.WriteFile(configPath, data, 0600)
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents CLI configuration
type Config struct {
	ServerURL    string            `json:"server_url"`
	AuthToken    string            `json:"auth_token"`
	HistoryFile  string            `json:"history_file"`
	OutputFormat string            `json:"output_format"`
	Theme        string            `json:"theme"`
	AutoComplete bool              `json:"auto_complete"`
	Verbose      bool              `json:"verbose"`
	Aliases      map[string]string `json:"aliases"`
}

// Default returns default configuration
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".peervault")

	return &Config{
		ServerURL:    "http://localhost:8080",
		AuthToken:    "demo-token",
		HistoryFile:  filepath.Join(configDir, "history"),
		OutputFormat: "table",
		Theme:        "default",
		AutoComplete: true,
		Verbose:      false,
		Aliases: map[string]string{
			"ls":   "list",
			"bc":   "blockchain",
			"quit": "exit",
		},
	}
}

// Load loads configuration from file
func Load() (*Config, error) {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".peervault")
	configFile := filepath.Join(configDir, "config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config
		config := Default()
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".peervault")
	configFile := filepath.Join(configDir, "config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Set sets a configuration value
func (c *Config) Set(key, value string) error {
	switch key {
	case "server_url":
		c.ServerURL = value
	case "auth_token":
		c.AuthToken = value
	case "output_format":
		if value != "table" && value != "json" && value != "yaml" {
			return fmt.Errorf("invalid output format: %s (must be table, json, or yaml)", value)
		}
		c.OutputFormat = value
	case "theme":
		c.Theme = value
	case "auto_complete":
		c.AutoComplete = value == "true"
	case "verbose":
		c.Verbose = value == "true"
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return c.Save()
}

// Get gets a configuration value
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "server_url":
		return c.ServerURL, nil
	case "auth_token":
		return c.AuthToken, nil
	case "output_format":
		return c.OutputFormat, nil
	case "theme":
		return c.Theme, nil
	case "auto_complete":
		return fmt.Sprintf("%t", c.AutoComplete), nil
	case "verbose":
		return fmt.Sprintf("%t", c.Verbose), nil
	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// GetConfigDir returns the configuration directory
func GetConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".peervault")
}

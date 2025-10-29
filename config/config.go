package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Preset represents a configured archiving preset
type Preset struct {
	Name   string `yaml:"name"`
	Remote string `yaml:"remote"`
}

// Config represents the application configuration
type Config struct {
	Presets []Preset `yaml:"presets"`
}

// DefaultConfigContent is the default configuration if none exists
const DefaultConfigContent = `presets:
  - name: "Camera Pictures"
    remote: "/media/pictures/camera_pictures"
  - name: "Documents"
    remote: "/media/documents"
  - name: "Music"
    remote: "/media/music/archives"
`

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".jotta-archiver.yaml"), nil
}

// Load reads and parses the configuration file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(config.Presets) == 0 {
		return nil, fmt.Errorf("no presets defined in config file")
	}

	return &config, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) error {
	return os.WriteFile(path, []byte(DefaultConfigContent), 0644)
}


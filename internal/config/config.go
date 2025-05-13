package config

import (
	"os"
	"path/filepath"
	"simple-sops/pkg/logging"
)

// AppConfig represents the application configuration
type AppConfig struct {
	// KeyFile is the path to the Age key file
	KeyFile string
	// OnePasswordEnabled indicates whether to use 1Password for key storage
	OnePasswordEnabled bool
	// AlwaysUseOnePassword indicates whether to always get the key from 1Password for each operation
	AlwaysUseOnePassword bool
	// Debug mode
	Debug bool
	// Quiet mode
	Quiet bool
	// List of supported file extensions
	SupportedExtensions []string
}

// DefaultConfig returns the default application configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		KeyFile:              getDefaultKeyPath(),
		OnePasswordEnabled:   true,
		AlwaysUseOnePassword: true,
		Debug:                false,
		Quiet:                false,
		SupportedExtensions: []string{
			".yaml", ".yml", ".json", ".ini", ".env",
			".properties", ".toml", ".hcl", ".tfvars",
			".tfstate", ".pem", ".crt", ".key",
		},
	}
}

// getDefaultKeyPath returns the default path for the Age key file
func getDefaultKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logging.Error("Failed to get user home directory: %v", err)
		return ""
	}

	return filepath.Join(home, ".config", "simple-sops", "key.txt")
}

// IsSupportedFileType checks if a file is a supported type
func (c *AppConfig) IsSupportedFileType(filename string) bool {
	ext := filepath.Ext(filename)
	for _, supportedExt := range c.SupportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// GetConfigDir returns the directory where application config is stored
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "simple-sops")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return configDir, nil
}

// LoadConfig loads the application configuration
func LoadConfig() (*AppConfig, error) {
	// For now, just return the default config
	// In the future, this could load from a config file
	return DefaultConfig(), nil
}

package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"simple-sops/pkg/logging"
	"strings"

	"gopkg.in/yaml.v3"
)

// SopsConfig represents the structure of a .sops.yaml file
type SopsConfig struct {
	CreationRules []CreationRule `yaml:"creation_rules"`
}

// CreationRule represents a rule in the .sops.yaml file
type CreationRule struct {
	PathRegex      string `yaml:"path_regex"`
	Age            string `yaml:"age"`
	EncryptedRegex string `yaml:"encrypted_regex,omitempty"`
}

// GetSopsConfigPath returns the path to the .sops.yaml file
// If in a Git repository, returns the path at the root of the repository
// Otherwise, returns the path in the current directory
func GetSopsConfigPath() (string, error) {
	// Check if we're in a Git repository
	if isGitAvailable() {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		output, err := cmd.Output()
		if err == nil {
			repoRoot := strings.TrimSpace(string(output))
			configPath := filepath.Join(repoRoot, ".sops.yaml")
			logging.Debug("In Git repository. Using config path: %s", configPath)
			return configPath, nil
		}
	}

	// Not in a Git repository or git command failed
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(wd, ".sops.yaml")
	logging.Debug("Not in Git repository. Using config path: %s", configPath)
	return configPath, nil
}

// LoadSopsConfig loads the .sops.yaml file
func LoadSopsConfig(configPath string) (*SopsConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &SopsConfig{
			CreationRules: []CreationRule{},
		}, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SOPS config file: %w", err)
	}

	var config SopsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse SOPS config file: %w", err)
	}

	return &config, nil
}

// SaveSopsConfig saves the .sops.yaml file
func SaveSopsConfig(configPath string, config *SopsConfig) error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal SOPS config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write SOPS config file: %w", err)
	}

	return nil
}

// AddCreationRule adds or updates a rule in the .sops.yaml file
func AddCreationRule(config *SopsConfig, filename string, publicKey string, encryptedRegex string) error {
	// Check if a rule for this file already exists
	for i, rule := range config.CreationRules {
		if rule.PathRegex == filename {
			// Update existing rule
			config.CreationRules[i].Age = publicKey
			if encryptedRegex != "" {
				config.CreationRules[i].EncryptedRegex = encryptedRegex
			}
			// Don't return yet, we still need to check for the wildcard rule
			break
		}
	}

	// Create new rule if it doesn't exist
	ruleExists := false
	for _, rule := range config.CreationRules {
		if rule.PathRegex == filename {
			ruleExists = true
			break
		}
	}

	if !ruleExists {
		// Create new rule
		rule := CreationRule{
			PathRegex: filename,
			Age:       publicKey,
		}
		if encryptedRegex != "" {
			rule.EncryptedRegex = encryptedRegex
		}

		// Add the new rule at the beginning of the list
		config.CreationRules = append([]CreationRule{rule}, config.CreationRules...)
	}

	// Check if we already have a wildcard rule
	wildcardPattern := `.*\.(ya?ml|json|ini|env)`
	hasWildcard := false

	for _, rule := range config.CreationRules {
		if rule.PathRegex == wildcardPattern {
			hasWildcard = true
			break
		}
	}

	// Add the wildcard rule if it doesn't exist
	if !hasWildcard {
		wildcardRule := CreationRule{
			PathRegex: wildcardPattern,
			Age:       publicKey,
		}
		config.CreationRules = append(config.CreationRules, wildcardRule)
	}

	return nil
}

// RemoveCreationRule removes a rule from the .sops.yaml file
func RemoveCreationRule(config *SopsConfig, filename string) error {
	for i, rule := range config.CreationRules {
		if rule.PathRegex == filename {
			// Remove rule
			config.CreationRules = append(config.CreationRules[:i], config.CreationRules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("no rule found for %s", filename)
}

// CleanOrphanedRules removes rules for files that no longer exist
func CleanOrphanedRules(config *SopsConfig) (int, error) {
	var cleanedRules []CreationRule
	orphanedCount := 0

	// Keep only rules for wildcard patterns and existing files
	for _, rule := range config.CreationRules {
		// Keep rules with wildcard patterns
		if strings.Contains(rule.PathRegex, "*") || strings.Contains(rule.PathRegex, "?") {
			cleanedRules = append(cleanedRules, rule)
			continue
		}

		// Check if the file exists
		if _, err := os.Stat(rule.PathRegex); os.IsNotExist(err) {
			logging.Info("Removing orphaned rule for file: %s", rule.PathRegex)
			orphanedCount++
		} else {
			cleanedRules = append(cleanedRules, rule)
		}
	}

	config.CreationRules = cleanedRules
	return orphanedCount, nil
}

// isGitAvailable checks if Git is available on the system
func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// GetCreationRule gets the rule for a specific file
func GetCreationRule(config *SopsConfig, filename string) (CreationRule, bool) {
	for _, rule := range config.CreationRules {
		if rule.PathRegex == filename {
			return rule, true
		}
	}

	return CreationRule{}, false
}

// IsFileEncrypted checks if a file is encrypted using SOPS
func IsFileEncrypted(filePath string) bool {
	// Read the first few KB of the file to check for SOPS markers
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil {
		return false
	}

	content := string(buffer[:n])

	// Check for common SOPS encryption markers
	markers := []string{
		"sops:",
		"[sops]",
		"ENC[AES256_GCM",
		"sops_",
		"encrypted_suffix",
	}

	for _, marker := range markers {
		if strings.Contains(content, marker) {
			return true
		}
	}

	return false
}

// AddCreationRuleWithMultipleKeys adds or updates a rule in the .sops.yaml file with multiple keys
func AddCreationRuleWithMultipleKeys(config *SopsConfig, filename string, publicKeys string, encryptedRegex string) error {
	// Check if a rule for this file already exists
	for i, rule := range config.CreationRules {
		if rule.PathRegex == filename {
			// Update existing rule
			config.CreationRules[i].Age = publicKeys
			if encryptedRegex != "" {
				config.CreationRules[i].EncryptedRegex = encryptedRegex
			}
			// Don't return yet, we still need to check for the wildcard rule
			break
		}
	}

	// Create new rule if it doesn't exist
	ruleExists := false
	for _, rule := range config.CreationRules {
		if rule.PathRegex == filename {
			ruleExists = true
			break
		}
	}

	if !ruleExists {
		// Create new rule
		rule := CreationRule{
			PathRegex: filename,
			Age:       publicKeys,
		}
		if encryptedRegex != "" {
			rule.EncryptedRegex = encryptedRegex
		}

		// Add the new rule at the beginning of the list
		config.CreationRules = append([]CreationRule{rule}, config.CreationRules...)
	}

	// Check if we already have a wildcard rule
	wildcardPattern := `.*\.(ya?ml|json|ini|env)`
	hasWildcard := false

	for _, rule := range config.CreationRules {
		if rule.PathRegex == wildcardPattern {
			hasWildcard = true
			break
		}
	}

	// Add the wildcard rule if it doesn't exist
	if !hasWildcard {
		// Extract the first key from the comma-separated list
		firstKey := publicKeys
		if idx := strings.Index(publicKeys, ","); idx > 0 {
			firstKey = publicKeys[:idx]
		}

		wildcardRule := CreationRule{
			PathRegex: wildcardPattern,
			Age:       firstKey, // Use just the first key for the wildcard rule
		}
		config.CreationRules = append(config.CreationRules, wildcardRule)
	}

	return nil
}

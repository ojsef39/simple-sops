// internal/config/sops_config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSopsConfig(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, ".sops.yaml")
	configContent := `creation_rules:
  - path_regex: test.env
    age: age123
  - path_regex: .*\.(ya?ml|json|ini|env)
    age: age456
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test loading config
	config, err := LoadSopsConfig(configPath)
	if err != nil {
		t.Fatalf("LoadSopsConfig failed: %v", err)
	}

	// Verify the config was loaded correctly
	if len(config.CreationRules) != 2 {
		t.Errorf("Expected 2 creation rules, got %d", len(config.CreationRules))
	}

	// Check first rule
	if config.CreationRules[0].PathRegex != "test.env" {
		t.Errorf("Expected path_regex 'test.env', got '%s'", config.CreationRules[0].PathRegex)
	}
	if config.CreationRules[0].Age != "age123" {
		t.Errorf("Expected age 'age123', got '%s'", config.CreationRules[0].Age)
	}

	// Test loading non-existent config
	config, err = LoadSopsConfig(filepath.Join(tempDir, "nonexistent.yaml"))
	if err != nil {
		t.Fatalf("LoadSopsConfig should not fail for missing file: %v", err)
	}
	if len(config.CreationRules) != 0 {
		t.Errorf("Expected empty config for non-existent file")
	}
}

func TestAddCreationRuleWithMultipleKeys(t *testing.T) {
	// Create a new config
	config := &SopsConfig{
		CreationRules: []CreationRule{},
	}

	// Test adding a rule with multiple keys
	err := AddCreationRuleWithMultipleKeys(config, "test.env", "age123,age456", "")
	if err != nil {
		t.Fatalf("AddCreationRuleWithMultipleKeys failed: %v", err)
	}

	// Verify the rule was added correctly
	if len(config.CreationRules) != 2 { // One for the file, one for wildcard
		t.Errorf("Expected 2 creation rules, got %d", len(config.CreationRules))
	}

	// Check the added rule
	found := false
	for _, rule := range config.CreationRules {
		if rule.PathRegex == "test.env" {
			found = true
			if rule.Age != "age123,age456" {
				t.Errorf("Expected age 'age123,age456', got '%s'", rule.Age)
			}
		}
	}

	if !found {
		t.Errorf("Rule for 'test.env' not found in config")
	}

	// Test updating an existing rule
	err = AddCreationRuleWithMultipleKeys(config, "test.env", "age789,age101112", "")
	if err != nil {
		t.Fatalf("AddCreationRuleWithMultipleKeys failed when updating: %v", err)
	}

	// Verify the rule was updated correctly
	for _, rule := range config.CreationRules {
		if rule.PathRegex == "test.env" {
			if rule.Age != "age789,age101112" {
				t.Errorf("Expected updated age 'age789,age101112', got '%s'", rule.Age)
			}
		}
	}
}

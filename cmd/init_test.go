package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	t.Run("creates config file successfully", func(t *testing.T) {
		// Run init command
		cmd := rootCmd
		cmd.SetArgs([]string{"init"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("init command failed: %v", err)
		}

		// Check that file was created
		configPath := ".prtool.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Configuration file was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		contentStr := string(content)

		// Check for key sections
		expectedSections := []string{
			"# prtool Configuration File",
			"github_token:",
			"org:",
			"team:",
			"user:",
			"repo:",
			"since:",
			"llm_provider:",
			"llm_api_key:",
			"llm_model:",
			"output:",
			"dry_run:",
			"verbose:",
			"ci:",
		}

		for _, section := range expectedSections {
			if !strings.Contains(contentStr, section) {
				t.Errorf("Expected section '%s' not found in config file", section)
			}
		}

		// Check for comments explaining options
		expectedComments := []string{
			"# GitHub configuration",
			"# Scope configuration",
			"# LLM configuration",
			"# Output configuration",
			"# Behavior flags",
		}

		for _, comment := range expectedComments {
			if !strings.Contains(contentStr, comment) {
				t.Errorf("Expected comment '%s' not found in config file", comment)
			}
		}
	})

	t.Run("fails when config file already exists", func(t *testing.T) {
		// Create a second temp directory
		tempDir2 := t.TempDir()
		if err := os.Chdir(tempDir2); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Create existing config file
		configPath := ".prtool.yaml"
		if err := os.WriteFile(configPath, []byte("existing"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Run init command
		cmd := rootCmd
		cmd.SetArgs([]string{"init"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Expected init command to fail when config file exists")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected error about existing file, got: %v", err)
		}
	})
}

func TestGenerateAnnotatedYAML(t *testing.T) {
	content := generateAnnotatedYAML()

	// Test that content is non-empty
	if len(content) == 0 {
		t.Error("Generated YAML content is empty")
	}

	// Test that it contains YAML structure
	if !strings.Contains(content, ":") {
		t.Error("Generated content doesn't appear to be YAML")
	}

	// Test for required fields
	requiredFields := []string{
		"github_token",
		"org",
		"team",
		"user",
		"repo",
		"since",
		"llm_provider",
		"llm_api_key",
		"llm_model",
		"output",
		"dry_run",
		"verbose",
		"ci",
	}

	for _, field := range requiredFields {
		if !strings.Contains(content, field+":") {
			t.Errorf("Required field '%s' not found in generated YAML", field)
		}
	}

	// Test for environment variable documentation
	envVars := []string{
		"PRTOOL_GITHUB_TOKEN",
		"PRTOOL_ORG",
		"PRTOOL_TEAM",
		"PRTOOL_USER",
		"PRTOOL_REPO",
		"PRTOOL_SINCE",
		"PRTOOL_LLM_PROVIDER",
		"PRTOOL_LLM_API_KEY",
		"PRTOOL_LLM_MODEL",
	}

	for _, envVar := range envVars {
		if !strings.Contains(content, envVar) {
			t.Errorf("Environment variable '%s' not documented in generated YAML", envVar)
		}
	}
}

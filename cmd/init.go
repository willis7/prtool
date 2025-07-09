package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a sample configuration file",
	Long: `Generate a sample prtool configuration file (.prtool.yaml) in the current directory.
This file contains all available configuration options with explanatory comments.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := ".prtool.yaml"

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file %s already exists", configPath)
	}

	// Generate annotated YAML content
	content := generateAnnotatedYAML()

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("Configuration file created: %s\n", absPath)
	fmt.Println("Edit this file to customize your prtool settings.")

	return nil
}

func generateAnnotatedYAML() string {
	return `# prtool Configuration File
# This file contains all available configuration options for prtool.
# Values can be overridden by environment variables or command-line flags.

# GitHub configuration
# Required: Your GitHub personal access token
# Environment variable: PRTOOL_GITHUB_TOKEN
github_token: ""

# Scope configuration (choose ONE of the following)
# Organization to scan for repositories
# Environment variable: PRTOOL_ORG
org: ""

# Team within an organization to scan
# Environment variable: PRTOOL_TEAM  
team: ""

# Specific user to scan repositories for
# Environment variable: PRTOOL_USER
user: ""

# Specific repository to scan (format: owner/repo)
# Environment variable: PRTOOL_REPO
repo: ""

# Time range configuration
# How far back to look for merged PRs (e.g., "-7d", "-1m", "-1yr")
# Environment variable: PRTOOL_SINCE
since: "-7d"

# LLM configuration
# LLM provider: "stub", "openai", or "ollama"
# Environment variable: PRTOOL_LLM_PROVIDER
llm_provider: "stub"

# API key for LLM provider (required for OpenAI)
# Environment variable: PRTOOL_LLM_API_KEY
llm_api_key: ""

# Model name for LLM provider
# OpenAI: "gpt-3.5-turbo", "gpt-4", etc.
# Ollama: "llama3.2", "codellama", etc.
# Environment variable: PRTOOL_LLM_MODEL
llm_model: ""

# Custom prompt file path (optional)
# Environment variable: PRTOOL_PROMPT
prompt: ""

# Output configuration
# Output file path (leave empty for stdout)
# Environment variable: PRTOOL_OUTPUT
output: ""

# Log file path (leave empty for no file logging)
# Environment variable: PRTOOL_LOG_FILE
log_file: ""

# Behavior flags
# Skip LLM processing and show PR data only
# Environment variable: PRTOOL_DRY_RUN
dry_run: false

# Enable verbose logging
# Environment variable: PRTOOL_VERBOSE
verbose: false

# Non-interactive mode for CI environments
# Environment variable: PRTOOL_CI
ci: false
`
}

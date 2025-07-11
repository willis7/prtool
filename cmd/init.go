package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a sample configuration file",
	Long:  `Generates a sample .prtool.yaml configuration file in the current directory with annotations explaining each option.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	filename := ".prtool.yaml"
	
	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("configuration file %s already exists", filename)
	}

	// Generate annotated YAML content
	content := generateConfigTemplate()

	// Write to file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	absPath, _ := filepath.Abs(filename)
	fmt.Printf("Created configuration file: %s\n", absPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the configuration file to add your GitHub token and preferences")
	fmt.Println("2. Run 'prtool run' to fetch and summarize pull requests")
	
	return nil
}

func generateConfigTemplate() string {
	return `# prtool configuration file
# This file configures how prtool fetches and summarizes pull requests

# GitHub configuration
github:
  # Personal access token for GitHub API
  # Required: Create at https://github.com/settings/tokens
  # Needs: repo (full), read:org, read:user permissions
  token: ""
  
  # Scope configuration - specify exactly one of these:
  # organization: "myorg"        # Fetch PRs from all repos in an organization
  # team: "myteam"              # Fetch PRs from repos accessible to a team
  # user: "myusername"          # Fetch PRs from a user's repos
  # repositories:               # Fetch PRs from specific repositories
  #   - "owner/repo1"
  #   - "owner/repo2"

# LLM (Language Model) configuration for generating summaries
llm:
  # Provider: "openai", "ollama", or "stub" (for testing)
  provider: "openai"
  
  # Model name (e.g., "gpt-3.5-turbo", "gpt-4" for OpenAI; "llama2", "codellama" for Ollama)
  model: "gpt-3.5-turbo"
  
  # API key for OpenAI (not needed for Ollama)
  # Get from: https://platform.openai.com/api-keys
  api_key: ""
  
  # Base URL (only for Ollama or custom endpoints)
  # base_url: "http://localhost:11434"
  
  # Temperature for text generation (0.0-1.0, higher = more creative)
  temperature: 0.7
  
  # Maximum tokens for summary generation
  max_tokens: 500

# Output configuration
output:
  # Output format (currently only "markdown" is supported)
  format: "markdown"
  
  # Output file path (leave empty to print to stdout)
  # file: "pr-summary.md"

# General options
# Enable verbose output for debugging
verbose: false

# Dry run mode - show what would be processed without calling LLM
dry_run: false

# Time range for fetching PRs (examples: "-7d", "-1m", "-1yr")
# time_range: "-7d"

# Note: Command-line flags override these settings
# Example: prtool run --github-token=xyz --since=-14d
`
}

// WriteConfigFile is exported for testing
func WriteConfigFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
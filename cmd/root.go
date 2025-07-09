package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/willis7/prtool/internal/config"
)

var version = "dev"

// CLI flags
var (
	cfgFile     string
	githubToken string
	org         string
	team        string
	user        string
	repo        string
	since       string
	llmProvider string
	llmAPIKey   string
	llmModel    string
	prompt      string
	output      string
	dryRun      bool
	verbose     bool
	ci          bool
	logFile     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prtool",
	Short: "A CLI tool for summarizing GitHub pull requests",
	Long: `prtool is a command-line tool that fetches GitHub pull requests (PRs) 
for a specified time period and scope (organization, team, user, or repository), 
summarizes them using an LLM (OpenAI or Ollama), and outputs the result in Markdown format.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.prtool.yaml)")
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// GitHub flags
	rootCmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub personal access token")

	// Scope flags (mutually exclusive)
	rootCmd.Flags().StringVar(&org, "org", "", "GitHub organization")
	rootCmd.Flags().StringVar(&team, "team", "", "GitHub team (format: org/team)")
	rootCmd.Flags().StringVar(&user, "user", "", "GitHub user")
	rootCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository (format: owner/repo)")

	// Time range
	rootCmd.Flags().StringVar(&since, "since", "", "Time range (e.g., -7d, -1m, -1yr)")

	// LLM flags
	rootCmd.Flags().StringVar(&llmProvider, "llm-provider", "", "LLM provider (openai, ollama)")
	rootCmd.Flags().StringVar(&llmAPIKey, "llm-api-key", "", "LLM API key")
	rootCmd.Flags().StringVar(&llmModel, "llm-model", "", "LLM model name")
	rootCmd.Flags().StringVar(&prompt, "prompt", "", "Path to custom prompt file")

	// Output flags
	rootCmd.Flags().StringVar(&output, "output", "", "Output file path")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Skip LLM processing and show PR data")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.Flags().BoolVar(&ci, "ci", false, "Non-interactive mode for CI")
	rootCmd.Flags().StringVar(&logFile, "log-file", "", "Log file path")

	// Handle version flag and basic command execution
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println(version)
			return
		}

		// For now, just show help if no version flag
		cmd.Help()
	}
}

// GetConfig loads and merges configuration from all sources
func GetConfig() (*config.Config, error) {
	// Load from YAML file
	configPath := cfgFile
	if configPath == "" {
		configPath = "~/.prtool.yaml"
	}

	yamlConfig, err := config.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Load from environment
	envConfig := config.LoadFromEnv()

	// Create CLI config from flags
	cliConfig := &config.Config{
		GitHubToken: githubToken,
		Org:         org,
		Team:        team,
		User:        user,
		Repo:        repo,
		Since:       since,
		LLMProvider: llmProvider,
		LLMAPIKey:   llmAPIKey,
		LLMModel:    llmModel,
		Prompt:      prompt,
		Output:      output,
		DryRun:      dryRun,
		Verbose:     verbose,
		CI:          ci,
		LogFile:     logFile,
	}

	// Merge with precedence: CLI > env > YAML
	merged := config.MergeConfig(cliConfig, envConfig, yamlConfig)

	return merged, nil
}

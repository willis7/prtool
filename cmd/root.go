package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/llm"
	"github.com/willis7/prtool/internal/model"
	"github.com/willis7/prtool/internal/render"
	"github.com/willis7/prtool/internal/scope"
	"github.com/willis7/prtool/internal/service"
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

		// Load configuration
		cfg, err := GetConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Validate configuration
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(1)
		}

		// Create GitHub client
		ghClient, err := gh.NewRestClient(cfg.GitHubToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create GitHub client: %v\n", err)
			os.Exit(1)
		}

		// Fetch PRs
		prs, err := service.Fetch(cfg, ghClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch PRs: %v\n", err)
			os.Exit(1)
		}

		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "Fetched %d pull requests\n", len(prs))
		}

		// Handle dry-run mode
		if cfg.DryRun {
			fmt.Print(render.RenderTable(prs))
			return
		}

		// Generate metadata
		metadata := generateMetadata(cfg, prs)

		// Generate LLM summary if not in dry-run mode
		if !cfg.DryRun {
			llmClient := createLLMClient(cfg)
			if llmClient != nil {
				if cfg.Verbose {
					fmt.Fprintf(os.Stderr, "Generating AI summary...\n")
				}

				context := llm.BuildContext(prs)
				summary, err := llmClient.Summarise(context)
				if err != nil {
					if cfg.Verbose {
						fmt.Fprintf(os.Stderr, "Warning: Failed to generate AI summary: %v\n", err)
					}
					// Continue without summary rather than failing completely
				} else {
					metadata.Summary = summary
				}
			}
		}

		// Render markdown
		markdownOutput := render.Render(metadata, prs)

		// Output to file or stdout
		if cfg.Output != "" {
			if err := writeToFile(cfg.Output, markdownOutput); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write output file: %v\n", err)
				os.Exit(1)
			}
			if cfg.Verbose {
				fmt.Fprintf(os.Stderr, "Output written to: %s\n", cfg.Output)
			}
		} else {
			fmt.Print(markdownOutput)
		}
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

// validateConfig validates the configuration
func validateConfig(cfg *config.Config) error {
	if cfg.GitHubToken == "" {
		return fmt.Errorf("GitHub token is required")
	}

	// Validate scope using the scope package
	if err := scope.ValidateScope(cfg); err != nil {
		return err
	}

	return nil
}

// generateMetadata creates metadata for the report
func generateMetadata(cfg *config.Config, prs []*model.PR) render.Metadata {
	// Determine scope type and value
	var scopeType, scopeValue string
	if cfg.Org != "" {
		scopeType, scopeValue = "organization", cfg.Org
	} else if cfg.Team != "" {
		scopeType, scopeValue = "team", cfg.Team
	} else if cfg.User != "" {
		scopeType, scopeValue = "user", cfg.User
	} else if cfg.Repo != "" {
		scopeType, scopeValue = "repository", cfg.Repo
	}

	// Collect unique repositories
	repoSet := make(map[string]bool)
	for _, pr := range prs {
		if pr.Repository != "" {
			repoSet[pr.Repository] = true
		}
	}

	var repositories []string
	for repo := range repoSet {
		repositories = append(repositories, repo)
	}

	// Determine since value
	since := cfg.Since
	if since == "" {
		since = "-7d" // default
	}

	return render.Metadata{
		GeneratedAt:  time.Now().UTC(),
		Scope:        scopeType,
		ScopeValue:   scopeValue,
		Since:        since,
		TotalPRs:     len(prs),
		Repositories: repositories,
		LLMProvider:  cfg.LLMProvider,
		LLMModel:     cfg.LLMModel,
		Summary:      "", // Will be filled by LLM in later iterations
	}
}

// writeToFile writes content to a file
func writeToFile(filename, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// createLLMClient creates an LLM client based on configuration
func createLLMClient(cfg *config.Config) llm.LLM {
	if cfg.LLMProvider == "" {
		// Default to stub for testing
		return llm.NewStubLLM()
	}

	switch cfg.LLMProvider {
	case "stub":
		return llm.NewStubLLM()
	case "openai":
		if cfg.LLMAPIKey == "" {
			fmt.Fprintf(os.Stderr, "Warning: OpenAI API key not provided, falling back to stub\n")
			return llm.NewStubLLM()
		}
		return llm.NewOpenAILLM(cfg.LLMAPIKey, cfg.LLMModel)
	case "ollama":
		return llm.NewOllamaLLM("", cfg.LLMModel) // Use default localhost URL
	default:
		// Unsupported provider, return stub as fallback
		fmt.Fprintf(os.Stderr, "Warning: Unknown LLM provider '%s', falling back to stub\n", cfg.LLMProvider)
		return llm.NewStubLLM()
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedConfig *Config
		expectError    bool
	}{
		{
			name: "valid yaml config",
			configContent: `
github_token: "test-token"
org: "test-org"
since: "-7d"
llm_provider: "openai"
llm_api_key: "test-key"
output: "test.md"
dry_run: true
verbose: false
`,
			expectedConfig: &Config{
				GitHubToken: "test-token",
				Org:         "test-org",
				Since:       "-7d",
				LLMProvider: "openai",
				LLMAPIKey:   "test-key",
				Output:      "test.md",
				DryRun:      true,
				Verbose:     false,
			},
			expectError: false,
		},
		{
			name:           "empty path returns empty config",
			configContent:  "",
			expectedConfig: &Config{},
			expectError:    false,
		},
		{
			name:           "invalid yaml returns error",
			configContent:  "invalid: yaml: content:",
			expectedConfig: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string

			if tt.name == "empty path returns empty config" {
				configPath = ""
			} else {
				// Create temporary config file
				tempDir := t.TempDir()
				configPath = filepath.Join(tempDir, "config.yaml")

				if tt.configContent != "" {
					err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
					if err != nil {
						t.Fatalf("Failed to write test config file: %v", err)
					}
				}
			}

			config, err := LoadFromFile(configPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !configsEqual(config, tt.expectedConfig) {
				t.Errorf("Config mismatch.\nGot: %+v\nWant: %+v", config, tt.expectedConfig)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "all env vars set",
			envVars: map[string]string{
				"PRTOOL_GITHUB_TOKEN": "env-token",
				"PRTOOL_ORG":          "env-org",
				"PRTOOL_SINCE":        "-30d",
				"PRTOOL_LLM_PROVIDER": "ollama",
				"PRTOOL_LLM_API_KEY":  "env-key",
				"PRTOOL_OUTPUT":       "env.md",
				"PRTOOL_DRY_RUN":      "true",
				"PRTOOL_VERBOSE":      "false",
				"PRTOOL_CI":           "true",
			},
			expected: &Config{
				GitHubToken: "env-token",
				Org:         "env-org",
				Since:       "-30d",
				LLMProvider: "ollama",
				LLMAPIKey:   "env-key",
				Output:      "env.md",
				DryRun:      true,
				Verbose:     false,
				CI:          true,
			},
		},
		{
			name:    "no env vars set",
			envVars: map[string]string{},
			expected: &Config{
				DryRun:  false,
				Verbose: false,
				CI:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars first
			envVarNames := []string{
				"PRTOOL_GITHUB_TOKEN", "PRTOOL_ORG", "PRTOOL_TEAM", "PRTOOL_USER", "PRTOOL_REPO",
				"PRTOOL_SINCE", "PRTOOL_LLM_PROVIDER", "PRTOOL_LLM_API_KEY", "PRTOOL_LLM_MODEL",
				"PRTOOL_PROMPT", "PRTOOL_OUTPUT", "PRTOOL_DRY_RUN", "PRTOOL_VERBOSE", "PRTOOL_CI",
				"PRTOOL_LOG_FILE",
			}

			originalValues := make(map[string]string)
			for _, name := range envVarNames {
				originalValues[name] = os.Getenv(name)
				_ = os.Unsetenv(name) // Ignore error in test cleanup
			}

			// Restore original values after test
			defer func() {
				for name, value := range originalValues {
					if value != "" {
						_ = os.Setenv(name, value) // Ignore error in test cleanup
					} else {
						_ = os.Unsetenv(name) // Ignore error in test cleanup
					}
				}
			}()

			// Set test env vars
			for name, value := range tt.envVars {
				_ = os.Setenv(name, value) // Ignore error in test setup
			}

			config := LoadFromEnv()

			if !configsEqual(config, tt.expected) {
				t.Errorf("Config mismatch.\nGot: %+v\nWant: %+v", config, tt.expected)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name       string
		cliConfig  *Config
		envConfig  *Config
		yamlConfig *Config
		expected   *Config
	}{
		{
			name: "cli takes precedence over env and yaml",
			cliConfig: &Config{
				GitHubToken: "cli-token",
				Org:         "cli-org",
				DryRun:      true,
			},
			envConfig: &Config{
				GitHubToken: "env-token",
				Since:       "-7d",
				Verbose:     true,
			},
			yamlConfig: &Config{
				GitHubToken: "yaml-token",
				LLMProvider: "openai",
				Output:      "yaml.md",
			},
			expected: &Config{
				GitHubToken: "cli-token", // CLI wins
				Org:         "cli-org",   // CLI only
				Since:       "-7d",       // ENV only
				LLMProvider: "openai",    // YAML only
				Output:      "yaml.md",   // YAML only
				DryRun:      true,        // CLI wins
				Verbose:     true,        // ENV only
			},
		},
		{
			name:      "env takes precedence over yaml",
			cliConfig: &Config{
				// Empty CLI config
			},
			envConfig: &Config{
				GitHubToken: "env-token",
				Since:       "-30d",
			},
			yamlConfig: &Config{
				GitHubToken: "yaml-token",
				Since:       "-7d",
				LLMProvider: "yaml-provider",
			},
			expected: &Config{
				GitHubToken: "env-token",     // ENV wins over YAML
				Since:       "-30d",          // ENV wins over YAML
				LLMProvider: "yaml-provider", // YAML only
			},
		},
		{
			name:      "handles nil configs",
			cliConfig: nil,
			envConfig: nil,
			yamlConfig: &Config{
				GitHubToken: "yaml-token",
				DryRun:      true,
			},
			expected: &Config{
				GitHubToken: "yaml-token",
				DryRun:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeConfig(tt.cliConfig, tt.envConfig, tt.yamlConfig)

			if !configsEqual(result, tt.expected) {
				t.Errorf("Config mismatch.\nGot: %+v\nWant: %+v", result, tt.expected)
			}
		})
	}
}

// configsEqual compares two Config structs for equality
func configsEqual(a, b *Config) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.GitHubToken == b.GitHubToken &&
		a.Org == b.Org &&
		a.Team == b.Team &&
		a.User == b.User &&
		a.Repo == b.Repo &&
		a.Since == b.Since &&
		a.LLMProvider == b.LLMProvider &&
		a.LLMAPIKey == b.LLMAPIKey &&
		a.LLMModel == b.LLMModel &&
		a.Prompt == b.Prompt &&
		a.Output == b.Output &&
		a.DryRun == b.DryRun &&
		a.Verbose == b.Verbose &&
		a.CI == b.CI &&
		a.LogFile == b.LogFile
}

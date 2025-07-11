package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		want        *Config
		wantErr     bool
	}{
		{
			name: "valid config",
			yamlContent: `
github:
  token: "file-token"
  organization: "test-org"
llm:
  provider: "openai"
  model: "gpt-4"
  temperature: 0.5
verbose: true
`,
			want: &Config{
				GitHub: GitHubConfig{
					Token:        "file-token",
					Organization: "test-org",
				},
				LLM: LLMConfig{
					Provider:    "openai",
					Model:       "gpt-4",
					Temperature: 0.5,
					MaxTokens:   2000, // default
				},
				Output: OutputConfig{
					Format: "markdown", // default
				},
				Verbose: true,
			},
			wantErr: false,
		},
		{
			name:        "invalid yaml",
			yamlContent: `invalid: [yaml content`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := LoadFromFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadFromFile() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
	}{
		{
			name: "all env vars set",
			envVars: map[string]string{
				"PRTOOL_GITHUB_TOKEN":         "env-token",
				"PRTOOL_GITHUB_ORGANIZATION":  "env-org",
				"PRTOOL_GITHUB_TEAM":          "env-team",
				"PRTOOL_GITHUB_USER":          "env-user",
				"PRTOOL_GITHUB_REPOSITORIES":  "repo1,repo2",
				"PRTOOL_LLM_PROVIDER":         "ollama",
				"PRTOOL_LLM_MODEL":            "llama2",
				"PRTOOL_LLM_API_KEY":          "env-key",
				"PRTOOL_LLM_BASE_URL":         "http://localhost:11434",
				"PRTOOL_LLM_TEMPERATURE":      "0.9",
				"PRTOOL_LLM_MAX_TOKENS":       "1500",
				"PRTOOL_OUTPUT_FORMAT":        "json",
				"PRTOOL_OUTPUT_FILE":          "output.json",
				"PRTOOL_VERBOSE":              "true",
				"PRTOOL_DRY_RUN":              "1",
			},
			want: &Config{
				GitHub: GitHubConfig{
					Token:        "env-token",
					Organization: "env-org",
					Team:         "env-team",
					User:         "env-user",
					Repositories: []string{"repo1", "repo2"},
				},
				LLM: LLMConfig{
					Provider:    "ollama",
					Model:       "llama2",
					APIKey:      "env-key",
					BaseURL:     "http://localhost:11434",
					Temperature: 0.9,
					MaxTokens:   1500,
				},
				Output: OutputConfig{
					Format: "json",
					File:   "output.json",
				},
				Verbose: true,
				DryRun:  true,
			},
		},
		{
			name:    "no env vars set",
			envVars: map[string]string{},
			want: &Config{
				LLM: LLMConfig{
					Temperature: 0.7,
					MaxTokens:   2000,
				},
				Output: OutputConfig{
					Format: "markdown",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			for _, key := range []string{
				"PRTOOL_GITHUB_TOKEN", "PRTOOL_GITHUB_ORGANIZATION", "PRTOOL_GITHUB_TEAM",
				"PRTOOL_GITHUB_USER", "PRTOOL_GITHUB_REPOSITORIES", "PRTOOL_LLM_PROVIDER",
				"PRTOOL_LLM_MODEL", "PRTOOL_LLM_API_KEY", "PRTOOL_LLM_BASE_URL",
				"PRTOOL_LLM_TEMPERATURE", "PRTOOL_LLM_MAX_TOKENS", "PRTOOL_OUTPUT_FORMAT",
				"PRTOOL_OUTPUT_FILE", "PRTOOL_VERBOSE", "PRTOOL_DRY_RUN",
			} {
				os.Unsetenv(key)
			}

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got := LoadFromEnv()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadFromEnv() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name    string
		base    *Config
		overlay *Config
		want    *Config
	}{
		{
			name: "overlay overwrites base",
			base: &Config{
				GitHub: GitHubConfig{
					Token:        "base-token",
					Organization: "base-org",
				},
				LLM: LLMConfig{
					Provider:    "openai",
					Temperature: 0.7,
					MaxTokens:   2000,
				},
				Verbose: false,
			},
			overlay: &Config{
				GitHub: GitHubConfig{
					Token: "overlay-token",
					User:  "overlay-user",
				},
				LLM: LLMConfig{
					Provider:    "ollama",
					Temperature: 0.9,
				},
				Verbose: true,
			},
			want: &Config{
				GitHub: GitHubConfig{
					Token:        "overlay-token",
					Organization: "base-org",
					User:         "overlay-user",
				},
				LLM: LLMConfig{
					Provider:    "ollama",
					Temperature: 0.9,
					MaxTokens:   2000,
				},
				Output: OutputConfig{}, // zero value
				Verbose: true,
			},
		},
		{
			name: "nil base returns overlay",
			base: nil,
			overlay: &Config{
				GitHub: GitHubConfig{
					Token: "overlay-token",
				},
			},
			want: &Config{
				GitHub: GitHubConfig{
					Token: "overlay-token",
				},
				LLM: LLMConfig{
					Temperature: 0.7,
					MaxTokens:   2000,
				},
				Output: OutputConfig{
					Format: "markdown",
				},
			},
		},
		{
			name: "nil overlay returns base",
			base: &Config{
				GitHub: GitHubConfig{
					Token: "base-token",
				},
			},
			overlay: nil,
			want: &Config{
				GitHub: GitHubConfig{
					Token: "base-token",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeConfig(tt.base, tt.overlay)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigPrecedence(t *testing.T) {
	// This test simulates the full precedence chain: file -> env -> CLI
	fileConfig := &Config{
		GitHub: GitHubConfig{
			Token:        "file-token",
			Organization: "file-org",
			Team:         "file-team",
		},
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-3.5",
			Temperature: 0.5,
			MaxTokens:   1000,
		},
		Verbose: false,
	}

	envConfig := &Config{
		GitHub: GitHubConfig{
			Token: "env-token",
			User:  "env-user",
		},
		LLM: LLMConfig{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   1500,
		},
		Output: OutputConfig{
			Format: "json",
		},
		Verbose: true,
	}

	cliConfig := &Config{
		GitHub: GitHubConfig{
			Token: "cli-token",
		},
		LLM: LLMConfig{
			Temperature: 0.9,
		},
		DryRun: true,
	}

	// Apply precedence: file -> env -> CLI
	result := MergeConfig(fileConfig, envConfig)
	result = MergeConfig(result, cliConfig)

	expected := &Config{
		GitHub: GitHubConfig{
			Token:        "cli-token",     // CLI wins
			Organization: "file-org",      // Only in file
			Team:         "file-team",     // Only in file
			User:         "env-user",      // Only in env
		},
		LLM: LLMConfig{
			Provider:    "openai",    // Only in file
			Model:       "gpt-4",     // Env overwrites file
			Temperature: 0.9,         // CLI wins
			MaxTokens:   1500,        // Env overwrites file
		},
		Output: OutputConfig{
			Format: "json",           // Only in env
		},
		Verbose: true,               // Env overwrites file
		DryRun:  true,               // Only in CLI
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Config precedence failed:\ngot  = %+v\nwant = %+v", result, expected)
	}
}
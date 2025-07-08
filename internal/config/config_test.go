package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestLoadFromFile(t *testing.T) {
	content := `
org: test-org
since: -1m
`
	tmpfile, err := os.CreateTemp("", "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := LoadFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if config.Org != "test-org" {
		t.Errorf("expected org to be 'test-org', got '%s'", config.Org)
	}
	if config.Since != "-1m" {
		t.Errorf("expected since to be '-1m', got '%s'", config.Since)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PRTOOL_ORG", "env-org")
	t.Setenv("PRTOOL_SINCE", "-2m")

	config := LoadFromEnv()

	if config.Org != "env-org" {
		t.Errorf("expected org to be 'env-org', got '%s'", config.Org)
	}
	if config.Since != "-2m" {
		t.Errorf("expected since to be '-2m', got '%s'", config.Since)
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name     string
		cli      *Config
		env      *Config
		yaml     *Config
		expected *Config
	}{
		{
			name: "CLI overrides all",
			cli:  &Config{Org: "cli-org", Since: "-1d"},
			env:  &Config{Org: "env-org", Since: "-1w"},
			yaml: &Config{Org: "yaml-org", Since: "-1y"},
			expected: &Config{Org: "cli-org", Since: "-1d"},
		},
		{
			name: "Env overrides YAML",
			cli:  &Config{Since: "-7d"}, // Default value
			env:  &Config{Org: "env-org", Since: "-1w"},
			yaml: &Config{Org: "yaml-org", Since: "-1y"},
			expected: &Config{Org: "env-org", Since: "-1w"},
		},
		{
			name: "YAML only",
			cli:  &Config{Since: "-7d"}, // Default value
			env:  &Config{},
			yaml: &Config{Org: "yaml-org", Since: "-1y"},
			expected: &Config{Org: "yaml-org", Since: "-1y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := MergeConfig(tt.cli, tt.env, tt.yaml)
			if merged.Org != tt.expected.Org {
				t.Errorf("expected org '%s', got '%s'", tt.expected.Org, merged.Org)
			}
			if merged.Since != tt.expected.Since {
				t.Errorf("expected since '%s', got '%s'", tt.expected.Since, merged.Since)
			}
		})
	}
}

func TestBindFlags(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	config := BindFlags(flags)

	if err := flags.Parse([]string{"--org", "flag-org", "--since", "-3d"}); err != nil {
		t.Fatal(err)
	}

	if config.Org != "flag-org" {
		t.Errorf("expected org to be 'flag-org', got '%s'", config.Org)
	}
	if config.Since != "-3d" {
		t.Errorf("expected since to be '-3d', got '%s'", config.Since)
	}
}

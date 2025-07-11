# prtool

A command-line tool for fetching and summarizing GitHub pull requests using AI.

## Features

- Fetch merged pull requests from GitHub repositories
- Generate AI-powered summaries using OpenAI or Ollama
- Flexible scope: organization, team, user, or specific repositories
- Multiple configuration options: YAML files, environment variables, CLI flags
- Dry-run mode for testing without LLM calls
- CI mode for automation with structured output and exit codes
- Comprehensive logging with verbose and file output options

## Installation

```bash
go install github.com/yourorg/prtool@latest
```

## Quick Start

1. Generate a configuration file:
   ```bash
   prtool init
   ```

2. Edit `.prtool.yaml` to add your GitHub token and preferences.

3. Fetch and summarize PRs:
   ```bash
   prtool run
   ```

## Configuration

prtool supports three configuration methods with the following precedence:
1. CLI flags (highest priority)
2. Environment variables
3. YAML configuration file (lowest priority)

### Configuration File

Generate a sample configuration:
```bash
prtool init
```

This creates `.prtool.yaml` with all available options documented.

### Environment Variables

- `PRTOOL_GITHUB_TOKEN` - GitHub personal access token
- `PRTOOL_GITHUB_ORGANIZATION` - GitHub organization
- `PRTOOL_GITHUB_TEAM` - GitHub team
- `PRTOOL_GITHUB_USER` - GitHub user
- `PRTOOL_GITHUB_REPOSITORIES` - Comma-separated list of repositories
- `PRTOOL_LLM_PROVIDER` - LLM provider (openai, ollama, stub)
- `PRTOOL_LLM_MODEL` - LLM model name
- `PRTOOL_LLM_API_KEY` - LLM API key
- `PRTOOL_LLM_BASE_URL` - LLM base URL (for Ollama)
- `PRTOOL_OUTPUT_FORMAT` - Output format (markdown)
- `PRTOOL_OUTPUT_FILE` - Output file path
- `PRTOOL_VERBOSE` - Enable verbose output (true/false)
- `PRTOOL_DRY_RUN` - Enable dry-run mode (true/false)
- `PRTOOL_CI` - Enable CI mode (true/false)
- `PRTOOL_LOG_FILE` - Log file path

### CLI Flags

```bash
prtool run [flags]

Flags:
      --github-token string      GitHub personal access token
      --github-org string        GitHub organization
      --github-team string       GitHub team  
      --github-user string       GitHub user
      --github-repos strings     GitHub repositories
      --llm-provider string      LLM provider (openai, ollama)
      --llm-model string         LLM model
      --llm-api-key string       LLM API key
      --llm-base-url string      LLM base URL
      --output-format string     Output format (default "markdown")
      --output-file string       Output file
      --since string            Time range for PRs (default "-7d")
      --verbose                 Verbose output
      --dry-run                 Dry run mode
      --ci                      CI mode
      --log-file string         Log file path
      --config string           Config file (default $HOME/.prtool.yaml)
  -h, --help                    Help for run
```

## Usage Examples

### Basic Usage

Fetch PRs from the last 7 days for a specific repository:
```bash
prtool run --github-token $GITHUB_TOKEN --github-repos owner/repo
```

### Organization-wide Search

Fetch PRs from all repositories in an organization:
```bash
prtool run --github-token $GITHUB_TOKEN --github-org myorg --since -14d
```

### Dry Run Mode

Preview what PRs would be processed without calling the LLM:
```bash
prtool run --dry-run --github-repos owner/repo
```

### File Output

Save the summary to a file:
```bash
prtool run --output-file pr-summary.md --github-repos owner/repo
```

### Using Ollama

Use a local Ollama instance for summaries:
```bash
prtool run --llm-provider ollama --llm-model llama2 --github-repos owner/repo
```

### CI Mode

Run in CI with structured output and exit codes:
```bash
prtool run --ci --log-file prtool.log
```

Exit codes in CI mode:
- 0: Success
- 1: General error
- 2: Configuration error (missing token, etc.)
- 3: API error (GitHub API issues)
- 4: LLM error (summary generation failed)

### Verbose Logging

Enable detailed logging to see what's happening:
```bash
prtool run --verbose --log-file debug.log
```

## Time Range Format

The `--since` flag accepts relative time formats:
- `-7d` - Last 7 days
- `-2w` - Last 2 weeks  
- `-1m` - Last 1 month
- `-3mo` - Last 3 months
- `-1yr` - Last 1 year

## Scope Configuration

You must specify exactly one scope:
- `--github-org` - All repositories in an organization
- `--github-team` - Repositories accessible to a team
- `--github-user` - A user's repositories
- `--github-repos` - Specific repositories (comma-separated)

## LLM Providers

### OpenAI
Requires an API key from https://platform.openai.com/api-keys

```bash
prtool run --llm-provider openai --llm-api-key $OPENAI_API_KEY --llm-model gpt-3.5-turbo
```

### Ollama
Requires Ollama running locally (https://ollama.ai)

```bash
prtool run --llm-provider ollama --llm-model llama2
```

### Stub (Testing)
Uses a stub implementation for testing:

```bash
prtool run --llm-provider stub
```

## Version Management

Check current version:
```bash
prtool --version
```

Check for updates:
```bash
prtool --version-check
```

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o prtool .
```

### Linting
```bash
go vet ./...
golangci-lint run
```

## License

[Your License Here]
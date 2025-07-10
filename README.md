# prtool

A command-line tool that fetches GitHub pull requests (PRs) for a specified time period and scope (organization, team, user, or repository), summarizes them using an LLM (OpenAI or Ollama), and outputs the result in Markdown format.

## Features

- **Multi-scope PR fetching**: Fetch PRs from organizations, teams, users, or specific repositories
- **Time-based filtering**: Filter PRs by merge date using relative time ranges (`-7d`, `-1m`, `-1yr`)
- **AI-powered summaries**: Generate intelligent summaries using OpenAI or Ollama
- **Multiple output formats**: Output to stdout or save to files
- **Flexible configuration**: Configure via YAML files, environment variables, or CLI flags
- **CI/CD friendly**: Special mode for automated environments
- **Dry-run support**: Preview data without generating summaries

## Installation

### From Source

```bash
git clone https://github.com/willis7/prtool.git
cd prtool
go build .
```

### Binary Releases

Download the latest binary from the [releases page](https://github.com/willis7/prtool/releases).

## Quick Start

### 1. Generate Configuration File

```bash
prtool init
```

This creates a `.prtool.yaml` file with all available configuration options and comments.

### 2. Configure GitHub Token

Set your GitHub personal access token:

```bash
export PRTOOL_GITHUB_TOKEN="your_github_token_here"
```

Or edit the `.prtool.yaml` file:

```yaml
github_token: "your_github_token_here"
```

### 3. Fetch PRs

```bash
# Fetch PRs from a user for the last 7 days
prtool --user=octocat --since=-7d

# Fetch PRs from an organization for the last month
prtool --org=github --since=-1m

# Fetch PRs from a specific repository
prtool --repo=owner/repository --since=-2w
```

## Usage Examples

### Basic Usage

```bash
# Fetch PRs from user 'octocat' for the last week
prtool --user=octocat --since=-7d

# Fetch PRs from organization 'github' for the last month
prtool --org=github --since=-1m

# Fetch PRs from specific repository
prtool --repo=microsoft/vscode --since=-14d

# Fetch PRs from team (format: org/team)
prtool --team=github/docs --since=-1w
```

### AI Summary Generation

```bash
# Use OpenAI for AI summaries
prtool --user=octocat --llm-provider=openai --llm-api-key=sk-...

# Use Ollama (local)
prtool --user=octocat --llm-provider=ollama --llm-model=llama3.2

# Skip AI summary generation (dry-run)
prtool --user=octocat --dry-run
```

### Output Options

```bash
# Save to file
prtool --user=octocat --output=report.md

# Verbose logging
prtool --user=octocat --verbose

# Log to file
prtool --user=octocat --log-file=prtool.log

# CI-friendly mode (no progress indicators)
prtool --user=octocat --ci
```

### Configuration File

Create a configuration file with `prtool init`, then customize:

```yaml
# .prtool.yaml
github_token: "ghp_xxxxxxxxxxxx"
user: "octocat"
since: "-7d"
llm_provider: "openai"
llm_api_key: "sk-xxxxxxxxxxxx"
llm_model: "gpt-3.5-turbo"
output: "weekly-report.md"
verbose: true
```

Then run simply:

```bash
prtool
```

## Configuration

Configuration can be provided via:

1. **Command-line flags** (highest precedence)
2. **Environment variables**
3. **YAML configuration file** (lowest precedence)

### CLI Flags

| Flag             | Description                       | Example                  |
| ---------------- | --------------------------------- | ------------------------ |
| `--github-token` | GitHub personal access token      | `--github-token=ghp_xxx` |
| `--org`          | GitHub organization               | `--org=github`           |
| `--team`         | GitHub team (org/team)            | `--team=github/docs`     |
| `--user`         | GitHub user                       | `--user=octocat`         |
| `--repo`         | GitHub repository (owner/repo)    | `--repo=owner/repo`      |
| `--since`        | Time range for PRs                | `--since=-7d`            |
| `--llm-provider` | LLM provider (stub/openai/ollama) | `--llm-provider=openai`  |
| `--llm-api-key`  | LLM API key                       | `--llm-api-key=sk-xxx`   |
| `--llm-model`    | LLM model name                    | `--llm-model=gpt-4`      |
| `--output`       | Output file path                  | `--output=report.md`     |
| `--dry-run`      | Skip LLM processing               | `--dry-run`              |
| `--verbose`      | Enable verbose logging            | `--verbose`              |
| `--ci`           | CI-friendly mode                  | `--ci`                   |
| `--log-file`     | Log file path                     | `--log-file=app.log`     |

### Environment Variables

All flags have corresponding environment variables with the `PRTOOL_` prefix:

```bash
export PRTOOL_GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export PRTOOL_USER="octocat"
export PRTOOL_SINCE="-7d"
export PRTOOL_LLM_PROVIDER="openai"
export PRTOOL_LLM_API_KEY="sk-xxxxxxxxxxxx"
export PRTOOL_VERBOSE="true"
```

### Time Range Format

| Format | Description   | Example        |
| ------ | ------------- | -------------- |
| `-7d`  | Last 7 days   | `--since=-7d`  |
| `-2w`  | Last 2 weeks  | `--since=-2w`  |
| `-1m`  | Last 1 month  | `--since=-1m`  |
| `-3mo` | Last 3 months | `--since=-3mo` |
| `-1yr` | Last 1 year   | `--since=-1yr` |

## LLM Providers

### OpenAI

```bash
prtool --llm-provider=openai --llm-api-key=sk-xxx --llm-model=gpt-3.5-turbo
```

Supported models:

- `gpt-3.5-turbo` (default)
- `gpt-4`
- `gpt-4-turbo`

### Ollama (Local)

First, start Ollama locally:

```bash
ollama serve
ollama pull llama3.2
```

Then use with prtool:

```bash
prtool --llm-provider=ollama --llm-model=llama3.2
```

### Stub (Testing)

```bash
prtool --llm-provider=stub
```

Returns a fixed summary for testing purposes.

## CI/CD Usage

For automated environments, use the `--ci` flag:

```bash
prtool --user=octocat --since=-7d --output=report.md --ci
```

This mode:

- Suppresses progress indicators and spinners
- Uses clean exit codes (0 for success, 1 for failure)
- Reduces verbose output for cleaner logs

## Commands

### `prtool` (default)

Fetch and summarize pull requests.

### `prtool init`

Generate a sample configuration file (`.prtool.yaml`) in the current directory.

```bash
prtool init
```

### Version Commands

```bash
# Show current version
prtool --version

# Check for latest version on GitHub
prtool --version-check
```

## Authentication

Create a GitHub Personal Access Token with the following permissions:

- `repo` (for private repositories)
- `public_repo` (for public repositories)
- `read:org` (for organization access)

Set the token via:

1. Environment variable: `export PRTOOL_GITHUB_TOKEN="ghp_xxx"`
2. Configuration file: `github_token: "ghp_xxx"`
3. Command line: `--github-token=ghp_xxx`

## Examples

### Weekly Team Report

```bash
# Generate weekly report for a team
prtool --team=myorg/backend --since=-7d --output=weekly-report.md --llm-provider=openai

# With configuration file
echo "team: myorg/backend
since: -7d
output: weekly-report.md
llm_provider: openai
llm_api_key: sk-xxx" > .prtool.yaml

prtool
```

### Monthly Organization Summary

```bash
prtool --org=mycompany --since=-1m --llm-provider=ollama --verbose
```

### Individual Contributor Report

```bash
prtool --user=developer --since=-2w --dry-run
```

## Troubleshooting

### Common Issues

**Error: GitHub token is required**

- Set your GitHub token via environment variable, config file, or CLI flag

**Error: No pull requests found**

- Check your scope (org/team/user/repo) and time range
- Verify you have access to the specified repositories

**OpenAI API error**

- Verify your API key is correct
- Check you have sufficient credits/quota

**Ollama connection failed**

- Ensure Ollama is running locally (`ollama serve`)
- Verify the model is installed (`ollama pull llama3.2`)

### Debug Mode

Use verbose logging to troubleshoot:

```bash
prtool --user=octocat --verbose --log-file=debug.log
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

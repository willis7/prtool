# prtool Specification

## Overview
`prtool` is a command-line tool written in Go that fetches GitHub pull requests (PRs) for a specified time period and scope (organization, team, user, or repository), summarizes them using an LLM (OpenAI or Ollama), and outputs the result in Markdown format. It supports configuration via CLI flags, environment variables, and a YAML config file.

## Core Features
- Fetch merged PRs from GitHub by org, team, user, or repo (mutually exclusive)
- Filter by time range (e.g. `-7d`, `-1m`, `-1yr`)
- Output summary in Markdown format
- Support for OpenAI and Ollama as LLM providers
- Customizable prompt and output file
- Configurable via CLI, environment variables, and YAML config file
- `--dry-run` and `--verbose` flags for inspection and debugging
- Optional metadata prepended to Markdown output
- Built-in `init`, `help`, and `version` commands
- Manual `--version-check` support

## Command Structure
- `prtool run`: Execute the tool
- `prtool init`: Generate a sample YAML config
- `prtool --help`: Show help
- `prtool --version`: Show version
- `prtool --version-check`: Check for latest version

## Configuration Sources (priority order)
1. CLI flags
2. Environment variables
3. YAML config file (`~/.prtool.yaml` by default)

## Input Fields
### GitHub Scope (mutually exclusive)
- `--org <org>`
- `--team <org/team>`
- `--user <username>`
- `--repo <org/repo>`

### Time Range
- `--since "-7d"` or `--since "-1m"`

### LLM Options
- `--llm-provider [openai|ollama]`
- `--llm-api-key <key>`
- `--llm-model <model>`
- `--prompt <path-to-prompt-file>`

### Output Options
- `--output <file.md>` (Markdown file)
- `--ci`: Non-interactive mode
- `--dry-run`: Skip LLM and output
- `--verbose`: Show logs

## Data Captured per PR
- Title
- Description (Body)
- Status (Merged)
- Author
- Created at
- Merged at
- Labels
- Changed file paths

## LLM Prompt
Default prompt:
> Summarise the following list of GitHub pull requests in a concise, human-readable way for a non-technical audience. Highlight patterns, focus areas, or significant pieces of work. Avoid technical jargon where possible. Include bullet points.

## Output
- Markdown format with:
  - Metadata block (timeframe, scope, PR count, date)
  - Summary body from LLM

## Error Handling
- **Fail Fast**:
  - Invalid/missing GitHub or LLM credentials
  - Invalid scope
- **Graceful**:
  - Retry (max 3 times) on network/API errors
  - Skip failed repositories with warning
- **Dry Run** disables external calls

## Logging & Telemetry
- Default: Silent
- `--verbose`: Console logging
- Config option to write to a file
- No telemetry unless added later

## Architecture
- Written in Go
- Uses Cobra for CLI
- Modular structure for GitHub API, LLM interface, config parsing, and output formatting
- YAML config using `gopkg.in/yaml.v2`
- OpenAI & Ollama via pluggable interface

## Testing Plan
- **Unit Tests**
  - Config loading and precedence logic
  - Time range parsing
  - Markdown rendering
- **Integration Tests**
  - GitHub API (stubbed or fixture-based)
  - Scope resolution logic
  - End-to-end dry-run execution
- **LLM Interface**
  - Stub for OpenAI/Ollama
  - Test error cases and malformed payloads

## Tooling & Conventions
- Versioning via `ldflags`
- Semantic version tags (e.g. `v0.1.0`)
- Go Modules (`go.mod`, `go.sum`)
- No colorized output
- No plugin loading
- English only
- Compatible with CI tools (via `--ci`)

## Dependency Guidelines
- Use well-maintained libraries only
- Must be under MIT, BSD, Apache 2.0 licenses
- Pin versions and optionally audit with `go-licenses`

## Future Enhancements (not in MVP)
- Support for `open` and `closed` PRs
- Combined scope (e.g. user + repo)
- Caching of GitHub API responses
- Grouping summaries by repo/author
- Redaction or filtering of PRs
- Output formats beyond Markdown
- Plugin system for custom handlers
- Automatic version checks


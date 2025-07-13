# prtool

A CLI for summarizing GitHub PRs with LLMs.

## Usage

```sh
prtool --help
```

### Init config

```sh
prtool init
# Generates prtool.config.yaml in current directory
```

### Version check

```sh
prtool --version-check
# Prints latest release
```

### CI mode

```sh
prtool --ci
# Disables interactive output/spinners, sets exit codes
```

### Verbose and logging

```sh
prtool --verbose --log-file=prtool.log
# Enables verbose logging and writes logs to prtool.log
```

## Flags

- `--ci`: CI mode (no spinners, strict exit codes)
- `--verbose`: Verbose logging
- `--log-file`: Path to log file
- `--version`: Print version
- `--version-check`: Check for latest release
- `--output`: Output file for Markdown

## Example

```sh
prtool --scope=myorg --github-token=... --llm-provider=openai --output=prs.md
```

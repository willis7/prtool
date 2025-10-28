# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.3] - 2025-10-28

### Added

- **Multiple Team Support**: Enhanced PR search functionality to support searching across one or more teams simultaneously
  - Accept comma-separated team names via `--team` flag (e.g., `--team org/team1,org/team2`)
  - Support multiple teams in YAML configuration and environment variables
  - Automatically deduplicate repositories when teams share common repos
  - Maintain backward compatibility with single team usage

### Changed

- Updated internal configuration structure to handle team arrays instead of single team strings
- Modified GitHub API client to aggregate repositories from multiple teams
- Enhanced CLI flag descriptions to indicate multiple team support

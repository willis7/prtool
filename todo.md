# prtool TODO Checklist

Mark each task with `[x]` as you complete it.

---

## P0 — Repository & Tooling
- [ ] **S0.1** Initialize Git repository and add `.gitignore` for Go
- [ ] **S0.2** Create `go.mod` (module `github.com/yourorg/prtool`) and minimal `main.go`
- [ ] **S0.3** Add `Makefile` with `test` target and GitHub Actions CI workflow running `make test`

## P1 — CLI Skeleton (Cobra)
- [ ] **S1.1** Add dependency `github.com/spf13/cobra/v4`
- [ ] **S1.2** Generate `cmd/root.go` with global `--version` flag (default `dev`)
- [ ] **S1.3** Wire `main.go` → `cmd.Execute()`
- [ ] **S1.4** Add unit test ensuring `prtool --version` prints version string

## P2 — Configuration System
- [ ] **S2.1** Define `Config` struct in `internal/config/config.go`
- [ ] **S2.2** Implement YAML loader `LoadFromFile(path)`
- [ ] **S2.3** Implement environment loader `LoadFromEnv()`
- [ ] **S2.4** Bind Cobra flags to config fields
- [ ] **S2.5** Implement `MergeConfig(cli, env, yaml)` with precedence CLI > env > YAML
- [ ] **S2.6** Write unit tests covering precedence matrix

## P3 — Time‑Range Parsing
- [ ] **S3.1** Create `ParseRelativeDuration("-7d")` utility
- [ ] **S3.2** Add edge‑case tests (valid/invalid inputs)

## P4 — GitHub Client
- [ ] **S4.1** Define `GitHubClient` interface (`ListRepos`, `ListPRs`)
- [ ] **S4.2** Implement `RestClient` using PAT auth via `go-github`
- [ ] **S4.3** Factory to inject client in `internal/gh`
- [ ] **S4.4** Create `MockClient` for tests
- [ ] **S4.5** Write auth‑failure unit tests

## P5 — Scope Resolution
- [ ] **S5.1** Functions to resolve repos for org, team, user, repo
- [ ] **S5.2** Ensure mutual‑exclusion validation
- [ ] **S5.3** Unit tests with `MockClient`

## P6 — PR Retrieval
- [ ] **S6.1** Define `model.PR` struct
- [ ] **S6.2** Implement `FetchPRs` filtering by `since` and `merged`
- [ ] **S6.3** Print PR table when `--dry-run`
- [ ] **S6.4** Integration tests with `MockClient`

## P7 — Output Layer
- [ ] **S7.1** Create Markdown renderer with metadata header
- [ ] **S7.2** Write output to stdout or file (`--output` flag)
- [ ] **S7.3** Implement `--dry-run` & `--verbose` behaviour
- [ ] **S7.4** Unit tests (golden files)

## P8 — LLM Abstraction
- [ ] **S8.1** Define `LLM` interface
- [ ] **S8.2** Implement `StubLLM` returning canned summary
- [ ] **S8.3** Unit tests for stub and error paths

## P9 — LLM Integration
- [ ] **S9.1** OpenAI provider using `go-openai`
- [ ] **S9.2** Ollama provider via local HTTP API
- [ ] **S9.3** Provider factory based on config
- [ ] **S9.4** End‑to‑end flow: PRs → context → LLM → Markdown
- [ ] **S9.5** Integration test with stubs

## P10 — Ancillary Commands
- [ ] **S10.1** Implement `prtool init` to generate annotated YAML config
- [ ] **S10.2** Add `--version-check` flag querying GitHub releases (mocked tests)

## P11 — Polish & CI Enhancements
- [ ] **S11.1** Add `--ci` flag (non‑interactive mode)
- [ ] **S11.2** Implement logging to file when configured; honor `--verbose`
- [ ] **S11.3** Update `README.md` with examples and autocompletion steps
- [ ] **S11.4** Tag `v0.1.0` release and configure release workflow

---

### Continuous Integration / Quality Gates
- [ ] Run `go test ./...` — all green
- [ ] Run `go vet ./...`
- [ ] Run `golangci-lint run`
- [ ] Ensure `go mod tidy` produces no diff

### Manual Smoke Tests
- [ ] Build binary (`go build ./cmd/prtool`)
- [ ] Execute `prtool run --dry-run --org <org> --since "-7d"` with a valid PAT
- [ ] Execute full run with OpenAI creds and confirm Markdown output

---

_Keep this TODO list updated; it reflects every incremental task needed to deliver the MVP._


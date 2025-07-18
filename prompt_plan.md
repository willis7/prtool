# prtool Development Blueprint

This blueprint maps out a **progressive, test‑driven path** from an empty repository to a fully‑functioning MVP that satisfies the `spec.md` requirements.

The plan is intentionally granular: each step is sized to ship a small, safe increment that **compiles**, **passes tests**, and **moves the project forward** without leaving orphaned code.

---

## 1  High‑Level Phases

| Phase | Goal                                                   |
| ----- | ------------------------------------------------------ |
| P0    | Repository & tooling bootstrap                         |
| P1    | Core CLI skeleton (Cobra) & version flag               |
| P2    | Configuration system (YAML + env + CLI precedence)     |
| P3    | Time‑range parsing utility                             |
| P4    | GitHub client with PAT auth                            |
| P5    | Scope resolution (org / team / user / repo)            |
| P6    | PR retrieval & in‑memory storage                       |
| P7    | Output layer: Markdown writer, dry‑run & verbose flags |
| P8    | LLM interface abstraction & stub                       |
| P9    | LLM integration & summary generation                   |
| P10   | Ancillary commands (`init`, `version‑check`)           |
| P11   | End‑to‑end flows & polishing                           |

Phases P0‑P11 map 1‑to‑1 onto the iterative chunks below.

---

## 2  Iterative Chunks → Tiny Steps

Each **chunk** delivers one phase. Each chunk is decomposed into **atomic steps** that can be completed in <30 min and include tests.

> **Notation**\
> `Sx.y` = Step *y* inside Chunk *x*.

### Chunk 0 (P0) – Repo & Tooling

- **S0.1** `git init` new repo, commit `.gitignore` for Go
- **S0.2** Add `go.mod` (module `github.com/yourorg/prtool`) & blank `main.go` with `package main` + empty `func main()`
- **S0.3** Add `make test` target and CI stub (GitHub Actions YAML) that runs `go test ./...`

### Chunk 1 (P1) – CLI Skeleton

- **S1.1** Add dependency `spf13/cobra` to `go.mod`
- **S1.2** Generate root `cmd/root.go` with global `--version` flag
- **S1.3** Wire `main.go` -> `cmd.Execute()`
- **S1.4** Unit test: invoking `prtool --version` prints placeholder version string

### Chunk 2 (P2) – Configuration

- **S2.1** Define struct `Config` in `internal/config` with all fields (token, scope, llm etc.)
- **S2.2** Implement YAML loader (`LoadFromFile`)
- **S2.3** Implement env override (`LoadFromEnv`)
- **S2.4** Implement CLI binding via Cobra flags (`BindFlags`)
- **S2.5** Function `MergeConfig(cli, env, yaml)` applying precedence (CLI > env > YAML)
- **S2.6** Unit tests covering precedence matrix

### Chunk 3 (P3) – Time Range Parsing

- **S3.1** Utility `ParseRelativeDuration("-7d") (time.Time, error)`
- **S3.2** Edge‑case tests (`-1m`, `-1yr`, bad input)

### Chunk 4 (P4) – GitHub Client

- **S4.1** Interface `GitHubClient` with `ListRepos`, `ListPRs`
- **S4.2** Concrete impl using `google/go‑github` w/ PAT
- **S4.3** Inject client via factory in `internal/gh`
- **S4.4** Stub/mock implementation for tests
- **S4.5** Unit tests validating auth failure handling

### Chunk 5 (P5) – Scope Resolution

- **S5.1** Functions: `ReposForOrg`, `ReposForTeam`, `ReposForUser`, `ReposForRepo`
- **S5.2** Mutually exclusive validation logic
- **S5.3** Tests with stub GitHub client verifying repo lists

### Chunk 6 (P6) – PR Retrieval

- **S6.1** Define `PR` struct capturing required fields
- **S6.2** Implement `FetchPRs(repos, since)` that paginates merged PRs
- **S6.3** Dry‑run print table of PRs (behind flag)
- **S6.4** Integration test using stub GitHub client

### Chunk 7 (P7) – Output Layer

- **S7.1** Markdown renderer `RenderMarkdown(metadata, []PR)`
- **S7.2** Write to stdout or file based on `--output`
- **S7.3** Implement `--dry‑run` (skip LLM) and `--verbose` (log)
- **S7.4** Unit tests for renderer & flag behavior

### Chunk 8 (P8) – LLM Abstraction

- **S8.1** Interface `LLM` with `Summarise(context string) (string, error)`
- **S8.2** Stub implementation returning canned text (for tests)
- **S8.3** Unit tests on error paths

### Chunk 9 (P9) – LLM Integration

- **S9.1** OpenAI implementation using `github.com/sashabaranov/go‑openai`
- **S9.2** Ollama implementation via local HTTP API
- **S9.3** Factory selects provider from config
- **S9.4** End‑to‑end flow: PRs -> context -> LLM -> Markdown -> file/stdout
- **S9.5** Integration test with stub LLM & stub GitHub

### Chunk 10 (P10) – Ancillary Commands

- **S10.1** `prtool init` generates annotated YAML config
- **S10.2** `--version-check` flag hits GitHub releases API (mocked in tests)

### Chunk 11 (P11) – Polish & CI

- **S11.1** Add `--ci` flag: silence progress bars, use exit codes
- **S11.2** Update README with usage examples
- **S11.3** Tag `v0.1.0` release in GitHub Actions workflow

---

## 3  Right‑Sizing Review

All steps:

- Deliver ≤200 lines of code (LoC) each
- Introduce ≤1 new dependency
- Include unit or integration tests that **fail before** code and **pass after**
- Leave repository in a compiling, green‑test state

If any step feels too large during implementation, split it (e.g. write interface in one step, implementation in the next).

---

## 4  Code‑Generation Prompts

Below are **15 prompts** (one per chunk, plus a final wiring prompt). Paste each prompt into your code‑generation LLM **sequentially**; wait for tests to pass before moving on.

---

### Prompt 1 — Bootstrap Repository

```text
You are implementing `prtool`, a Go CLI.
Objective: create an initial repository scaffold that compiles and is testable.
Tasks:
1. Initialise a Go module `github.com/yourorg/prtool`.
2. Add `main.go` with an empty `func main()`.
3. Add `.gitignore` for Go.
4. Add a basic `Makefile` with `test` target.
5. Add GitHub Actions workflow `.github/workflows/ci.yml` that runs `make test`.
6. Add one placeholder test (`internal/placeholder/placeholder_test.go`) that asserts `true`.
All code must compile with `go 1.24`. Run `go test ./...` to ensure green tests.
Output the full file list plus contents.
```

### Prompt 2 — Add Cobra Skeleton

```text
Extend the existing `prtool` repository.
Objective: add Cobra CLI skeleton with a root command and `--version` flag.
Tasks:
1. Add dependency `github.com/spf13/cobra/v4`.
2. Create `cmd/root.go` implementing `Execute()` and `--version` flag that prints `dev` by default.
3. Wire `main.go` to call `cmd.Execute()`.
4. Add unit test `cmd/root_test.go` that executes the command with `--version` and asserts output contains `dev`.
5. Ensure `go test ./...` passes.
Include only new/changed files in your output.
```

### Prompt 3 — Config System

```text
Objective: implement configuration loading with YAML, env vars, and CLI flags.
Tasks:
1. Define `Config` struct in `internal/config/config.go`.
2. Implement `LoadFromFile(path)` using `gopkg.in/yaml.v2`.
3. Implement `LoadFromEnv()` reading env variables (`PRTOOL_GITHUB_TOKEN`, etc.).
4. Bind Cobra flags to config fields.
5. Implement `MergeConfig()` that resolves precedence (CLI > env > YAML).
6. Add tests `internal/config/config_test.go` covering precedence scenarios.
Use table‑driven tests. Ensure all tests pass.
```

### Prompt 4 — Time Range Parser

```text
Objective: add utility to parse relative durations.
Tasks:
1. Create `internal/timeutil/relative.go` with `ParseRelativeDuration(r string) (time.Time, error)` supporting `-7d`, `-1m`, `-1yr`.
2. Edge cases: empty string returns error; positive durations not allowed.
3. Tests `internal/timeutil/relative_test.go` covering valid and invalid inputs.
Run `go test ./...` – all green.
```

### Prompt 5 — GitHub Client Interface

```text
Objective: introduce GitHub client abstraction.
Tasks:
1. Define interface `GitHubClient` in `internal/gh/client.go` with `ListRepos(scope Config) ([]*github.Repository, error)` and `ListPRs(repo string, since time.Time) ([]PR, error)`.
2. Create concrete `RestClient` using `github.com/google/go-github/v55/github` and PAT auth.
3. Add stub `MockClient` in `internal/gh/mock.go` for tests.
4. Tests `internal/gh/client_test.go` verifying auth failure returns helpful error using `MockClient`.
```

### Prompt 6 — Scope Resolution Logic

```text
Objective: implement mutually exclusive scope resolution.
Tasks:
1. Add `internal/scope/scope.go` with functions `ResolveRepos(cfg Config, gh GitHubClient) ([]string, error)`.
2. Validate exactly one of org/team/user/repo is set; else return error.
3. Use `MockClient` to unit test all paths.
Ensure tests pass.
```

### Prompt 7 — PR Fetch & In‑Memory Store

```text
Objective: fetch merged PRs and store in memory.
Tasks:
1. Define `internal/model/pr.go` with struct capturing required fields.
2. Implement `internal/service/fetcher.go` with `Fetch(cfg Config, gh GitHubClient) ([]model.PR, error)` applying since filter and merged state.
3. Add integration test with `MockClient` containing fake PRs; assert correct filtering.
```

### Prompt 8 — Output Layer & Flags

```text
Objective: render PR list to Markdown with metadata, stdout/file, and dry‑run/verbose flags.
Tasks:
1. Create `internal/render/markdown.go` with `Render(meta Metadata, prs []model.PR) string`.
2. Update root command: if `--dry-run` print table of PRs; else call renderer.
3. Add `--output` flag; write file if set.
4. Add unit tests for renderer output (golden files).
```

### Prompt 9 — LLM Interface Stub

```text
Objective: add pluggable LLM interface with stub implementation.
Tasks:
1. Define interface `LLM` in `internal/llm/llm.go`.
2. Implement `StubLLM` that returns fixed summary.
3. Inject into service layer; when dry‑run skip, otherwise call `LLM.Summarise()`.
4. Unit tests verify stub is invoked and output inserted into Markdown.
```

### Prompt 10 — OpenAI & Ollama Implementations

```text
Objective: integrate real providers behind interface.
Tasks:
1. Add OpenAI client using `go-openai`.
2. Add Ollama client hitting `http://localhost:11434/api/generate`.
3. Switch via `cfg.LLMProvider`.
4. Tests use stub; add one smoke test (skipped by default) for OpenAI with env var `OPENAI_KEY`.
```

### Prompt 11 — `init` and `version-check`

```text
Objective: add ancillary commands.
Tasks:
1. Add `cmd/init.go` generating annotated YAML into current dir.
2. Add `--version-check` flag (root cmd) that queries latest GitHub release (mock in tests).
3. Tests covering file generation and mocked release check.
```

### Prompt 12 — CI Flag & Polishing

```text
Objective: final polish.
Tasks:
1. Implement `--ci` flag that disables interactive output/spinners and sets exit codes.
2. Ensure all commands honor `cfg.LogFile` and `--verbose`.
3. Update README with usage samples.
4. All tests green, `go vet`, `golangci-lint` (add to CI).
```

### Prompt 13 — End‑to‑End Integration Test

```text
Objective: wire everything together.
Tasks:
1. Add test `internal/e2e/e2e_test.go` that spins MockGitHub + StubLLM, runs `prtool run` via `os/exec`, asserts Markdown output contains stub summary.
2. Ensure binary builds with `go build ./...`.
3. Tag `v0.1.0` in Makefile (but do not create release).
```

### Prompt 14 — Release Artifact

```text
Objective: produce release build workflow.
Tasks:
1. Extend GitHub Actions to build multi‑OS binaries on tag and upload as release assets.
2. Ensure version injected via `ldflags`.
3. Add integration test verifying `--version` reflects ldflag value.
```

### Prompt 15 — Final Wiring & Cleanup

```text
Objective: verify no orphaned code.
Tasks:
1. Run `go mod tidy`.
2. Ensure 100% test pass.
3. Confirm `prtool run --dry-run` works against live GitHub when token present (manual smoke test step).
4. Push changes and open PR.
```

---

> **Implementation protocol:** After each prompt, run tests locally (or via CI). Only proceed when green. If a step grows too large, split it and iterate.


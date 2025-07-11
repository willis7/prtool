# Testing prtool

## Prerequisites

1. GitHub Personal Access Token with the following permissions:
   - `repo` (full)
   - `read:org`
   - `read:user`

2. Set your token:
   ```bash
   export GITHUB_TOKEN=your_github_pat_here
   ```

## Quick Smoke Test

Run the automated smoke test:
```bash
./smoke-test.sh
```

## Manual Testing

### 1. Basic Dry Run Test
Test against a public repository without calling LLM:
```bash
./prtool run --dry-run \
  --github-token $GITHUB_TOKEN \
  --github-repos "golang/go" \
  --since "-7d"
```

Expected output:
- Table showing recent merged PRs from golang/go
- No LLM summaries (dry-run mode)

### 2. Test with Stub LLM
Test the full flow with stub LLM:
```bash
./prtool run \
  --github-token $GITHUB_TOKEN \
  --github-repos "golang/go" \
  --since "-7d" \
  --llm-provider stub
```

Expected output:
- Markdown report with PR summaries
- Each summary should contain "introduces significant changes"

### 3. Test Multiple Repositories
```bash
./prtool run --dry-run \
  --github-token $GITHUB_TOKEN \
  --github-repos "golang/go,kubernetes/kubernetes" \
  --since "-3d"
```

### 4. Test Organization Scope
```bash
./prtool run --dry-run \
  --github-token $GITHUB_TOKEN \
  --github-org "golang" \
  --since "-1d"
```

### 5. Test CI Mode
```bash
./prtool run --ci --dry-run \
  --github-token $GITHUB_TOKEN \
  --github-repos "golang/go" \
  --since "-7d" \
  --log-file test.log
```

Expected:
- Structured output with [INFO] prefixes
- Log file created with detailed logs

### 6. Test with Real LLM (Optional)

#### OpenAI
```bash
export OPENAI_API_KEY=your_openai_key
./prtool run \
  --github-token $GITHUB_TOKEN \
  --github-repos "golang/go" \
  --since "-7d" \
  --llm-provider openai \
  --llm-api-key $OPENAI_API_KEY \
  --llm-model gpt-3.5-turbo
```

#### Ollama (requires local Ollama running)
```bash
./prtool run \
  --github-token $GITHUB_TOKEN \
  --github-repos "golang/go" \
  --since "-7d" \
  --llm-provider ollama \
  --llm-model llama2
```

## Troubleshooting

### Authentication Error
If you see "authentication failed", verify:
1. Your token is valid
2. Token has required permissions
3. Token is correctly set in environment

### No PRs Found
- Try a longer time range (e.g., `-14d` or `-30d`)
- Try a more active repository
- Verify the repository name is correct (owner/repo format)

### Rate Limiting
GitHub API has rate limits. If you hit them:
- Wait a few minutes
- Use a smaller time range
- Test with fewer repositories
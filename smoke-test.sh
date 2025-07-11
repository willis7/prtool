#!/bin/bash
# Smoke test script for prtool

echo "=== prtool Smoke Test ==="
echo "This script tests prtool against live GitHub API"
echo ""

# Check for GitHub token
if [ -z "$GITHUB_TOKEN" ]; then
    echo "ERROR: GITHUB_TOKEN environment variable not set"
    echo "Please set your GitHub token:"
    echo "  export GITHUB_TOKEN=your_github_pat"
    exit 1
fi

# Build the tool
echo "Building prtool..."
make build || exit 1

echo ""
echo "Running smoke tests..."
echo ""

# Test 1: Version check
echo "1. Testing version flag..."
./prtool --version
echo ""

# Test 2: Help
echo "2. Testing help..."
./prtool --help | head -5
echo ""

# Test 3: Dry run with a popular repo
echo "3. Testing dry-run mode with kubernetes/kubernetes (last 7 days)..."
./prtool run --dry-run \
    --github-token "$GITHUB_TOKEN" \
    --github-repos "kubernetes/kubernetes" \
    --since "-7d"

echo ""
echo "4. Testing dry-run mode with golang/go (last 3 days)..."
./prtool run --dry-run \
    --github-token "$GITHUB_TOKEN" \
    --github-repos "golang/go" \
    --since "-3d"

echo ""
echo "5. Testing init command..."
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"
"$OLDPWD/prtool" init
if [ -f .prtool.yaml ]; then
    echo "✓ Config file created successfully"
    head -10 .prtool.yaml
else
    echo "✗ Config file creation failed"
fi
cd "$OLDPWD"
rm -rf "$TEMP_DIR"

echo ""
echo "=== Smoke Test Complete ==="
echo ""
echo "If all tests passed, you can test with real LLM by running:"
echo "  ./prtool run --github-token \$GITHUB_TOKEN --github-repos owner/repo --llm-provider stub"
echo ""
echo "For OpenAI (requires OPENAI_API_KEY):"
echo "  ./prtool run --github-token \$GITHUB_TOKEN --github-repos owner/repo --llm-provider openai --llm-api-key \$OPENAI_API_KEY"
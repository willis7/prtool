package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockPR represents a pull request for testing
type mockPR struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	MergedAt  time.Time `json:"merged_at"`
	User      mockUser  `json:"user"`
	HTMLURL   string    `json:"html_url"`
	Labels    []mockLabel `json:"labels"`
}

type mockUser struct {
	Login string `json:"login"`
}

type mockLabel struct {
	Name string `json:"name"`
}

type mockRepo struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	Owner       mockUser `json:"owner"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
}

func TestE2E_PRTool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "prtool")
	projectRoot := filepath.Join("..", "..")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build prtool: %v\nOutput: %s", err, output)
	}

	// Create mock GitHub server
	mockGitHub := createMockGitHubServer(t)
	defer mockGitHub.Close()

	// Create test config file
	configPath := filepath.Join(t.TempDir(), ".prtool.yaml")
	configContent := fmt.Sprintf(`
github:
  token: "test-token"
  repositories:
    - "testorg/testrepo"

llm:
  provider: "stub"

output:
  format: "markdown"
`)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Run prtool
	cmd := exec.Command(binPath, "run", "--config", configPath)
	
	// Set environment to use mock server
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GITHUB_API_URL=%s", mockGitHub.URL),
		"PRTOOL_GITHUB_TOKEN=test-token",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("prtool run failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify output - note: some output may go to stderr (progress messages)
	output := stdout.String()
	stderrOutput := stderr.String()
	
	// Debug: print what we got
	t.Logf("STDERR (%d bytes):\n%s", len(stderrOutput), stderrOutput)
	
	// The output may contain info messages before the markdown
	// Find where the markdown starts
	markdownStart := strings.Index(output, "# Pull Request Summary Report")
	if markdownStart == -1 {
		t.Fatalf("Output doesn't contain markdown report:\n%s", output)
	}
	
	// Extract just the markdown portion
	markdownOutput := output[markdownStart:]
	
	// Check that it contains the expected structure
	expectedElements := []string{
		"# Pull Request Summary Report",
		"Generated at",
		"Time range",
		"Total PRs",
		"### testorg/testrepo",
		"Test PR for e2e",
		"@testuser",
		"introduces significant changes", // Stub LLM summary
	}

	for _, element := range expectedElements {
		if !strings.Contains(markdownOutput, element) {
			t.Errorf("Output missing expected element: %q", element)
		}
	}

	// Verify it's valid markdown structure
	if !strings.HasPrefix(markdownOutput, "# Pull Request Summary Report") {
		t.Error("Markdown doesn't start with expected header")
	}

	// Check that PR details are included
	if !strings.Contains(output, "#123") {
		t.Error("Output doesn't contain PR number")
	}
}

func TestE2E_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "prtool")
	projectRoot := filepath.Join("..", "..")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build prtool: %v\nOutput: %s", err, output)
	}

	// Create mock GitHub server
	mockGitHub := createMockGitHubServer(t)
	defer mockGitHub.Close()

	// Run prtool in dry-run mode
	cmd := exec.Command(binPath, "run",
		"--dry-run",
		"--github-token", "test-token",
		"--github-repos", "testorg/testrepo",
	)
	
	// Set environment to use mock server
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GITHUB_API_URL=%s", mockGitHub.URL),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("prtool run --dry-run failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify dry-run output
	output := stdout.String()
	
	// Check for table format
	expectedElements := []string{
		"Repository",
		"PR #",  // Note the space
		"Title",
		"Author",
		"testorg/testrepo",
		"123",
		"Test PR for e2e",
		"testuser",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Dry-run output missing expected element: %q\nFull output:\n%s", element, output)
		}
	}

	// Should NOT contain stub summaries in dry-run
	if strings.Contains(output, "introduces significant changes") {
		t.Error("Dry-run output should not contain LLM summaries")
	}
}

func TestE2E_CIMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "prtool")
	projectRoot := filepath.Join("..", "..")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build prtool: %v\nOutput: %s", err, output)
	}

	// Create mock GitHub server
	mockGitHub := createMockGitHubServer(t)
	defer mockGitHub.Close()

	// Create log file path
	logFile := filepath.Join(t.TempDir(), "prtool.log")

	// Run prtool in CI mode
	cmd := exec.Command(binPath, "run",
		"--ci",
		"--log-file", logFile,
		"--github-token", "test-token",
		"--github-repos", "testorg/testrepo",
		"--llm-provider", "stub",
	)
	
	// Set environment to use mock server
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GITHUB_API_URL=%s", mockGitHub.URL),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("prtool run --ci failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Check stderr for CI-formatted messages
	stderrOutput := stderr.String()
	if stderrOutput != "" && !strings.Contains(stderrOutput, "[INFO]") && !strings.Contains(stderrOutput, "[VERBOSE]") {
		t.Errorf("CI mode stderr should contain structured log messages, got: %s", stderrOutput)
	}

	// Check that log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Verify log file content
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(logContent) == 0 {
		t.Error("Log file is empty")
	}
}

func createMockGitHubServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	// Mock repository endpoint
	mux.HandleFunc("/repos/testorg/testrepo", func(w http.ResponseWriter, r *http.Request) {
		repo := mockRepo{
			FullName:    "testorg/testrepo",
			Name:        "testrepo",
			Owner:       mockUser{Login: "testorg"},
			Description: "Test repository",
			HTMLURL:     "https://github.com/testorg/testrepo",
		}
		json.NewEncoder(w).Encode(repo)
	})

	// Mock pulls endpoint
	mux.HandleFunc("/repos/testorg/testrepo/pulls", func(w http.ResponseWriter, r *http.Request) {
		// Check state parameter
		state := r.URL.Query().Get("state")
		if state != "closed" {
			http.Error(w, "Expected state=closed", http.StatusBadRequest)
			return
		}

		prs := []mockPR{
			{
				Number:   123,
				Title:    "Test PR for e2e",
				Body:     "This is a test PR for end-to-end testing",
				State:    "closed",
				MergedAt: time.Now().Add(-24 * time.Hour),
				User:     mockUser{Login: "testuser"},
				HTMLURL:  "https://github.com/testorg/testrepo/pull/123",
				Labels: []mockLabel{
					{Name: "enhancement"},
					{Name: "test"},
				},
			},
			{
				Number:   122,
				Title:    "Another test PR",
				Body:     "This is another test PR",
				State:    "closed",
				MergedAt: time.Now().Add(-48 * time.Hour),
				User:     mockUser{Login: "anotheruser"},
				HTMLURL:  "https://github.com/testorg/testrepo/pull/122",
				Labels:   []mockLabel{},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prs)
	})

	// Mock rate limit endpoint
	mux.HandleFunc("/rate_limit", func(w http.ResponseWriter, r *http.Request) {
		rateLimit := map[string]interface{}{
			"rate": map[string]interface{}{
				"limit":     5000,
				"remaining": 4999,
				"reset":     time.Now().Add(time.Hour).Unix(),
			},
		}
		json.NewEncoder(w).Encode(rateLimit)
	})

	// Default handler for unmatched routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Unhandled request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	return httptest.NewServer(mux)
}
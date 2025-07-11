package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// Checker checks for new versions
type Checker interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*GitHubRelease, error)
}

// GitHubChecker checks versions using GitHub API
type GitHubChecker struct {
	client *http.Client
}

// NewGitHubChecker creates a new GitHub version checker
func NewGitHubChecker() *GitHubChecker {
	return &GitHubChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLatestRelease fetches the latest release from GitHub
func (c *GitHubChecker) GetLatestRelease(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// GitHub recommends including User-Agent
	req.Header.Set("User-Agent", "prtool")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}
	
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &release, nil
}

// CompareVersions compares two version strings
// Returns:
//   -1 if v1 < v2
//    0 if v1 == v2
//    1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	
	// Simple string comparison for now
	// In production, you'd want semantic version comparison
	if v1 < v2 {
		return -1
	} else if v1 > v2 {
		return 1
	}
	return 0
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate(ctx context.Context, checker Checker, currentVersion, owner, repo string) (bool, *GitHubRelease, error) {
	latest, err := checker.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return false, nil, err
	}
	
	comparison := CompareVersions(currentVersion, latest.TagName)
	hasUpdate := comparison < 0
	
	return hasUpdate, latest, nil
}
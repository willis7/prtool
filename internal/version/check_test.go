package version

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockChecker implements the Checker interface for testing
type MockChecker struct {
	Release *GitHubRelease
	Err     error
}

func (m *MockChecker) GetLatestRelease(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Release, nil
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"v1.0.0", "v2.0.0", -1},
		{"v2.0.0", "v1.0.0", 1},
		{"1.0.0", "v2.0.0", -1},
		{"v2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"dev", "v1.0.0", 1}, // dev > 1.0.0 alphabetically
		{"0.9.0", "1.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got := CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestCheckForUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		mockErr        error
		wantUpdate     bool
		wantErr        bool
	}{
		{
			name:           "update available",
			currentVersion: "1.0.0",
			latestVersion:  "2.0.0",
			wantUpdate:     true,
			wantErr:        false,
		},
		{
			name:           "already on latest",
			currentVersion: "2.0.0",
			latestVersion:  "2.0.0",
			wantUpdate:     false,
			wantErr:        false,
		},
		{
			name:           "current is newer",
			currentVersion: "3.0.0",
			latestVersion:  "2.0.0",
			wantUpdate:     false,
			wantErr:        false,
		},
		{
			name:           "error checking",
			currentVersion: "1.0.0",
			mockErr:        errors.New("API error"),
			wantUpdate:     false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockChecker{
				Release: &GitHubRelease{
					TagName:     tt.latestVersion,
					Name:        "Release " + tt.latestVersion,
					PublishedAt: time.Now(),
					HTMLURL:     "https://github.com/owner/repo/releases/tag/" + tt.latestVersion,
				},
				Err: tt.mockErr,
			}

			hasUpdate, release, err := CheckForUpdate(context.Background(), mock, tt.currentVersion, "owner", "repo")

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckForUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if hasUpdate != tt.wantUpdate {
				t.Errorf("CheckForUpdate() hasUpdate = %v, want %v", hasUpdate, tt.wantUpdate)
			}

			if !tt.wantErr && release == nil {
				t.Error("CheckForUpdate() returned nil release without error")
			}
		})
	}
}

func TestGitHubChecker_GetLatestRelease(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		wantErr        bool
		errContains    string
		wantTagName    string
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v1.2.3",
				"name": "Release v1.2.3",
				"published_at": "2024-01-15T10:00:00Z",
				"html_url": "https://github.com/owner/repo/releases/tag/v1.2.3"
			}`,
			wantErr:     false,
			wantTagName: "v1.2.3",
		},
		{
			name:         "no releases found",
			statusCode:   http.StatusNotFound,
			responseBody: `{"message": "Not Found"}`,
			wantErr:      true,
			errContains:  "no releases found",
		},
		{
			name:         "API error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"message": "Internal Server Error"}`,
			wantErr:      true,
			errContains:  "GitHub API error: status 500",
		},
		{
			name:         "invalid JSON",
			statusCode:   http.StatusOK,
			responseBody: `{invalid json`,
			wantErr:      true,
			errContains:  "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "GET" {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/repos/owner/repo/releases/latest" {
					t.Errorf("expected path /repos/owner/repo/releases/latest, got %s", r.URL.Path)
				}
				if r.Header.Get("User-Agent") != "prtool" {
					t.Errorf("expected User-Agent prtool, got %s", r.Header.Get("User-Agent"))
				}
				if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
					t.Errorf("expected Accept header, got %s", r.Header.Get("Accept"))
				}

				// Send response
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create checker with custom transport
			checker := &GitHubChecker{
				client: &http.Client{
					Timeout: 1 * time.Second,
					Transport: &testTransport{
						testServer: server,
					},
				},
			}

			ctx := context.Background()

			release, err := checker.GetLatestRelease(ctx, "owner", "repo")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("GetLatestRelease() error = %v, want error containing %q", err, tt.errContains)
			}

			if !tt.wantErr && release != nil {
				if release.TagName != tt.wantTagName {
					t.Errorf("GetLatestRelease() TagName = %q, want %q", release.TagName, tt.wantTagName)
				}
			}
		})
	}
}

// testTransport redirects requests to our test server
type testTransport struct {
	testServer *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create new request to test server
	testReq, err := http.NewRequest(req.Method, t.testServer.URL+req.URL.Path, req.Body)
	if err != nil {
		return nil, err
	}
	
	// Copy headers
	testReq.Header = req.Header
	
	// Use default transport to make the request
	return http.DefaultTransport.RoundTrip(testReq)
}

func TestNewGitHubChecker(t *testing.T) {
	checker := NewGitHubChecker()
	
	if checker == nil {
		t.Fatal("NewGitHubChecker() returned nil")
	}
	
	if checker.client == nil {
		t.Error("NewGitHubChecker() client is nil")
	}
	
	if checker.client.Timeout != 10*time.Second {
		t.Errorf("NewGitHubChecker() client timeout = %v, want %v", checker.client.Timeout, 10*time.Second)
	}
}

// TestGitHubRelease_JSON verifies JSON marshaling/unmarshaling
func TestGitHubRelease_JSON(t *testing.T) {
	release := &GitHubRelease{
		TagName:     "v1.0.0",
		Name:        "Release 1.0.0",
		PublishedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		HTMLURL:     "https://github.com/owner/repo/releases/tag/v1.0.0",
	}
	
	// Marshal to JSON
	data, err := json.Marshal(release)
	if err != nil {
		t.Fatalf("Failed to marshal release: %v", err)
	}
	
	// Unmarshal back
	var decoded GitHubRelease
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal release: %v", err)
	}
	
	// Verify fields
	if decoded.TagName != release.TagName {
		t.Errorf("TagName = %q, want %q", decoded.TagName, release.TagName)
	}
	if decoded.Name != release.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, release.Name)
	}
	if !decoded.PublishedAt.Equal(release.PublishedAt) {
		t.Errorf("PublishedAt = %v, want %v", decoded.PublishedAt, release.PublishedAt)
	}
	if decoded.HTMLURL != release.HTMLURL {
		t.Errorf("HTMLURL = %q, want %q", decoded.HTMLURL, release.HTMLURL)
	}
}
package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prtool/internal/model"
)

func TestNewOllamaLLM(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   struct {
			model   string
			baseURL string
			temp    float64
		}
	}{
		{
			name:   "default values",
			config: Config{Provider: "ollama"},
			want: struct {
				model   string
				baseURL string
				temp    float64
			}{
				model:   "llama2",
				baseURL: "http://localhost:11434",
				temp:    0.7,
			},
		},
		{
			name: "custom values",
			config: Config{
				Provider:    "ollama",
				Model:       "codellama",
				BaseURL:     "http://custom:8080",
				Temperature: 0.5,
			},
			want: struct {
				model   string
				baseURL string
				temp    float64
			}{
				model:   "codellama",
				baseURL: "http://custom:8080",
				temp:    0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := NewOllamaLLM(tt.config)
			if err != nil {
				t.Fatalf("NewOllamaLLM() error = %v", err)
			}

			if llm.model != tt.want.model {
				t.Errorf("model = %v, want %v", llm.model, tt.want.model)
			}
			if llm.baseURL != tt.want.baseURL {
				t.Errorf("baseURL = %v, want %v", llm.baseURL, tt.want.baseURL)
			}
			if llm.temperature != tt.want.temp {
				t.Errorf("temperature = %v, want %v", llm.temperature, tt.want.temp)
			}
		})
	}
}

func TestOllamaLLM_Summarize_MockServer(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path /api/generate, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Return mock response
		response := `{
			"model": "llama2",
			"response": "This PR implements a new authentication system with improved security features and better user experience.",
			"done": true,
			"created_at": "2024-01-15T10:30:00Z"
		}`
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := Config{
		Provider: "ollama",
		Model:    "llama2",
		BaseURL:  server.URL,
	}

	llm, err := NewOllamaLLM(config)
	if err != nil {
		t.Fatalf("Failed to create Ollama LLM: %v", err)
	}

	ctx := context.Background()
	pr := model.PR{
		Repository:  "test/repo",
		Number:      123,
		Title:       "Add authentication",
		Description: "Implements OAuth2",
		Author:      "alice",
		URL:         "https://github.com/test/repo/pull/123",
	}

	summary, err := llm.Summarize(ctx, pr)
	if err != nil {
		t.Fatalf("Summarize() error: %v", err)
	}

	expected := "This PR implements a new authentication system with improved security features and better user experience."
	if summary != expected {
		t.Errorf("Summarize() = %q, want %q", summary, expected)
	}
}

func TestOllamaLLM_Summarize_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server error"))
			},
			wantErr:     true,
			errContains: "Ollama API error (status 500)",
		},
		{
			name: "invalid JSON response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			wantErr:     true,
			errContains: "failed to decode response",
		},
		{
			name: "connection refused",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Server will be closed before request
			},
			wantErr:     true,
			errContains: "failed to send request to Ollama",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.name != "connection refused" {
				server = httptest.NewServer(http.HandlerFunc(tt.serverResponse))
				defer server.Close()
			} else {
				// Create and immediately close server
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				server.Close()
			}

			config := Config{
				Provider: "ollama",
				BaseURL:  server.URL,
			}

			llm, _ := NewOllamaLLM(config)
			ctx := context.Background()
			
			pr := model.PR{
				Title:  "Test PR",
				Author: "test",
			}

			_, err := llm.Summarize(ctx, pr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Summarize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Summarize() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestOllamaLLM_CheckHealth(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
	}{
		{
			name: "healthy",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/tags" {
					t.Errorf("Expected path /api/tags, got %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"models":[]}`))
			},
			wantErr: false,
		},
		{
			name: "unhealthy",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			llm := &OllamaLLM{
				baseURL:    server.URL,
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			err := llm.CheckHealth(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckHealth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildOllamaPrompt(t *testing.T) {
	pr := model.PR{
		Repository:  "owner/repo",
		Number:      123,
		Title:       "Add new feature",
		Description: "This PR adds feature X",
		Author:      "alice",
		Labels:      []string{"enhancement", "feature"},
	}

	prompt := buildOllamaPrompt(pr)

	// Check that prompt contains key information
	expectedParts := []string{
		"Pull Request: Add new feature",
		"Author: @alice",
		"Repository: owner/repo",
		"Description:",
		"This PR adds feature X",
		"Labels: enhancement, feature",
		"2-3 sentence summary",
	}

	for _, part := range expectedParts {
		if !strings.Contains(prompt, part) {
			t.Errorf("Prompt missing expected part: %q", part)
		}
	}
}

// TestOllamaLLM_Summarize_Integration is an integration test that requires Ollama to be running
func TestOllamaLLM_Summarize_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Ollama integration test in short mode")
	}

	// Check if Ollama is running
	config := Config{
		Provider: "ollama",
		Model:    "llama2",
	}

	llm, err := NewOllamaLLM(config)
	if err != nil {
		t.Fatalf("Failed to create Ollama LLM: %v", err)
	}

	ollama := llm
	
	// Check health first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := ollama.CheckHealth(ctx); err != nil {
		t.Skipf("Skipping Ollama integration test: %v", err)
	}

	// Test summarization
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pr := model.PR{
		Repository:  "test/repo",
		Number:      123,
		Title:       "Implement caching layer",
		Description: "This PR adds a Redis-based caching layer to improve API response times.",
		Author:      "testdev",
		URL:         "https://github.com/test/repo/pull/123",
		Labels:      []string{"performance", "backend"},
	}

	summary, err := llm.Summarize(ctx, pr)
	if err != nil {
		// If the model is not found, skip the test
		if strings.Contains(err.Error(), "model") && strings.Contains(err.Error(), "not found") {
			t.Skipf("Skipping test: Ollama model not available: %v", err)
		}
		t.Fatalf("Summarize() error: %v", err)
	}

	if summary == "" {
		t.Error("Summarize() returned empty summary")
	}

	t.Logf("Generated summary: %s", summary)
}
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yourorg/prtool/internal/model"
)

// OllamaLLM implements the LLM interface using Ollama's local API
type OllamaLLM struct {
	baseURL     string
	model       string
	temperature float64
	httpClient  *http.Client
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	Temperature float64                `json:"temperature,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Stream      bool                   `json:"stream"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Model     string    `json:"model"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// NewOllamaLLM creates a new Ollama LLM instance
func NewOllamaLLM(config Config) (*OllamaLLM, error) {
	if config.Model == "" {
		config.Model = "llama2"
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	temperature := config.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	return &OllamaLLM{
		baseURL:     baseURL,
		model:       config.Model,
		temperature: temperature,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Summarize generates a summary for a single pull request
func (o *OllamaLLM) Summarize(ctx context.Context, pr model.PR) (string, error) {
	prompt := buildOllamaPrompt(pr)

	reqBody := OllamaRequest{
		Model:       o.model,
		Prompt:      prompt,
		Temperature: o.temperature,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// SummarizeBatch generates summaries for multiple pull requests
func (o *OllamaLLM) SummarizeBatch(ctx context.Context, prs []model.PR) (map[string]string, error) {
	summaries := make(map[string]string)

	// Process sequentially to avoid overwhelming local Ollama instance
	for _, pr := range prs {
		summary, err := o.Summarize(ctx, pr)
		if err != nil {
			// Log error but continue with other PRs
			summaries[pr.URL] = fmt.Sprintf("Error generating summary: %v", err)
			continue
		}
		summaries[pr.URL] = summary
	}

	return summaries, nil
}

// Name returns the name of the Ollama provider
func (o *OllamaLLM) Name() string {
	return fmt.Sprintf("ollama (%s)", o.model)
}

// buildOllamaPrompt creates a prompt for summarizing a pull request
func buildOllamaPrompt(pr model.PR) string {
	var sb strings.Builder

	sb.WriteString("You are a helpful assistant that summarizes pull requests. ")
	sb.WriteString("Provide a concise, informative summary focusing on the key changes and their impact.\n\n")
	
	sb.WriteString(fmt.Sprintf("Pull Request: %s\n", pr.Title))
	sb.WriteString(fmt.Sprintf("Author: @%s\n", pr.Author))
	sb.WriteString(fmt.Sprintf("Repository: %s\n", pr.Repository))
	
	if pr.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", pr.Description))
	}
	
	if len(pr.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("\nLabels: %s\n", strings.Join(pr.Labels, ", ")))
	}

	sb.WriteString("\nProvide a 2-3 sentence summary that captures the essence of this pull request, its purpose, and key changes.")

	return sb.String()
}

// CheckHealth verifies that Ollama is running and accessible
func (o *OllamaLLM) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama is not accessible at %s: %w", o.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama health check failed with status %d", resp.StatusCode)
	}

	return nil
}
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaClient implements the LLM interface for Ollama.
type OllamaClient struct {
	BaseURL string
	Model   string
}

// NewOllamaClient creates a new OllamaClient.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
	}
}

// Summarise sends a request to Ollama to summarise the given context.
func (c *OllamaClient) Summarise(ctx context.Context, context string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", c.BaseURL)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  c.Model,
		"prompt": context,
		"stream": false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var responseBody struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	return responseBody.Response, nil
}

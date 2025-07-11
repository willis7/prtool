package llm

import (
	"context"
	"fmt"

	"github.com/yourorg/prtool/internal/model"
)

// LLM defines the interface for Large Language Model providers
type LLM interface {
	// Summarize generates a summary for a single pull request
	Summarize(ctx context.Context, pr model.PR) (string, error)
	
	// SummarizeBatch generates summaries for multiple pull requests
	// Returns a map of PR URL to summary
	SummarizeBatch(ctx context.Context, prs []model.PR) (map[string]string, error)
	
	// Name returns the name of the LLM provider
	Name() string
}

// Config contains LLM configuration
type Config struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	Temperature float64
	MaxTokens   int
}

// NewLLM creates an LLM instance based on the provider
func NewLLM(config Config) (LLM, error) {
	switch config.Provider {
	case "stub", "":
		return NewStubLLM(), nil
	case "openai":
		return NewOpenAILLM(config)
	case "ollama":
		return NewOllamaLLM(config)
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", config.Provider)
	}
}
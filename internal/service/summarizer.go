package service

import (
	"context"
	"fmt"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/llm"
	"github.com/yourorg/prtool/internal/model"
)

// Summarizer handles PR summarization using LLM
type Summarizer struct {
	llm llm.LLM
}

// NewSummarizer creates a new summarizer with the configured LLM
func NewSummarizer(cfg *config.Config) (*Summarizer, error) {
	llmConfig := llm.Config{
		Provider:    cfg.LLM.Provider,
		Model:       cfg.LLM.Model,
		APIKey:      cfg.LLM.APIKey,
		BaseURL:     cfg.LLM.BaseURL,
		Temperature: cfg.LLM.Temperature,
		MaxTokens:   cfg.LLM.MaxTokens,
	}

	llmInstance, err := llm.NewLLM(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	return &Summarizer{
		llm: llmInstance,
	}, nil
}

// SummarizePRs adds AI-generated summaries to pull requests
func (s *Summarizer) SummarizePRs(ctx context.Context, prs []model.PR, verbose bool) ([]model.PR, error) {
	if len(prs) == 0 {
		return prs, nil
	}

	if verbose {
		fmt.Printf("Generating summaries for %d pull requests using %s...\n", len(prs), s.llm.Name())
	}

	// Use batch summarization for efficiency
	summaries, err := s.llm.SummarizeBatch(ctx, prs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summaries: %w", err)
	}

	// Update PRs with summaries
	for i := range prs {
		if summary, ok := summaries[prs[i].URL]; ok {
			prs[i].Summary = summary
		}
	}

	if verbose {
		fmt.Printf("Successfully generated %d summaries\n", len(summaries))
	}

	return prs, nil
}
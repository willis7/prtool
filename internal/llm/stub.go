package llm

import (
	"context"
	"fmt"

	"github.com/yourorg/prtool/internal/model"
)

// StubLLM is a stub implementation of the LLM interface for testing
type StubLLM struct {
	// CallCount tracks how many times Summarize was called
	CallCount int
	// LastPR tracks the last PR that was summarized
	LastPR *model.PR
}

// NewStubLLM creates a new stub LLM instance
func NewStubLLM() *StubLLM {
	return &StubLLM{}
}

// Summarize returns a fixed summary for testing
func (s *StubLLM) Summarize(ctx context.Context, pr model.PR) (string, error) {
	s.CallCount++
	s.LastPR = &pr
	
	summary := fmt.Sprintf("This PR %s by @%s introduces significant changes to improve the codebase. Key improvements include enhanced performance, better error handling, and increased test coverage. The changes are well-structured and follow best practices.",
		pr.Title, pr.Author)
	
	return summary, nil
}

// SummarizeBatch generates summaries for multiple PRs
func (s *StubLLM) SummarizeBatch(ctx context.Context, prs []model.PR) (map[string]string, error) {
	summaries := make(map[string]string)
	
	for _, pr := range prs {
		summary, err := s.Summarize(ctx, pr)
		if err != nil {
			return nil, err
		}
		summaries[pr.URL] = summary
	}
	
	return summaries, nil
}

// Name returns the name of the stub provider
func (s *StubLLM) Name() string {
	return "stub"
}
package llm

import "context"

// LLM defines the interface for interacting with Language Model providers.
type LLM interface {
	Summarise(ctx context.Context, context string) (string, error)
}

// StubLLM is a mock implementation of the LLM interface for testing.
type StubLLM struct {
	Summary string
	Err     error
}

// Summarise returns a fixed summary or an error if configured.
func (s *StubLLM) Summarise(ctx context.Context, context string) (string, error) {
	if s.Err != nil {
		return "", s.Err
	}
	return s.Summary, nil
}

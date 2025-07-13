package llm

import (
	"github.com/yourorg/prtool/internal/config"
)

type LLM interface {
	Summarise(context string) (string, error)
}

type StubLLM struct{}

func (s *StubLLM) Summarise(context string) (string, error) {
	return "This is a stub summary.", nil
}

func NewLLM(cfg config.Config) LLM {
	switch cfg.LLMProvider {
	case "openai":
		return NewOpenAIClient(cfg.GitHubToken) // token overloaded for demo
	case "ollama":
		return NewOllamaClient()
	default:
		return &StubLLM{}
	}
}

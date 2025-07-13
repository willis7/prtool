package llm

type LLM interface {
	Summarise(context string) (string, error)
}

type StubLLM struct{}

func (s *StubLLM) Summarise(context string) (string, error) {
	return "This is a stub summary.", nil
}

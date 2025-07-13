package llm

import (
	"os"
	"testing"
)

func TestStubLLM_Summarise(t *testing.T) {
	stub := &StubLLM{}
	got, err := stub.Summarise("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "This is a stub summary."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOpenAIClient_Summarise_Smoke(t *testing.T) {
	key := os.Getenv("OPENAI_KEY")
	if key == "" {
		t.Skip("OPENAI_KEY not set; skipping OpenAI smoke test")
	}
	llm := NewOpenAIClient(key)
	got, err := llm.Summarise("Summarise: The quick brown fox jumps over the lazy dog.")
	if err != nil {
		t.Fatalf("OpenAIClient error: %v", err)
	}
	if got == "" {
		t.Errorf("OpenAIClient returned empty summary")
	}
}

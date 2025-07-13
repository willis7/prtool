package llm

import "testing"

func TestStubLLM_Summarise(t *testing.T) {
	stub := &StubLLM{}
	summary, err := stub.Summarise("context")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "This is a stub summary." {
		t.Errorf("unexpected summary: %q", summary)
	}
}

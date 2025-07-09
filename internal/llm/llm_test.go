package llm

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestStubLLM_Summarise(t *testing.T) {
	tests := []struct {
		name    string
		summary string
		err     error
		expected string
		wantErr bool
	}{
		{
			name:    "Successful summary",
			summary: "This is a test summary.",
			err:     nil,
			expected: "This is a test summary.",
			wantErr: false,
		},
		{
			name:    "Error case",
			summary: "",
			err:     errors.New("LLM error"),
			expected: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &StubLLM{Summary: tt.summary, Err: tt.err}
			got, err := stub.Summarise(context.Background(), "some test context")

			if (err != nil) != tt.wantErr {
				t.Errorf("Summarise() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("Summarise() got = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestOpenAIClient_Summarise_Smoke(t *testing.T) {
	if os.Getenv("OPENAI_KEY") == "" {
		t.Skip("OPENAI_KEY not set, skipping OpenAI smoke test")
	}

	client := NewOpenAIClient(os.Getenv("OPENAI_KEY"), "gpt-3.5-turbo")
	testContext := "Summarize the following: The quick brown fox jumps over the lazy dog."

	summary, err := client.Summarise(context.Background(), testContext)
	if err != nil {
		t.Fatalf("OpenAI Summarise failed: %v", err)
	}

	if summary == "" {
		t.Errorf("OpenAI Summarise returned empty summary")
	}

	// Basic check for content, not too strict as LLM output varies
	if !strings.Contains(summary, "fox") && !strings.Contains(summary, "dog") {
		t.Errorf("Summary did not contain expected keywords: %s", summary)
	}
}
package llm

import (
	"errors"
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
			got, err := stub.Summarise("some context")

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

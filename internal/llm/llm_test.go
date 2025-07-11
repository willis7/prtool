package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/yourorg/prtool/internal/model"
)

func TestNewLLM(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "stub provider",
			config: Config{
				Provider: "stub",
			},
			wantErr: false,
		},
		{
			name: "empty provider defaults to stub",
			config: Config{
				Provider: "",
			},
			wantErr: false,
		},
		{
			name: "openai not yet implemented",
			config: Config{
				Provider: "openai",
			},
			wantErr: true,
			errMsg:  "OpenAI provider not yet implemented",
		},
		{
			name: "ollama not yet implemented",
			config: Config{
				Provider: "ollama",
			},
			wantErr: true,
			errMsg:  "Ollama provider not yet implemented",
		},
		{
			name: "unknown provider",
			config: Config{
				Provider: "unknown",
			},
			wantErr: true,
			errMsg:  "unknown LLM provider: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := NewLLM(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLLM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("NewLLM() error = %v, want error containing %q", err, tt.errMsg)
			}
			if !tt.wantErr && llm == nil {
				t.Error("NewLLM() returned nil without error")
			}
		})
	}
}

func TestStubLLM_Summarize(t *testing.T) {
	stub := NewStubLLM()
	ctx := context.Background()

	pr := model.PR{
		Repository: "owner/repo",
		Number:     123,
		Title:      "Add new feature",
		Author:     "alice",
		URL:        "https://github.com/owner/repo/pull/123",
	}

	// Test single summarization
	summary, err := stub.Summarize(ctx, pr)
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}

	// Check summary contains expected content
	if !strings.Contains(summary, "Add new feature") {
		t.Errorf("Summary should contain PR title, got: %s", summary)
	}
	if !strings.Contains(summary, "@alice") {
		t.Errorf("Summary should contain author, got: %s", summary)
	}
	if !strings.Contains(summary, "significant changes") {
		t.Errorf("Summary should contain standard text, got: %s", summary)
	}

	// Check call tracking
	if stub.CallCount != 1 {
		t.Errorf("CallCount = %d, want 1", stub.CallCount)
	}
	if stub.LastPR == nil || stub.LastPR.Number != 123 {
		t.Errorf("LastPR not correctly tracked")
	}
}

func TestStubLLM_SummarizeBatch(t *testing.T) {
	stub := NewStubLLM()
	ctx := context.Background()

	prs := []model.PR{
		{
			Repository: "owner/repo",
			Number:     123,
			Title:      "Add feature A",
			Author:     "alice",
			URL:        "https://github.com/owner/repo/pull/123",
		},
		{
			Repository: "owner/repo",
			Number:     124,
			Title:      "Fix bug B",
			Author:     "bob",
			URL:        "https://github.com/owner/repo/pull/124",
		},
	}

	summaries, err := stub.SummarizeBatch(ctx, prs)
	if err != nil {
		t.Fatalf("SummarizeBatch() error = %v", err)
	}

	// Check we got summaries for all PRs
	if len(summaries) != len(prs) {
		t.Errorf("got %d summaries, want %d", len(summaries), len(prs))
	}

	// Check each summary
	for _, pr := range prs {
		summary, ok := summaries[pr.URL]
		if !ok {
			t.Errorf("missing summary for PR %s", pr.URL)
			continue
		}
		if !strings.Contains(summary, pr.Title) {
			t.Errorf("summary for %s should contain title %q", pr.URL, pr.Title)
		}
	}

	// Check call count
	if stub.CallCount != 2 {
		t.Errorf("CallCount = %d, want 2", stub.CallCount)
	}
}

func TestStubLLM_Name(t *testing.T) {
	stub := NewStubLLM()
	if stub.Name() != "stub" {
		t.Errorf("Name() = %q, want %q", stub.Name(), "stub")
	}
}
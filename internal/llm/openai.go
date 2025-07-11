package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/yourorg/prtool/internal/model"
)

// OpenAILLM implements the LLM interface using OpenAI's API
type OpenAILLM struct {
	client      *openai.Client
	model       string
	temperature float32
	maxTokens   int
}

// NewOpenAILLM creates a new OpenAI LLM instance
func NewOpenAILLM(config Config) (*OpenAILLM, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	client := openai.NewClient(config.APIKey)

	temperature := float32(config.Temperature)
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 500
	}

	return &OpenAILLM{
		client:      client,
		model:       config.Model,
		temperature: temperature,
		maxTokens:   maxTokens,
	}, nil
}

// Summarize generates a summary for a single pull request
func (o *OpenAILLM) Summarize(ctx context.Context, pr model.PR) (string, error) {
	prompt := buildPRPrompt(pr)

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: o.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that summarizes pull requests. Provide concise, informative summaries focusing on the key changes and their impact.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: o.temperature,
			MaxTokens:   o.maxTokens,
		},
	)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// SummarizeBatch generates summaries for multiple pull requests
func (o *OpenAILLM) SummarizeBatch(ctx context.Context, prs []model.PR) (map[string]string, error) {
	summaries := make(map[string]string)

	// For simplicity, process sequentially
	// In production, you might want to add concurrency with rate limiting
	for _, pr := range prs {
		summary, err := o.Summarize(ctx, pr)
		if err != nil {
			// Log error but continue with other PRs
			summaries[pr.URL] = fmt.Sprintf("Error generating summary: %v", err)
			continue
		}
		summaries[pr.URL] = summary
	}

	return summaries, nil
}

// Name returns the name of the OpenAI provider
func (o *OpenAILLM) Name() string {
	return fmt.Sprintf("openai (%s)", o.model)
}

// buildPRPrompt creates a prompt for summarizing a pull request
func buildPRPrompt(pr model.PR) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Please summarize the following pull request:\n\n"))
	sb.WriteString(fmt.Sprintf("Title: %s\n", pr.Title))
	sb.WriteString(fmt.Sprintf("Author: @%s\n", pr.Author))
	sb.WriteString(fmt.Sprintf("Repository: %s\n", pr.Repository))
	
	if pr.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", pr.Description))
	}
	
	if len(pr.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("\nLabels: %s\n", strings.Join(pr.Labels, ", ")))
	}

	sb.WriteString("\nProvide a concise summary (2-3 sentences) that captures the essence of this pull request, its purpose, and key changes.")

	return sb.String()
}
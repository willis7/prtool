package llm

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the LLM interface for OpenAI.
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new OpenAIClient.
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

// Summarise sends a request to OpenAI to summarise the given context.
func (c *OpenAIClient) Summarise(ctx context.Context, context string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: context,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI chat completion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI chat completion returned no choices")
	}

	return resp.Choices[0].Message.Content, nil
}

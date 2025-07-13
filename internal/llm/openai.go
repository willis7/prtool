package llm

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(token string) *OpenAIClient {
	return &OpenAIClient{client: openai.NewClient(token)}
}

func (o *OpenAIClient) Summarise(contextStr string) (string, error) {
	resp, err := o.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{{Role: "user", Content: contextStr}},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

package llm

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type OllamaClient struct{}

func NewOllamaClient() *OllamaClient {
	return &OllamaClient{}
}

func (o *OllamaClient) Summarise(contextStr string) (string, error) {
	body := map[string]string{"model": "llama2", "prompt": contextStr}
	b, _ := json.Marshal(body)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct{ Response string }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Response, nil
}

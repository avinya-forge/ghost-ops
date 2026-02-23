package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ghost-ops/pkg/protocol"
)

const (
	defaultOllamaModel   = "llama3"
	defaultOllamaBaseURL = "http://localhost:11434/v1"
)

// OllamaProvider implements LLMProvider using Ollama API.
type OllamaProvider struct {
	Model        string
	BaseURL      string
	SystemPrompt string
	Client       *http.Client
}

// NewOllamaProvider creates a new OllamaProvider.
func NewOllamaProvider(baseURL string, model string, systemPrompt string) *OllamaProvider {
	if model == "" {
		model = defaultOllamaModel
	}
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}
	return &OllamaProvider{
		Model:        model,
		BaseURL:      baseURL,
		SystemPrompt: systemPrompt,
		Client:       &http.Client{},
	}
}

// GenerateCode calls Ollama API to generate Go code based on intent.
func (p *OllamaProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, error) {
	userContent := blueprint.Intent
	if len(blueprint.Constraints) > 0 {
		constraintsJSON, err := json.Marshal(blueprint.Constraints)
		if err == nil {
			userContent += fmt.Sprintf("\n\nConstraints: %s", string(constraintsJSON))
		}
	}

	reqBody := chatCompletionRequest{
		Model: p.Model,
		Messages: []message{
			{Role: "system", Content: p.SystemPrompt},
			{Role: "user", Content: userContent},
		},
		Temperature: 0.2, // Low temperature for deterministic code generation
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("ollama API returned error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from Ollama API")
	}

	content := chatResp.Choices[0].Message.Content

	// Strip markdown code blocks if the model included them despite instructions
	content = stripMarkdown(content)

	return content, nil
}

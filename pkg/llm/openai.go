package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ghost-ops/pkg/protocol"
)

const (
	defaultOpenAIModel   = "gpt-4o"
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
)

// OpenAIProvider implements LLMProvider using OpenAI API.
type OpenAIProvider struct {
	APIKey       string
	Model        string
	BaseURL      string
	SystemPrompt string
	Client       *http.Client
}

// NewOpenAIProvider creates a new OpenAIProvider.
func NewOpenAIProvider(apiKey string, model string, baseURL string, systemPrompt string) *OpenAIProvider {
	if model == "" {
		model = defaultOpenAIModel
	}
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	return &OpenAIProvider{
		APIKey:       apiKey,
		Model:        model,
		BaseURL:      baseURL,
		SystemPrompt: systemPrompt,
		Client:       &http.Client{},
	}
}

// GenerateCode calls OpenAI Chat Completion API to generate Go code based on intent.
func (p *OpenAIProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, *protocol.TokenUsage, error) {
	if p.APIKey == "" {
		return "", nil, fmt.Errorf("OpenAI API key is missing")
	}

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
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", nil, fmt.Errorf("OpenAI API returned error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices returned from OpenAI API")
	}

	content := chatResp.Choices[0].Message.Content

	// Strip markdown code blocks if the model included them despite instructions
	content = stripMarkdown(content)

	var tokenUsage *protocol.TokenUsage
	if chatResp.Usage != nil {
		tokenUsage = &protocol.TokenUsage{
			InputTokens:  chatResp.Usage.PromptTokens,
			OutputTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:  chatResp.Usage.TotalTokens,
		}
	}

	return content, tokenUsage, nil
}

func stripMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```go") {
		s = strings.TrimPrefix(s, "```go")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

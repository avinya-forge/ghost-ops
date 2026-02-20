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
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

// NewOpenAIProvider creates a new OpenAIProvider.
func NewOpenAIProvider(apiKey string, model string) *OpenAIProvider {
	if model == "" {
		model = defaultOpenAIModel
	}
	return &OpenAIProvider{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: defaultOpenAIBaseURL,
		Client:  &http.Client{},
	}
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float32   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []choice  `json:"choices"`
	Error   *apiError `json:"error,omitempty"`
}

type choice struct {
	Message message `json:"message"`
}

type apiError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   string      `json:"param"`
	Code    interface{} `json:"code"`
}

// GenerateCode calls OpenAI Chat Completion API to generate Go code based on intent.
func (p *OpenAIProvider) GenerateCode(ctx context.Context, blueprint protocol.Blueprint) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is missing")
	}

	systemPrompt := `You are an expert Go programmer specializing in WebAssembly (WASM).
Your task is to write a valid, compilable Go program that compiles to WASM (GOOS=wasip1 GOARCH=wasm).
The program should be a standalone 'main' package.
Do not include any markdown formatting (e.g. triple backticks).
Output ONLY the raw Go source code.
Ensure the code imports necessary packages and handles errors gracefully.`

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
			{Role: "system", Content: systemPrompt},
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("OpenAI API returned error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI API")
	}

	content := chatResp.Choices[0].Message.Content

	// Strip markdown code blocks if the model included them despite instructions
	content = stripMarkdown(content)

	return content, nil
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

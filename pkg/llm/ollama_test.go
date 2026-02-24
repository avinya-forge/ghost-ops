package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestOllamaProvider_GenerateCode_Success(t *testing.T) {
	expectedCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		// No auth check for Ollama

		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if req.Model != "llama3" {
			t.Errorf("expected model llama3, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(req.Messages))
		}

		// Check system prompt
		if req.Messages[0].Content != "test-system-prompt" {
			t.Errorf("expected system prompt 'test-system-prompt', got %s", req.Messages[0].Content)
		}

		if req.Messages[1].Content != "hello" {
			t.Errorf("expected user message 'hello', got %s", req.Messages[1].Content)
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "```go\n" + expectedCode + "\n```",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	provider := NewOllamaProvider(ts.URL, "llama3", "test-system-prompt")

	blueprint := protocol.Blueprint{
		Intent: "hello",
	}

	code, _, err := provider.GenerateCode(context.Background(), blueprint)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if code != expectedCode {
		t.Errorf("expected code:\n%s\ngot:\n%s", expectedCode, code)
	}
}

func TestOllamaProvider_GenerateCode_DefaultURL(t *testing.T) {
	// Only checking if it uses default URL would require checking internal state
	// or mocking http.Client Transport, which is too much.
	// But we can check if BaseURL is set correctly in struct.
	provider := NewOllamaProvider("", "llama3", "prompt")
	if provider.BaseURL != defaultOllamaBaseURL {
		t.Errorf("expected default BaseURL %s, got %s", defaultOllamaBaseURL, provider.BaseURL)
	}
}

func TestOllamaProvider_GenerateCode_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid request",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	provider := NewOllamaProvider(ts.URL, "llama3", "prompt")

	blueprint := protocol.Blueprint{Intent: "hello"}
	_, _, err := provider.GenerateCode(context.Background(), blueprint)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain 400, got %v", err)
	}
}

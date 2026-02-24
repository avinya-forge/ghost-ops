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

func TestOpenAIProvider_GenerateCode_Success(t *testing.T) {
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
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected Authorization header Bearer test-key, got %s", auth)
		}

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

		if req.Model != "gpt-4o" {
			t.Errorf("expected model gpt-4o, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(req.Messages))
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

	provider := NewOpenAIProvider("test-key", "gpt-4o", ts.URL, "test-system-prompt")

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

func TestOpenAIProvider_GenerateCode_WithConstraints(t *testing.T) {
	expectedCode := `package main`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		userMsg := req.Messages[1].Content
		if !strings.Contains(userMsg, "hello") {
			t.Errorf("expected user message to contain 'hello', got %s", userMsg)
		}
		if !strings.Contains(userMsg, "Constraints:") {
			t.Errorf("expected user message to contain 'Constraints:', got %s", userMsg)
		}
		if !strings.Contains(userMsg, `"p99":50`) {
			t.Errorf("expected user message to contain '\"p99\":50', got %s", userMsg)
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

	provider := NewOpenAIProvider("test-key", "gpt-4o", ts.URL, "test-system-prompt")

	blueprint := protocol.Blueprint{
		Intent: "hello",
		Constraints: map[string]interface{}{
			"p99": 50,
		},
	}

	_, _, err := provider.GenerateCode(context.Background(), blueprint)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
}

func TestOpenAIProvider_GenerateCode_APIError(t *testing.T) {
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

	provider := NewOpenAIProvider("test-key", "gpt-4o", ts.URL, "test-system-prompt")

	blueprint := protocol.Blueprint{Intent: "hello"}
	_, _, err := provider.GenerateCode(context.Background(), blueprint)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain 400, got %v", err)
	}
}

func TestOpenAIProvider_GenerateCode_MissingKey(t *testing.T) {
	provider := NewOpenAIProvider("", "gpt-4o", "", "test-system-prompt")
	blueprint := protocol.Blueprint{Intent: "hello"}
	_, _, err := provider.GenerateCode(context.Background(), blueprint)
	if err == nil {
		t.Fatal("expected error for missing API key, got nil")
	}
	if !strings.Contains(err.Error(), "API key is missing") {
		t.Errorf("expected 'API key is missing' error, got %v", err)
	}
}

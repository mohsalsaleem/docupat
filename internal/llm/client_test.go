package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"docpatch/internal/document"
)

func TestOpenAICompatibleProvider(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("unexpected request: %s, authorization %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Model != "gpt-test" {
			t.Fatalf("request model = %q, error = %v", body.Model, err)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"replacement"}}]}`), nil
	})}

	client, err := NewClient(Config{Provider: "openai", BaseURL: "https://openai.example", APIKey: "secret", Model: "gpt-test"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.Generate(context.Background(), document.GenerateInput{Target: "old", Instruction: "rewrite"})
	if err != nil || got != "replacement" {
		t.Fatalf("Generate() = %q, %v", got, err)
	}
}

func TestAnthropicProvider(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != "secret" || r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Fatalf("unexpected Anthropic request: %s, headers %#v", r.URL.Path, r.Header)
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Model != "claude-test" {
			t.Fatalf("request model = %q, error = %v", body.Model, err)
		}
		return jsonResponse(200, `{"content":[{"type":"text","text":"replacement"}]}`), nil
	})}

	client, err := NewClient(Config{Provider: "anthropic", BaseURL: "https://anthropic.example", APIKey: "secret", Model: "claude-test"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.Generate(context.Background(), document.GenerateInput{Target: "old", Instruction: "rewrite"})
	if err != nil || got != "replacement" {
		t.Fatalf("Generate() = %q, %v", got, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestClientRejectsUnsupportedProvider(t *testing.T) {
	_, err := NewClient(Config{Provider: "unknown", BaseURL: "https://example.com", Model: "test"}, http.DefaultClient)
	if err == nil {
		t.Fatal("expected unsupported provider error")
	}
}

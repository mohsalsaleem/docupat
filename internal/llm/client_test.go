package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"docpatch/internal/document"
	"docpatch/internal/domain"
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

func TestProviderHealth(t *testing.T) {
	for _, test := range []struct {
		name     string
		provider string
		status   int
		want     string
	}{
		{name: "OpenAI connected", provider: "openai", status: 200, want: "connected"},
		{name: "Anthropic unavailable", provider: "anthropic", status: 503, want: "unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return jsonResponse(test.status, `{}`), nil
			})}
			client, err := NewClient(Config{Provider: test.provider, BaseURL: "https://provider.example", APIKey: "secret", Model: "model"}, httpClient)
			if err != nil {
				t.Fatal(err)
			}
			if got := client.Health(context.Background()); got != test.want {
				t.Fatalf("Health() = %q, want %q", got, test.want)
			}
		})
	}

	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})}
	client, err := NewClient(Config{Provider: "openai", BaseURL: "https://provider.example", Model: "model"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	if got := client.Health(context.Background()); got != "unavailable" {
		t.Fatalf("Health() = %q", got)
	}
}

func TestProviderErrors(t *testing.T) {
	for _, test := range []struct {
		name     string
		provider string
		response *http.Response
		failure  error
	}{
		{name: "provider message", provider: "openai", response: jsonResponse(400, `{"error":{"message":"bad request"}}`)},
		{name: "provider status", provider: "anthropic", response: jsonResponse(500, `{}`)},
		{name: "invalid JSON", provider: "openai", response: jsonResponse(200, `{`)},
		{name: "transport", provider: "anthropic", failure: errors.New("offline")},
	} {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return test.response, test.failure
			})}
			client, err := NewClient(Config{Provider: test.provider, BaseURL: "https://provider.example", APIKey: "secret", Model: "model"}, httpClient)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = client.Generate(context.Background(), document.GenerateInput{Target: "target", Context: []domain.ContextItem{{Kind: "reference", Title: "Related", Content: "context"}}, Instruction: "rewrite"}); err == nil {
				t.Fatal("expected provider error")
			}
		})
	}
}

func TestProviderRejectsEmptyResponsesAndConfiguration(t *testing.T) {
	for _, test := range []struct {
		provider string
		body     string
	}{
		{provider: "openai", body: `{"choices":[]}`},
		{provider: "anthropic", body: `{"content":[{"type":"tool_use"}]}`},
	} {
		httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonResponse(200, test.body), nil
		})}
		client, err := NewClient(Config{Provider: test.provider, BaseURL: "https://provider.example", APIKey: "secret", Model: "model"}, httpClient)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = client.Generate(context.Background(), document.GenerateInput{Target: "target", Instruction: "rewrite"}); err == nil {
			t.Fatal("expected empty response error")
		}
	}

	if _, err := NewClient(Config{Provider: "openai", Model: "model"}, http.DefaultClient); err == nil {
		t.Fatal("expected missing base URL error")
	}
	if _, err := NewClient(Config{Provider: "anthropic", BaseURL: "https://provider.example", Model: "model"}, http.DefaultClient); err == nil {
		t.Fatal("expected missing Anthropic key error")
	}
}

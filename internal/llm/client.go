package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"docpatch/internal/document"
)

const systemPrompt = "You are a precise technical document editor. Follow the output contract exactly."

type Config struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

type Client struct {
	provider adapter
}

type adapter interface {
	Generate(context.Context, string) (string, error)
	Health(context.Context) string
}

func NewClient(config Config, client *http.Client) (*Client, error) {
	config.Provider = strings.ToLower(strings.TrimSpace(config.Provider))
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.Model = strings.TrimSpace(config.Model)
	if client == nil || config.BaseURL == "" || config.Model == "" {
		return nil, errors.New("LLM HTTP client, base URL, and model are required")
	}

	var provider adapter
	switch config.Provider {
	case "openai":
		provider = &openAIAdapter{config: config, http: client}
	case "anthropic":
		if config.APIKey == "" {
			return nil, errors.New("LLM_API_KEY or ANTHROPIC_API_KEY is required for Anthropic")
		}
		provider = &anthropicAdapter{config: config, http: client}
	default:
		return nil, fmt.Errorf("unsupported LLM provider %q (use openai or anthropic)", config.Provider)
	}
	return &Client{provider: provider}, nil
}

func (c *Client) Health(ctx context.Context) string { return c.provider.Health(ctx) }

func (c *Client) Generate(ctx context.Context, input document.GenerateInput) (string, error) {
	return c.provider.Generate(ctx, buildPrompt(input))
}

type openAIAdapter struct {
	config Config
	http   *http.Client
}

func (a *openAIAdapter) Health(ctx context.Context) string {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.config.BaseURL+"/v1/models", nil)
	a.authorize(req)
	return health(a.http, req)
}

func (a *openAIAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model": a.config.Model, "temperature": 0.2,
		"messages": []map[string]string{{"role": "system", "content": systemPrompt}, {"role": "user", "content": prompt}},
	}
	var output struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := requestJSON(ctx, a.http, a.config.BaseURL+"/v1/chat/completions", body, a.authorize, &output); err != nil {
		return "", err
	}
	if len(output.Choices) == 0 {
		return "", errors.New("model returned no replacement")
	}
	return nonEmpty(output.Choices[0].Message.Content)
}

func (a *openAIAdapter) authorize(req *http.Request) {
	if a.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
}

type anthropicAdapter struct {
	config Config
	http   *http.Client
}

func (a *anthropicAdapter) Health(ctx context.Context) string {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.config.BaseURL+"/v1/models", nil)
	a.authorize(req)
	return health(a.http, req)
}

func (a *anthropicAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model": a.config.Model, "max_tokens": 4096, "temperature": 0.2, "system": systemPrompt,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	var output struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := requestJSON(ctx, a.http, a.config.BaseURL+"/v1/messages", body, a.authorize, &output); err != nil {
		return "", err
	}
	for _, block := range output.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return strings.TrimSpace(block.Text), nil
		}
	}
	return "", errors.New("model returned no replacement")
}

func (a *anthropicAdapter) authorize(req *http.Request) {
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

func requestJSON(ctx context.Context, client *http.Client, url string, body any, authorize func(*http.Request), output any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	authorize(req)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("LLM provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return providerError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode LLM response: %w", err)
	}
	return nil
}

func providerError(resp *http.Response) error {
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var output struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(payload, &output) == nil && output.Error.Message != "" {
		return errors.New(output.Error.Message)
	}
	return fmt.Errorf("LLM provider returned %d", resp.StatusCode)
}

func health(client *http.Client, req *http.Request) string {
	resp, err := client.Do(req)
	if err != nil {
		return "unavailable"
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "unavailable"
	}
	return "connected"
}

func nonEmpty(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("model returned no replacement")
	}
	return value, nil
}

func buildPrompt(input document.GenerateInput) string {
	var prompt strings.Builder
	prompt.WriteString("Rewrite ONLY the TARGET markdown. Return only replacement markdown with no fences around the response, no commentary, and no diff markers. Preserve unaffected wording and existing headings. If TARGET contains a Mermaid fence, keep the fence and produce valid Mermaid syntax.\n\n")
	if len(input.Context) > 0 {
		prompt.WriteString("READ-ONLY CONTEXT (cannot be edited):\n")
		for _, item := range input.Context {
			prompt.WriteString("--- ")
			prompt.WriteString(item.Kind)
			prompt.WriteString(": ")
			prompt.WriteString(item.Title)
			prompt.WriteString(" ---\n")
			prompt.WriteString(item.Content)
			prompt.WriteString("\n")
		}
		prompt.WriteString("\n\n")
	}
	prompt.WriteString("TARGET:\n")
	prompt.WriteString(input.Target)
	prompt.WriteString("\n\nINSTRUCTION:\n")
	prompt.WriteString(input.Instruction)
	return prompt.String()
}

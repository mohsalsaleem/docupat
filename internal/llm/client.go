package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"docpatch/internal/document"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, client *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: client}
}

func (c *Client) BaseURL() string { return c.baseURL }

func (c *Client) Health(ctx context.Context) string {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/models", nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return "unavailable"
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "unavailable"
	}
	return "connected"
}

func (c *Client) Generate(ctx context.Context, input document.GenerateInput) (string, error) {
	prompt := buildPrompt(input)
	body, _ := json.Marshal(map[string]any{"model": "local-model", "temperature": 0.2, "messages": []map[string]string{{"role": "system", "content": "You are a precise technical document editor. Follow the output contract exactly."}, {"role": "user", "content": prompt}}})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llama-server: %w", err)
	}
	defer resp.Body.Close()
	var output struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		if output.Error != nil {
			return "", errors.New(output.Error.Message)
		}
		return "", fmt.Errorf("llama-server returned %d", resp.StatusCode)
	}
	if len(output.Choices) == 0 {
		return "", errors.New("model returned no replacement")
	}
	return strings.TrimSpace(output.Choices[0].Message.Content), nil
}

func buildPrompt(input document.GenerateInput) string {
	var prompt strings.Builder
	prompt.WriteString("Rewrite ONLY the TARGET markdown. Return only replacement markdown with no fences around the response, no commentary, and no diff markers. Preserve unaffected wording and existing headings. If TARGET contains a Mermaid fence, keep the fence and produce valid Mermaid syntax.\n\n")
	if input.ReadOnly != "" {
		prompt.WriteString("READ-ONLY CONTEXT (cannot be edited):\n")
		prompt.WriteString(input.ReadOnly)
		prompt.WriteString("\n\n")
	}
	prompt.WriteString("TARGET:\n")
	prompt.WriteString(input.Target)
	prompt.WriteString("\n\nINSTRUCTION:\n")
	prompt.WriteString(input.Instruction)
	return prompt.String()
}

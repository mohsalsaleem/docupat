package llm

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type EmbeddingConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type EmbeddingClient struct {
	config EmbeddingConfig
	http   *http.Client
}

func NewEmbeddingClient(config EmbeddingConfig, client *http.Client) (*EmbeddingClient, error) {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.Model = strings.TrimSpace(config.Model)
	if client == nil || config.BaseURL == "" || config.Model == "" {
		return nil, errors.New("embedding HTTP client, base URL, and model are required")
	}
	return &EmbeddingClient{config: config, http: client}, nil
}

func (c *EmbeddingClient) Model() string { return c.config.Model }

func (c *EmbeddingClient) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	if len(inputs) == 0 {
		return nil, nil
	}
	body := map[string]any{"model": c.config.Model, "input": inputs}
	var output struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := requestJSON(ctx, c.http, c.config.BaseURL+"/v1/embeddings", body, c.authorize, &output); err != nil {
		return nil, err
	}
	if len(output.Data) != len(inputs) {
		return nil, errors.New("embedding provider returned an unexpected vector count")
	}
	result := make([][]float64, len(inputs))
	for _, item := range output.Data {
		if item.Index < 0 || item.Index >= len(result) || len(item.Embedding) == 0 {
			return nil, errors.New("embedding provider returned an invalid vector")
		}
		result[item.Index] = item.Embedding
	}
	return result, nil
}

func (c *EmbeddingClient) authorize(req *http.Request) {
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

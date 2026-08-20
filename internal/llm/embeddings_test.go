package llm

import (
	"context"
	"net/http"
	"testing"
)

func TestEmbeddingClientPreservesProviderIndexes(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/embeddings" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
		return jsonResponse(200, `{"data":[{"index":1,"embedding":[0,1]},{"index":0,"embedding":[1,0]}]}`), nil
	})}
	client, err := NewEmbeddingClient(EmbeddingConfig{BaseURL: "https://models.example", APIKey: "secret", Model: "embed"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	vectors, err := client.Embed(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if vectors[0][0] != 1 || vectors[1][1] != 1 || client.Model() != "embed" {
		t.Fatalf("unexpected vectors: %#v", vectors)
	}
}

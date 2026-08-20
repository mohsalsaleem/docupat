package semantic

import (
	"context"
	"testing"

	"docpatch/internal/domain"
)

func TestRetrieveRanksAndCachesSemanticSources(t *testing.T) {
	store := &fakeStore{}
	embedder := &fakeEmbedder{vectors: map[string][]float64{"security change": {1, 0}, "Authentication details": {0.9, 0.1}, "Billing details": {0, 1}}}
	retriever := New(store, embedder, Config{MinimumScore: 0.8, MaximumItems: 2})
	candidates := []domain.IndexedSection{{ID: "auth", Title: "Auth", DocumentTitle: "Security", Content: "Authentication details"}, {ID: "bill", Title: "Billing", DocumentTitle: "Payments", Content: "Billing details"}}
	items, err := retriever.Retrieve(context.Background(), "security change", candidates, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].SectionID != "auth" || items[0].Kind != "semantic" || len(store.items) != 2 {
		t.Fatalf("unexpected result: items=%#v cache=%#v", items, store.items)
	}
	firstCalls := embedder.calls
	if _, err := retriever.Retrieve(context.Background(), "security change", candidates, 1000); err != nil {
		t.Fatal(err)
	}
	if embedder.calls != firstCalls+1 {
		t.Fatalf("candidate cache missed; calls=%d", embedder.calls)
	}
}

type fakeStore struct{ items []domain.SectionEmbedding }

func (f *fakeStore) ListSectionEmbeddings(context.Context, string) ([]domain.SectionEmbedding, error) {
	return f.items, nil
}
func (f *fakeStore) UpsertSectionEmbeddings(_ context.Context, items []domain.SectionEmbedding) error {
	f.items = append(f.items, items...)
	return nil
}

type fakeEmbedder struct {
	vectors map[string][]float64
	calls   int
}

func (f *fakeEmbedder) Model() string { return "test" }
func (f *fakeEmbedder) Embed(_ context.Context, inputs []string) ([][]float64, error) {
	f.calls++
	result := make([][]float64, len(inputs))
	for i, input := range inputs {
		result[i] = f.vectors[input]
	}
	return result, nil
}

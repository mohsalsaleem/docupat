package semantic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"sort"
	"strings"

	"docpatch/internal/domain"
)

type Embedder interface {
	Model() string
	Embed(context.Context, []string) ([][]float64, error)
}

type Store interface {
	ListSectionEmbeddings(context.Context, string) ([]domain.SectionEmbedding, error)
	UpsertSectionEmbeddings(context.Context, []domain.SectionEmbedding) error
}

type Config struct {
	MinimumScore float64
	MaximumItems int
	BatchSize    int
}

type Retriever struct {
	store    Store
	embedder Embedder
	config   Config
}

func New(store Store, embedder Embedder, config Config) *Retriever {
	if config.MinimumScore <= 0 {
		config.MinimumScore = 0.72
	}
	if config.MaximumItems <= 0 {
		config.MaximumItems = 3
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 32
	}
	return &Retriever{store: store, embedder: embedder, config: config}
}

func (r *Retriever) Retrieve(ctx context.Context, query string, candidates []domain.IndexedSection, characterBudget int) ([]domain.ContextItem, error) {
	if characterBudget <= 0 || len(candidates) == 0 || strings.TrimSpace(query) == "" {
		return nil, nil
	}
	cached, err := r.store.ListSectionEmbeddings(ctx, r.embedder.Model())
	if err != nil {
		return nil, err
	}
	cache := map[string]domain.SectionEmbedding{}
	for _, item := range cached {
		cache[item.SectionID] = item
	}
	missing := make([]domain.IndexedSection, 0)
	inputs := make([]string, 0)
	for _, candidate := range candidates {
		hash := contentHash(candidate.Content)
		if item, ok := cache[candidate.ID]; !ok || item.ContentHash != hash {
			missing = append(missing, candidate)
			inputs = append(inputs, candidate.Content)
		}
	}
	for start := 0; start < len(inputs); start += r.config.BatchSize {
		end := min(start+r.config.BatchSize, len(inputs))
		vectors, embedErr := r.embedder.Embed(ctx, inputs[start:end])
		if embedErr != nil {
			return nil, embedErr
		}
		updates := make([]domain.SectionEmbedding, end-start)
		for i, candidate := range missing[start:end] {
			updates[i] = domain.SectionEmbedding{SectionID: candidate.ID, ContentHash: contentHash(candidate.Content), Model: r.embedder.Model(), Vector: vectors[i]}
			cache[candidate.ID] = updates[i]
		}
		if err := r.store.UpsertSectionEmbeddings(ctx, updates); err != nil {
			return nil, err
		}
	}
	queryVectors, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	type match struct {
		section domain.IndexedSection
		score   float64
	}
	matches := make([]match, 0)
	for _, candidate := range candidates {
		score := cosine(queryVectors[0], cache[candidate.ID].Vector)
		if score >= r.config.MinimumScore {
			matches = append(matches, match{section: candidate, score: score})
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].score > matches[j].score })
	items := make([]domain.ContextItem, 0, r.config.MaximumItems)
	remaining := characterBudget
	for _, match := range matches {
		if len(items) >= r.config.MaximumItems || remaining <= 0 {
			break
		}
		content := truncate(strings.TrimSpace(match.section.Content), remaining)
		if content == "" {
			continue
		}
		remaining -= len(content)
		items = append(items, domain.ContextItem{Kind: "semantic", Title: match.section.Title, DocumentID: match.section.DocumentID, DocumentTitle: match.section.DocumentTitle, SectionID: match.section.ID, Content: content, Score: match.score})
	}
	return items, nil
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func cosine(left, right []float64) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for i := range left {
		dot += left[i] * right[i]
		leftNorm += left[i] * left[i]
		rightNorm += right[i] * right[i]
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	end := 0
	for index := range value {
		if index > limit {
			break
		}
		end = index
	}
	return value[:end]
}

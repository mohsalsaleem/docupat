package storage

import (
	"context"
	"path/filepath"
	"testing"

	"docpatch/internal/domain"
	"docpatch/internal/markdownindex"
)

func TestPatchContextRoundTrip(t *testing.T) {
	next := 0
	repository, err := Open(filepath.Join(t.TempDir(), "stellarity.db"), func() string {
		next++
		return string(rune('a' + next))
	}, func() string { return "2026-01-01T00:00:00Z" })
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	document, err := repository.CreateDocument(context.Background(), "PRD", "target")
	if err != nil {
		t.Fatal(err)
	}
	patch := domain.Patch{ID: "patch", DocumentID: document.ID, BaseVersion: 1, Start: 0, End: 6, Original: "target", Replacement: "changed", Instruction: "rewrite", Status: "proposed", CreatedAt: "now", Context: []domain.ContextItem{{Kind: "reference", Title: "Architecture", Content: "## Architecture"}}, Assessment: domain.ContextAssessment{Score: 80, Level: "high", Explicit: 1, Summary: "Strong workspace support"}}
	if err := repository.CreatePatch(context.Background(), patch); err != nil {
		t.Fatal(err)
	}
	items, err := repository.ListPatches(context.Background(), document.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].Context) != 1 || items[0].Context[0].Title != "Architecture" || items[0].Assessment.Score != 80 {
		t.Fatalf("context did not round trip: %#v", items)
	}
}

func TestReplaceDocumentIndexOnlyChangesTargetDocument(t *testing.T) {
	next := 0
	repository, err := Open(filepath.Join(t.TempDir(), "index.db"), func() string {
		next++
		return string(rune('a' + next))
	}, func() string { return "2026-01-01T00:00:00Z" })
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	first, _ := repository.CreateDocument(context.Background(), "First", "# First\n\n[[Second]]")
	second, _ := repository.CreateDocument(context.Background(), "Second", "# Second\n")
	indexer := markdownindex.New()
	if err := repository.ReplaceDocumentIndex(context.Background(), indexer.Index(first)); err != nil {
		t.Fatal(err)
	}
	if err := repository.ReplaceDocumentIndex(context.Background(), indexer.Index(second)); err != nil {
		t.Fatal(err)
	}
	first.Content = "# First\n\n## Changed\n"
	if err := repository.ReplaceDocumentIndex(context.Background(), indexer.Index(first)); err != nil {
		t.Fatal(err)
	}
	sections, _ := repository.ListIndexedSections(context.Background())
	links, _ := repository.ListIndexedLinks(context.Background())
	if len(sections) != 3 || len(links) != 0 {
		t.Fatalf("unexpected incremental index: sections=%#v links=%#v", sections, links)
	}
}

func TestSectionEmbeddingRoundTrip(t *testing.T) {
	next := 0
	repository, err := Open(filepath.Join(t.TempDir(), "embeddings.db"), func() string {
		next++
		return string(rune('a' + next))
	}, func() string { return "2026-01-01T00:00:00Z" })
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	document, _ := repository.CreateDocument(context.Background(), "Design", "# Design\n")
	index := markdownindex.New().Index(document)
	if err := repository.ReplaceDocumentIndex(context.Background(), index); err != nil {
		t.Fatal(err)
	}
	want := domain.SectionEmbedding{SectionID: index.Sections[0].ID, ContentHash: "hash", Model: "embed", Vector: []float64{0.25, 0.75}}
	if err := repository.UpsertSectionEmbeddings(context.Background(), []domain.SectionEmbedding{want}); err != nil {
		t.Fatal(err)
	}
	items, err := repository.ListSectionEmbeddings(context.Background(), "embed")
	if err != nil || len(items) != 1 || items[0].Vector[1] != 0.75 {
		t.Fatalf("embedding round trip failed: items=%#v err=%v", items, err)
	}
	document.Content = "# Design\n\nUpdated body.\n"
	if err := repository.ReplaceDocumentIndex(context.Background(), markdownindex.New().Index(document)); err != nil {
		t.Fatal(err)
	}
	items, err = repository.ListSectionEmbeddings(context.Background(), "embed")
	if err != nil || len(items) != 1 {
		t.Fatalf("stable section embedding was discarded: items=%#v err=%v", items, err)
	}
}

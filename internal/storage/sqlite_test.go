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
	patch := domain.Patch{ID: "patch", DocumentID: document.ID, BaseVersion: 1, Start: 0, End: 6, Original: "target", Replacement: "changed", Instruction: "rewrite", Status: "proposed", CreatedAt: "now", Context: []domain.ContextItem{{Kind: "reference", Title: "Architecture", Content: "## Architecture"}}}
	if err := repository.CreatePatch(context.Background(), patch); err != nil {
		t.Fatal(err)
	}
	items, err := repository.ListPatches(context.Background(), document.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].Context) != 1 || items[0].Context[0].Title != "Architecture" {
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

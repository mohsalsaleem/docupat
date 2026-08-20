package contextcompile

import (
	"context"
	"testing"

	"docpatch/internal/domain"
	"docpatch/internal/markdownindex"
)

func TestCompileResolvesCrossDocumentReferencesAndBacklinks(t *testing.T) {
	api := domain.Document{ID: "api", Title: "API", Content: "# API\n\n## Endpoint\n\nUses [[Authentication#Sessions]].\n"}
	auth := domain.Document{ID: "auth", Title: "Authentication", Content: "# Authentication\n\n## Sessions\n\nTokens expire.\n\n## Consumers\n\nSee [[API#Endpoint]].\n"}
	workspace := indexedWorkspace(api, auth)
	start := len("# API\n\n")
	items, err := New(workspace, 6000).Compile(context.Background(), api, domain.Selection{Start: start, End: len(api.Content)})
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ kind, title, document string }{{"ancestor", "API", "API"}, {"reference", "Sessions", "Authentication"}, {"backlink", "Consumers", "Authentication"}}
	if len(items) != len(want) {
		t.Fatalf("items = %#v", items)
	}
	for i := range want {
		if items[i].Kind != want[i].kind || items[i].Title != want[i].title || items[i].DocumentTitle != want[i].document {
			t.Fatalf("item %d = %#v", i, items[i])
		}
	}
}

func TestCompileHonorsBudgetAndExcludesTarget(t *testing.T) {
	document := domain.Document{ID: "doc", Title: "Doc", Content: "# Root\n\n## Target\n\nUse [[Reference]].\n\n## Reference\n\n1234567890\n"}
	workspace := indexedWorkspace(document)
	start := len("# Root\n\n")
	end := start + len("## Target\n\nUse [[Reference]].\n\n")
	items, err := New(workspace, 12).Compile(context.Background(), document, domain.Selection{Start: start, End: end})
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, item := range items {
		total += len(item.Content)
		if item.Title == "Target" {
			t.Fatal("target leaked into context")
		}
	}
	if total > 12 {
		t.Fatalf("budget exceeded: %d", total)
	}
}

type fakeWorkspace struct {
	sections []domain.IndexedSection
	links    []domain.IndexedLink
}

func indexedWorkspace(documents ...domain.Document) *fakeWorkspace {
	workspace := &fakeWorkspace{}
	for _, document := range documents {
		index := markdownindex.New().Index(document)
		workspace.sections = append(workspace.sections, index.Sections...)
		workspace.links = append(workspace.links, index.Links...)
	}
	return workspace
}

func (f *fakeWorkspace) ListIndexedSections(context.Context) ([]domain.IndexedSection, error) {
	return f.sections, nil
}
func (f *fakeWorkspace) ListIndexedLinks(context.Context) ([]domain.IndexedLink, error) {
	return f.links, nil
}

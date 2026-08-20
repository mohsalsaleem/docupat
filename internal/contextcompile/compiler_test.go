package contextcompile

import (
	"context"
	"strings"
	"testing"

	"docpatch/internal/domain"
	"docpatch/internal/markdownindex"
)

func TestCompileResolvesCrossDocumentReferencesAndBacklinks(t *testing.T) {
	api := domain.Document{ID: "api", Title: "API", Content: "# API\n\n## Endpoint\n\nUses [[Authentication#Sessions]].\n"}
	auth := domain.Document{ID: "auth", Title: "Authentication", Content: "# Authentication\n\n## Sessions\n\nTokens expire.\n\n## Consumers\n\nSee [[API#Endpoint]].\n"}
	workspace := indexedWorkspace(api, auth)
	start := len("# API\n\n")
	compiled, err := New(workspace, 6000).Compile(context.Background(), api, domain.Selection{Start: start, End: len(api.Content)}, "clarify")
	if err != nil {
		t.Fatal(err)
	}
	items := compiled.Items
	want := []struct{ kind, title, document string }{{"ancestor", "API", "API"}, {"reference", "Sessions", "Authentication"}, {"backlink", "Consumers", "Authentication"}}
	if len(items) != len(want) {
		t.Fatalf("items = %#v", items)
	}
	if compiled.Assessment.Level != "high" || compiled.Assessment.Explicit != 2 {
		t.Fatalf("unexpected assessment: %#v", compiled.Assessment)
	}
	if len(compiled.Impacts) != 1 || compiled.Impacts[0].Title != "Consumers" {
		t.Fatalf("unexpected impacts: %#v", compiled.Impacts)
	}
	for i := range want {
		if items[i].Kind != want[i].kind || items[i].Title != want[i].title || items[i].DocumentTitle != want[i].document {
			t.Fatalf("item %d = %#v", i, items[i])
		}
	}
}

func TestCompileUsesSemanticRetrieverOnlyAsFallback(t *testing.T) {
	document := domain.Document{ID: "doc", Title: "Design", Content: "# Target\n\nShort target.\n\n## Related\n\nUseful context.\n"}
	workspace := indexedWorkspace(document)
	semantic := &fakeSemantic{items: []domain.ContextItem{{Kind: "semantic", Title: "Related", SectionID: workspace.sections[1].ID, Content: "Useful context.", Score: 0.91}}}
	compiler := New(workspace, 6000, semantic)
	end := len("# Target\n\nShort target.\n\n")
	compiled, err := compiler.Compile(context.Background(), document, domain.Selection{Start: 0, End: end}, "improve security")
	if err != nil {
		t.Fatal(err)
	}
	if semantic.calls != 1 || len(compiled.Items) != 1 || compiled.Items[0].Kind != "semantic" || !strings.Contains(semantic.query, "improve security") {
		t.Fatalf("fallback not used correctly: calls=%d query=%q items=%#v", semantic.calls, semantic.query, compiled.Items)
	}
}

func TestCompileHonorsBudgetAndExcludesTarget(t *testing.T) {
	document := domain.Document{ID: "doc", Title: "Doc", Content: "# Root\n\n## Target\n\nUse [[Reference]].\n\n## Reference\n\n1234567890\n"}
	workspace := indexedWorkspace(document)
	start := len("# Root\n\n")
	end := start + len("## Target\n\nUse [[Reference]].\n\n")
	compiled, err := New(workspace, 12).Compile(context.Background(), document, domain.Selection{Start: start, End: end}, "rewrite")
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, item := range compiled.Items {
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

type fakeSemantic struct {
	items []domain.ContextItem
	calls int
	query string
}

func (f *fakeSemantic) Retrieve(_ context.Context, query string, _ []domain.IndexedSection, _ int) ([]domain.ContextItem, error) {
	f.calls++
	f.query = query
	return f.items, nil
}

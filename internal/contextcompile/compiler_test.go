package contextcompile

import (
	"testing"

	"docpatch/internal/domain"
)

func TestCompileIncludesAncestorsReferencesAndBacklinks(t *testing.T) {
	content := "# Product\n\nSee [Auth](#authentication).\n\n## API\n\n### Endpoint\n\nUses [[Authentication]].\n\n## Authentication\n\nSessions expire. Link to [Endpoint](#endpoint).\n"
	start := len("# Product\n\nSee [Auth](#authentication).\n\n## API\n\n")
	end := start + len("### Endpoint\n\nUses [[Authentication]].\n\n")
	items, err := New(6000).Compile(content, domain.Selection{Start: start, End: end})
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ kind, title string }{{"ancestor", "Product"}, {"ancestor", "API"}, {"reference", "Authentication"}}
	if len(items) != len(want) {
		t.Fatalf("items = %#v", items)
	}
	for i := range want {
		if items[i].Kind != want[i].kind || items[i].Title != want[i].title {
			t.Fatalf("item %d = %#v", i, items[i])
		}
	}
}

func TestCompileHonorsBudgetAndExcludesTarget(t *testing.T) {
	content := "# Root\n\n## Target\n\nEdit me and use [[Reference]].\n\n## Reference\n\n1234567890\n"
	start := len("# Root\n\n")
	end := start + len("## Target\n\nEdit me and use [[Reference]].\n\n")
	items, err := New(12).Compile(content, domain.Selection{Start: start, End: end})
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

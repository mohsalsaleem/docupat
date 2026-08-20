package markdownindex

import (
	"testing"

	"docpatch/internal/domain"
)

func TestIndexBuildsStableSectionsAndLinks(t *testing.T) {
	document := domain.Document{ID: "doc-1", Title: "API", Content: "# API\n\nSee [[Auth#Sessions]] and [schema](data-model.md#users).\n\n## Endpoint\n\nBody.\n"}
	first := New().Index(document)
	second := New().Index(document)
	if len(first.Sections) != 2 || first.Sections[0].ID != second.Sections[0].ID {
		t.Fatalf("unstable sections: %#v", first.Sections)
	}
	if len(first.Links) != 2 || first.Links[0].TargetDocument != "Auth" || first.Links[0].TargetHeading != "Sessions" || first.Links[1].TargetDocument != "data-model" {
		t.Fatalf("unexpected links: %#v", first.Links)
	}
}

func TestIndexUsesDocumentAsSectionWithoutHeadings(t *testing.T) {
	index := New().Index(domain.Document{ID: "notes", Title: "Notes", Content: "Plain text"})
	if len(index.Sections) != 1 || index.Sections[0].Title != "Notes" || index.Sections[0].Content != "Plain text" {
		t.Fatalf("unexpected root: %#v", index.Sections)
	}
}

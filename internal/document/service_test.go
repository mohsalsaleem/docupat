package document

import (
	"context"
	"testing"

	"docpatch/internal/domain"
)

func TestProposeSeparatesTargetFromReadOnlyContext(t *testing.T) {
	repository := &fakeRepository{document: domain.Document{ID: "doc-1", Content: "before TARGET after", Version: 3}}
	generator := &fakeGenerator{replacement: "REPLACED"}
	service := NewService(repository, generator, func() string { return "patch-1" }, func() string { return "now" })

	patch, err := service.Propose(context.Background(), ProposeInput{DocumentID: "doc-1", Version: 3, Selection: domain.Selection{Start: 7, End: 13}, Instruction: "rewrite", UseContext: true})
	if err != nil {
		t.Fatal(err)
	}
	if generator.input.Target != "TARGET" {
		t.Fatalf("target = %q", generator.input.Target)
	}
	if generator.input.ReadOnly != "before \n[…]\n after" {
		t.Fatalf("context = %q", generator.input.ReadOnly)
	}
	if patch.Original != "TARGET" || patch.Replacement != "REPLACED" || repository.patch.ID != "patch-1" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestProposeRejectsStaleDocumentVersion(t *testing.T) {
	repository := &fakeRepository{document: domain.Document{ID: "doc-1", Content: "target", Version: 2}}
	service := NewService(repository, &fakeGenerator{}, func() string { return "id" }, func() string { return "now" })
	_, err := service.Propose(context.Background(), ProposeInput{DocumentID: "doc-1", Version: 1, Selection: domain.Selection{Start: 0, End: 6}, Instruction: "rewrite"})
	if err != domain.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

type fakeGenerator struct {
	input       GenerateInput
	replacement string
}

func (f *fakeGenerator) Generate(_ context.Context, input GenerateInput) (string, error) {
	f.input = input
	return f.replacement, nil
}
func (f *fakeGenerator) Health(context.Context) string { return "connected" }

type fakeRepository struct {
	document domain.Document
	patch    domain.Patch
}

func (f *fakeRepository) ListDocuments(context.Context) ([]domain.Document, error) {
	return []domain.Document{f.document}, nil
}
func (f *fakeRepository) GetDocument(context.Context, string) (domain.Document, error) {
	return f.document, nil
}
func (f *fakeRepository) CreateDocument(context.Context, string, string) (domain.Document, error) {
	return f.document, nil
}
func (f *fakeRepository) SaveDocument(context.Context, string, int, string, string) (domain.Document, error) {
	return f.document, nil
}
func (f *fakeRepository) ListPatches(context.Context, string) ([]domain.Patch, error) {
	return nil, nil
}
func (f *fakeRepository) CreatePatch(_ context.Context, p domain.Patch) error {
	f.patch = p
	return nil
}
func (f *fakeRepository) ApplyPatch(context.Context, string) (domain.Document, error) {
	return f.document, nil
}
func (f *fakeRepository) RejectPatch(context.Context, string) error { return nil }

package document

import (
	"context"
	"testing"

	"docpatch/internal/domain"
)

func TestProposeSeparatesTargetFromReadOnlyContext(t *testing.T) {
	repository := &fakeRepository{document: domain.Document{ID: "doc-1", Content: "before TARGET after", Version: 3}}
	generator := &fakeGenerator{replacement: "REPLACED"}
	compiler := &fakeCompiler{items: []domain.ContextItem{{Kind: "reference", Title: "Related", Content: "context"}}}
	service := NewService(repository, generator, compiler, fakeIndexer{}, func() string { return "patch-1" }, func() string { return "now" })

	patch, err := service.Propose(context.Background(), ProposeInput{DocumentID: "doc-1", Version: 3, Selection: domain.Selection{Start: 7, End: 13}, Instruction: "rewrite", UseContext: true})
	if err != nil {
		t.Fatal(err)
	}
	if generator.input.Target != "TARGET" {
		t.Fatalf("target = %q", generator.input.Target)
	}
	if len(generator.input.Context) != 1 || generator.input.Context[0].Content != "context" {
		t.Fatalf("context = %#v", generator.input.Context)
	}
	if patch.Original != "TARGET" || patch.Replacement != "REPLACED" || repository.patch.ID != "patch-1" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
	if len(patch.Context) != 1 || len(repository.patch.Context) != 1 {
		t.Fatalf("compiled context was not persisted on patch: %+v", patch.Context)
	}
}

func TestProposeRejectsStaleDocumentVersion(t *testing.T) {
	repository := &fakeRepository{document: domain.Document{ID: "doc-1", Content: "target", Version: 2}}
	service := NewService(repository, &fakeGenerator{}, &fakeCompiler{}, fakeIndexer{}, func() string { return "id" }, func() string { return "now" })
	_, err := service.Propose(context.Background(), ProposeInput{DocumentID: "doc-1", Version: 1, Selection: domain.Selection{Start: 0, End: 6}, Instruction: "rewrite"})
	if err != domain.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestPreserveBoundaryWhitespace(t *testing.T) {
	got := preserveBoundaryWhitespace("\n## Target\nBody\n\n", "## Target\nNew body")
	if got != "\n## Target\nNew body\n\n" {
		t.Fatalf("preserveBoundaryWhitespace() = %q", got)
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

type fakeCompiler struct{ items []domain.ContextItem }

func (f *fakeCompiler) Compile(context.Context, domain.Document, domain.Selection) ([]domain.ContextItem, error) {
	return f.items, nil
}

type fakeIndexer struct{}

func (fakeIndexer) Index(document domain.Document) domain.DocumentIndex {
	return domain.DocumentIndex{DocumentID: document.ID}
}

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
func (f *fakeRepository) ReplaceDocumentIndex(context.Context, domain.DocumentIndex) error {
	return nil
}
func (f *fakeRepository) ListIndexedSections(context.Context) ([]domain.IndexedSection, error) {
	return nil, nil
}
func (f *fakeRepository) ListIndexedLinks(context.Context) ([]domain.IndexedLink, error) {
	return nil, nil
}

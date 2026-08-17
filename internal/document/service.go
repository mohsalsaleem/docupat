package document

import (
	"context"
	"errors"
	"strings"

	"docpatch/internal/domain"
)

type Repository interface {
	ListDocuments(context.Context) ([]domain.Document, error)
	GetDocument(context.Context, string) (domain.Document, error)
	CreateDocument(context.Context, string, string) (domain.Document, error)
	SaveDocument(context.Context, string, int, string, string) (domain.Document, error)
	ListPatches(context.Context, string) ([]domain.Patch, error)
	CreatePatch(context.Context, domain.Patch) error
	ApplyPatch(context.Context, string) (domain.Document, error)
	RejectPatch(context.Context, string) error
}

type Generator interface {
	Generate(context.Context, GenerateInput) (string, error)
	Health(context.Context) string
}

type GenerateInput struct {
	Target      string
	ReadOnly    string
	Instruction string
}

type IDGenerator func() string
type Clock func() string

type Service struct {
	repository Repository
	generator  Generator
	newID      IDGenerator
	now        Clock
}

func NewService(repository Repository, generator Generator, newID IDGenerator, now Clock) *Service {
	return &Service{repository: repository, generator: generator, newID: newID, now: now}
}

func (s *Service) List(ctx context.Context) ([]domain.Document, error) {
	return s.repository.ListDocuments(ctx)
}
func (s *Service) Get(ctx context.Context, id string) (domain.Document, error) {
	return s.repository.GetDocument(ctx, id)
}
func (s *Service) Patches(ctx context.Context, id string) ([]domain.Patch, error) {
	return s.repository.ListPatches(ctx, id)
}
func (s *Service) Apply(ctx context.Context, id string) (domain.Document, error) {
	return s.repository.ApplyPatch(ctx, id)
}
func (s *Service) Reject(ctx context.Context, id string) error {
	return s.repository.RejectPatch(ctx, id)
}
func (s *Service) LLMHealth(ctx context.Context) string { return s.generator.Health(ctx) }

func (s *Service) Create(ctx context.Context, title, content string) (domain.Document, error) {
	if strings.TrimSpace(title) == "" {
		return domain.Document{}, domain.ErrInvalid
	}
	return s.repository.CreateDocument(ctx, strings.TrimSpace(title), content)
}

func (s *Service) Save(ctx context.Context, id string, version int, title, content string) (domain.Document, error) {
	if id == "" || version < 1 || strings.TrimSpace(title) == "" {
		return domain.Document{}, domain.ErrInvalid
	}
	return s.repository.SaveDocument(ctx, id, version, strings.TrimSpace(title), content)
}

type ProposeInput struct {
	DocumentID  string
	Version     int
	Selection   domain.Selection
	Instruction string
	UseContext  bool
}

func (s *Service) Propose(ctx context.Context, input ProposeInput) (domain.Patch, error) {
	if strings.TrimSpace(input.Instruction) == "" {
		return domain.Patch{}, domain.ErrInvalid
	}
	doc, err := s.repository.GetDocument(ctx, input.DocumentID)
	if err != nil {
		return domain.Patch{}, err
	}
	if doc.Version != input.Version {
		return domain.Patch{}, domain.ErrConflict
	}
	start, end, err := input.Selection.ByteRange(doc.Content)
	if err != nil {
		return domain.Patch{}, err
	}
	target := doc.Content[start:end]
	readOnly := ""
	if input.UseContext {
		readOnly = doc.Content[:start] + "\n[…]\n" + doc.Content[end:]
	}
	replacement, err := s.generator.Generate(ctx, GenerateInput{Target: target, ReadOnly: readOnly, Instruction: input.Instruction})
	if err != nil {
		return domain.Patch{}, err
	}
	if strings.TrimSpace(replacement) == "" {
		return domain.Patch{}, errors.New("model returned no replacement")
	}
	p := domain.Patch{ID: s.newID(), DocumentID: doc.ID, BaseVersion: doc.Version, Start: input.Selection.Start, End: input.Selection.End, Original: target, Replacement: preserveBoundaryWhitespace(target, replacement), Instruction: input.Instruction, Status: "proposed", CreatedAt: s.now()}
	if err := s.repository.CreatePatch(ctx, p); err != nil {
		return domain.Patch{}, err
	}
	return p, nil
}

func preserveBoundaryWhitespace(original, replacement string) string {
	leading := original[:len(original)-len(strings.TrimLeft(original, " \t\r\n"))]
	trailing := original[len(strings.TrimRight(original, " \t\r\n")):]
	return leading + strings.TrimSpace(replacement) + trailing
}

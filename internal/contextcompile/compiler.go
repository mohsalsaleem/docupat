package contextcompile

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"docpatch/internal/domain"
	"docpatch/internal/markdownindex"
)

const defaultBudget = 6000

type WorkspaceReader interface {
	ListIndexedSections(context.Context) ([]domain.IndexedSection, error)
	ListIndexedLinks(context.Context) ([]domain.IndexedLink, error)
}

type Compiler struct {
	workspace     WorkspaceReader
	MaxCharacters int
}

func New(workspace WorkspaceReader, maxCharacters int) *Compiler {
	if maxCharacters <= 0 {
		maxCharacters = defaultBudget
	}
	return &Compiler{workspace: workspace, MaxCharacters: maxCharacters}
}

// Compile resolves structural and explicit workspace relationships without a model call.
func (c *Compiler) Compile(ctx context.Context, document domain.Document, selection domain.Selection) ([]domain.ContextItem, error) {
	start, end, err := selection.ByteRange(document.Content)
	if err != nil {
		return nil, err
	}
	sections, err := c.workspace.ListIndexedSections(ctx)
	if err != nil {
		return nil, err
	}
	links, err := c.workspace.ListIndexedLinks(ctx)
	if err != nil {
		return nil, err
	}
	current := documentSections(sections, document.ID)
	selected := containingSection(current, start, end)
	if selected == nil {
		return nil, nil
	}

	items := make([]domain.ContextItem, 0)
	seen := map[string]bool{selected.ID: true}
	remaining := c.MaxCharacters
	add := func(section domain.IndexedSection, kind string, headingOnly bool) {
		if seen[section.ID] || remaining <= 0 {
			return
		}
		body := strings.TrimSpace(section.Content)
		if headingOnly {
			if newline := strings.IndexByte(body, '\n'); newline >= 0 {
				body = body[:newline]
			}
		}
		body = truncate(body, remaining)
		if body == "" {
			return
		}
		seen[section.ID] = true
		remaining -= len(body)
		items = append(items, domain.ContextItem{Kind: kind, Title: section.Title, DocumentID: section.DocumentID, DocumentTitle: section.DocumentTitle, SectionID: section.ID, Content: body})
	}

	for _, ancestor := range ancestors(current, *selected) {
		add(ancestor, "ancestor", true)
	}
	for _, link := range links {
		if link.SourceSectionID == selected.ID {
			if target, ok := resolveTarget(sections, document.ID, link); ok {
				add(target, "reference", false)
			}
		}
	}
	for _, link := range links {
		if target, ok := resolveTarget(sections, sourceDocument(sections, link.SourceSectionID), link); ok && target.ID == selected.ID {
			if source, found := sectionByID(sections, link.SourceSectionID); found {
				add(source, "backlink", false)
			}
		}
	}
	return items, nil
}

func documentSections(sections []domain.IndexedSection, documentID string) []domain.IndexedSection {
	result := make([]domain.IndexedSection, 0)
	for _, section := range sections {
		if section.DocumentID == documentID {
			result = append(result, section)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Start < result[j].Start })
	return result
}

func containingSection(sections []domain.IndexedSection, start, end int) *domain.IndexedSection {
	var best *domain.IndexedSection
	for index := range sections {
		candidate := &sections[index]
		if candidate.Start <= start && candidate.End >= end && (best == nil || candidate.Level >= best.Level) {
			best = candidate
		}
	}
	return best
}

func ancestors(sections []domain.IndexedSection, selected domain.IndexedSection) []domain.IndexedSection {
	result := []domain.IndexedSection{}
	level := selected.Level
	for i := len(sections) - 1; i >= 0; i-- {
		candidate := sections[i]
		if candidate.Start >= selected.Start || candidate.Level >= level {
			continue
		}
		result = append([]domain.IndexedSection{candidate}, result...)
		level = candidate.Level
	}
	return result
}

func resolveTarget(sections []domain.IndexedSection, sourceDocumentID string, link domain.IndexedLink) (domain.IndexedSection, bool) {
	documentID := sourceDocumentID
	heading := link.TargetHeading
	if link.TargetDocument != "" {
		if matched := findDocument(sections, link.TargetDocument); matched != "" {
			documentID = matched
		} else if heading == "" {
			heading = link.TargetDocument
		}
	}
	candidates := documentSections(sections, documentID)
	if heading != "" {
		wanted := markdownindex.Slug(heading)
		for _, section := range candidates {
			if section.Slug == wanted {
				return section, true
			}
		}
	}
	if len(candidates) > 0 && (link.TargetDocument != "" || heading == "") {
		return candidates[0], true
	}
	return domain.IndexedSection{}, false
}

func findDocument(sections []domain.IndexedSection, target string) string {
	wanted := markdownindex.Slug(strings.TrimSuffix(filepath.Base(target), filepath.Ext(target)))
	for _, section := range sections {
		if markdownindex.Slug(section.DocumentTitle) == wanted || markdownindex.Slug(section.DocumentID) == wanted {
			return section.DocumentID
		}
	}
	return ""
}

func sourceDocument(sections []domain.IndexedSection, sectionID string) string {
	if section, ok := sectionByID(sections, sectionID); ok {
		return section.DocumentID
	}
	return ""
}

func sectionByID(sections []domain.IndexedSection, id string) (domain.IndexedSection, bool) {
	for _, section := range sections {
		if section.ID == id {
			return section, true
		}
	}
	return domain.IndexedSection{}, false
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

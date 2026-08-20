package markdownindex

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"docpatch/internal/domain"
)

var (
	headingPattern  = regexp.MustCompile(`(?m)^(#{1,6})[ \t]+(.+?)[ \t]*$`)
	wikiPattern     = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	markdownPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
)

type Indexer struct{}

func New() *Indexer { return &Indexer{} }

func (i *Indexer) Index(document domain.Document) domain.DocumentIndex {
	sections := parseSections(document)
	links := make([]domain.IndexedLink, 0)
	for index, source := range sections {
		directEnd := source.End
		if index+1 < len(sections) && sections[index+1].Start < directEnd {
			directEnd = sections[index+1].Start
		}
		links = append(links, parseLinks(source.ID, document.Content[source.Start:directEnd])...)
	}
	return domain.DocumentIndex{DocumentID: document.ID, Sections: sections, Links: links}
}

func parseSections(document domain.Document) []domain.IndexedSection {
	matches := headingPattern.FindAllStringSubmatchIndex(document.Content, -1)
	if len(matches) == 0 {
		return []domain.IndexedSection{{ID: stableID(document.ID, "root"), DocumentID: document.ID, DocumentTitle: document.Title, Title: document.Title, Slug: Slug(document.Title), Start: 0, End: len(document.Content), Content: document.Content}}
	}
	result := make([]domain.IndexedSection, 0, len(matches))
	path := make([]string, 6)
	occurrences := map[string]int{}
	for index, match := range matches {
		level := match[3] - match[2]
		title := strings.TrimSpace(document.Content[match[4]:match[5]])
		path[level-1] = Slug(title)
		for j := level; j < len(path); j++ {
			path[j] = ""
		}
		key := strings.Join(path[:level], "/")
		occurrences[key]++
		if occurrences[key] > 1 {
			key += ":" + strconv.Itoa(occurrences[key])
		}
		end := len(document.Content)
		for next := index + 1; next < len(matches); next++ {
			if matches[next][3]-matches[next][2] <= level {
				end = matches[next][0]
				break
			}
		}
		result = append(result, domain.IndexedSection{ID: stableID(document.ID, key), DocumentID: document.ID, DocumentTitle: document.Title, Title: title, Slug: Slug(title), Level: level, Start: match[0], End: end, Content: document.Content[match[0]:end]})
	}
	return result
}

func parseLinks(sourceID, content string) []domain.IndexedLink {
	result := make([]domain.IndexedLink, 0)
	for _, match := range wikiPattern.FindAllStringSubmatch(content, -1) {
		document, heading := splitTarget(match[1])
		result = append(result, domain.IndexedLink{SourceSectionID: sourceID, TargetDocument: document, TargetHeading: heading, Kind: "wiki"})
	}
	for _, match := range markdownPattern.FindAllStringSubmatch(content, -1) {
		target := match[1]
		if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		document, heading := splitMarkdownTarget(target)
		result = append(result, domain.IndexedLink{SourceSectionID: sourceID, TargetDocument: document, TargetHeading: heading, Kind: "markdown"})
	}
	return result
}

func splitTarget(value string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(value), "#", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func splitMarkdownTarget(value string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(value), "#", 2)
	document := strings.TrimSuffix(parts[0], ".md")
	heading := ""
	if len(parts) == 2 {
		heading = parts[1]
	}
	return document, heading
}

func stableID(documentID, key string) string {
	sum := sha256.Sum256([]byte(documentID + "\x00" + key))
	return documentID + ":" + hex.EncodeToString(sum[:8])
}

func Slug(value string) string {
	var result strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			dash = false
		} else if !dash && result.Len() > 0 {
			result.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(result.String(), "-")
}

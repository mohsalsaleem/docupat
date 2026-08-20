package contextcompile

import (
	"regexp"
	"strings"
	"unicode"

	"docpatch/internal/domain"
)

const defaultBudget = 6000

var (
	headingPattern = regexp.MustCompile(`(?m)^(#{1,6})[ \t]+(.+?)[ \t]*$`)
	wikiPattern    = regexp.MustCompile(`\[\[([^\]|#]+)(?:#[^\]|]+)?(?:\|[^\]]+)?\]\]`)
	anchorPattern  = regexp.MustCompile(`\[[^\]]+\]\(#([^)]+)\)`)
)

type Compiler struct{ MaxCharacters int }

type section struct {
	title      string
	level      int
	start, end int
	headingEnd int
	content    string
}

func New(maxCharacters int) *Compiler {
	if maxCharacters <= 0 {
		maxCharacters = defaultBudget
	}
	return &Compiler{MaxCharacters: maxCharacters}
}

// Compile builds deterministic context from Markdown structure and explicit links.
// It deliberately performs no model calls and never includes the selected range.
func (c *Compiler) Compile(content string, selection domain.Selection) ([]domain.ContextItem, error) {
	start, end, err := selection.ByteRange(content)
	if err != nil {
		return nil, err
	}
	sections := parseSections(content)
	selected := containingSection(sections, start, end)
	items := make([]domain.ContextItem, 0)
	seen := map[int]bool{}
	remaining := c.MaxCharacters
	add := func(index int, kind, body string) {
		if index < 0 || seen[index] || remaining <= 0 {
			return
		}
		body = strings.TrimSpace(body)
		if body == "" {
			return
		}
		if len(body) > remaining {
			body = body[:remaining]
		}
		seen[index] = true
		remaining -= len(body)
		items = append(items, domain.ContextItem{Kind: kind, Title: sections[index].title, Content: body})
	}

	if selected >= 0 {
		for _, index := range ancestors(sections, selected) {
			add(index, "ancestor", content[sections[index].start:sections[index].headingEnd])
		}
	}

	target := content[start:end]
	for _, reference := range references(target) {
		if index := findSection(sections, reference); index >= 0 && index != selected {
			add(index, "reference", sections[index].content)
		}
	}

	if selected >= 0 {
		selectedSlug := slug(sections[selected].title)
		for index, candidate := range sections {
			if index != selected && containsReference(candidate.content, selectedSlug) {
				add(index, "backlink", candidate.content)
			}
		}
	}
	return items, nil
}

func parseSections(content string) []section {
	matches := headingPattern.FindAllStringSubmatchIndex(content, -1)
	result := make([]section, 0, len(matches))
	for i, match := range matches {
		level := match[3] - match[2]
		end := len(content)
		for j := i + 1; j < len(matches); j++ {
			if matches[j][3]-matches[j][2] <= level {
				end = matches[j][0]
				break
			}
		}
		result = append(result, section{title: strings.TrimSpace(content[match[4]:match[5]]), level: level, start: match[0], headingEnd: match[1], end: end, content: content[match[0]:end]})
	}
	return result
}

func containingSection(sections []section, start, end int) int {
	best := -1
	for i, candidate := range sections {
		if candidate.start <= start && candidate.end >= end && (best < 0 || candidate.level >= sections[best].level) {
			best = i
		}
	}
	return best
}

func ancestors(sections []section, selected int) []int {
	result := []int{}
	level := sections[selected].level
	for i := selected - 1; i >= 0 && level > 1; i-- {
		if sections[i].level < level {
			result = append([]int{i}, result...)
			level = sections[i].level
		}
	}
	return result
}

func references(content string) []string {
	result := []string{}
	for _, match := range wikiPattern.FindAllStringSubmatch(content, -1) {
		result = append(result, match[1])
	}
	for _, match := range anchorPattern.FindAllStringSubmatch(content, -1) {
		result = append(result, match[1])
	}
	return result
}

func findSection(sections []section, reference string) int {
	wanted := slug(reference)
	for i, candidate := range sections {
		if slug(candidate.title) == wanted {
			return i
		}
	}
	return -1
}

func containsReference(content, wanted string) bool {
	for _, reference := range references(content) {
		if slug(reference) == wanted {
			return true
		}
	}
	return false
}

func slug(value string) string {
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

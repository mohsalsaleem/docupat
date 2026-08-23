package domain

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid input")
)

type Document struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content,omitempty"`
	Excerpt   string `json:"excerpt,omitempty"`
	Version   int    `json:"version"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Patch struct {
	ID          string            `json:"id"`
	DocumentID  string            `json:"documentId"`
	BaseVersion int               `json:"baseVersion"`
	Start       int               `json:"start"`
	End         int               `json:"end"`
	Original    string            `json:"original"`
	Replacement string            `json:"replacement"`
	Instruction string            `json:"instruction"`
	Status      string            `json:"status"`
	CreatedAt   string            `json:"createdAt"`
	AppliedAt   *string           `json:"appliedAt"`
	Context     []ContextItem     `json:"context"`
	Assessment  ContextAssessment `json:"assessment"`
	Impacts     []ImpactFinding   `json:"impacts"`
}

type ContextItem struct {
	Kind          string  `json:"kind"`
	Title         string  `json:"title"`
	DocumentID    string  `json:"documentId"`
	DocumentTitle string  `json:"documentTitle"`
	SectionID     string  `json:"sectionId"`
	Content       string  `json:"content"`
	Score         float64 `json:"score,omitempty"`
}

type IndexedSection struct {
	ID            string
	DocumentID    string
	DocumentTitle string
	Title         string
	Slug          string
	Level         int
	Start         int
	End           int
	Content       string
}

type IndexedLink struct {
	SourceSectionID string
	TargetDocument  string
	TargetHeading   string
	Kind            string
}

type DocumentIndex struct {
	DocumentID string
	Sections   []IndexedSection
	Links      []IndexedLink
}

type SectionEmbedding struct {
	SectionID   string
	ContentHash string
	Model       string
	Vector      []float64
}

type CompiledContext struct {
	Items      []ContextItem
	Assessment ContextAssessment
	Impacts    []ImpactFinding
}

type ImpactFinding struct {
	Kind          string  `json:"kind"`
	DocumentID    string  `json:"documentId"`
	DocumentTitle string  `json:"documentTitle"`
	SectionID     string  `json:"sectionId"`
	Title         string  `json:"title"`
	Reason        string  `json:"reason"`
	Score         float64 `json:"score,omitempty"`
}

type ContextAssessment struct {
	Score      int    `json:"score"`
	Level      string `json:"level"`
	Structural int    `json:"structural"`
	Explicit   int    `json:"explicit"`
	Semantic   int    `json:"semantic"`
	Unresolved int    `json:"unresolved"`
	Summary    string `json:"summary"`
}

type Selection struct {
	Start int
	End   int
}

// ByteRange converts CodeMirror's UTF-16 offsets to safe Go UTF-8 byte offsets.
func (s Selection) ByteRange(content string) (int, int, error) {
	start, startOK := UTF16OffsetToByte(content, s.Start)
	end, endOK := UTF16OffsetToByte(content, s.End)
	if !startOK || !endOK || s.End <= s.Start {
		return 0, 0, ErrInvalid
	}
	return start, end, nil
}

func UTF16OffsetToByte(value string, offset int) (int, bool) {
	if offset < 0 {
		return 0, false
	}
	units := 0
	for byteIndex, r := range value {
		if units == offset {
			return byteIndex, true
		}
		width := 1
		if r > 0xFFFF {
			width = 2
		}
		if units+width > offset {
			return 0, false
		}
		units += width
	}
	return len(value), units == offset
}

func Apply(content string, selection Selection, expected, replacement string) (string, error) {
	start, end, err := selection.ByteRange(content)
	if err != nil || content[start:end] != expected {
		return "", ErrConflict
	}
	return content[:start] + replacement + content[end:], nil
}

package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/metalagman/tgbotspec/internal/openapi"
)

const (
	fieldNameColumnIndex        = 0
	fieldTypeColumnIndex        = 1
	fieldDescriptionColumnIndex = 2
	minUnionParts               = 2
)

// TypeRef represents a raw type string from the Telegram docs and helpers to
// convert it into OpenAPI friendly structures.
type TypeRef struct {
	RawType string
}

// NewTypeRef builds a TypeRef wrapper for the given raw type string.
func NewTypeRef(rawType string) *TypeRef {
	return &TypeRef{RawType: rawType}
}

// UnionParts splits a raw type that represents a union into its individual
// parts. It supports the following list forms commonly found in the Telegram
// docs:
//   - "A or B"
//   - "A and B"
//   - "A, B and C"
//   - "A, B, C" (less common, but handled)
//
// Returns trimmed non-empty parts when 2 or more items are present.
// If the type is not a union, it returns nil.
func (t *TypeRef) UnionParts() []string {
	raw := strings.TrimSpace(t.RawType)
	if raw == "" {
		return nil
	}
	// Normalize connectors to commas, then split.
	norm := raw
	norm = strings.ReplaceAll(norm, " or ", ", ")

	norm = strings.ReplaceAll(norm, " and ", ", ")
	if !strings.Contains(norm, ",") {
		return nil
	}

	parts := strings.Split(norm, ",")

	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		res = append(res, p)
	}

	if len(res) < minUnionParts {
		return nil
	}

	return res
}

// ToTypeSpec converts the parsed RawType into an OpenAPI TypeSpec.
// It handles nested arrays ("Array of X"), union types ("A or B" / "A, B and C"),
// primitive scalars, and the Telegram-specific pseudo-type "True" which means
// a boolean literal true (default: true).
func (t *TypeRef) ToTypeSpec() *openapi.TypeSpec { //nolint:cyclop // mapping type variants requires branching
	if t == nil || strings.TrimSpace(t.RawType) == "" {
		return &openapi.TypeSpec{}
	}

	raw := strings.TrimSpace(t.RawType)

	// Handle Arrays recursively: patterns like "Array of X" possibly nested
	const prefix = "Array of "
	if strings.HasPrefix(raw, prefix) {
		inner := strings.TrimSpace(strings.TrimPrefix(raw, prefix))

		return &openapi.TypeSpec{
			Type:  "array",
			Items: NewTypeRef(inner).ToTypeSpec(),
		}
	}
	// Also handle lower-case phrasing like "array of array of X"
	if strings.HasPrefix(strings.ToLower(raw), "array of array of ") {
		inner := strings.TrimSpace(raw[len("Array of "):])

		return &openapi.TypeSpec{
			Type:  "array",
			Items: NewTypeRef(inner).ToTypeSpec(),
		}
	}

	// Handle union types detected by the parser's TypeRef.UnionParts
	if parts := t.UnionParts(); parts != nil {
		anyOf := make([]openapi.TypeSpec, 0, len(parts))
		for _, p := range parts {
			anyOf = append(anyOf, *NewTypeRef(p).ToTypeSpec())
		}

		if len(anyOf) >= minUnionParts {
			return &openapi.TypeSpec{AnyOf: anyOf}
		}
	}

	switch strings.ToLower(raw) {
	case "string":
		return &openapi.TypeSpec{Type: "string"}
	case "integer", "int":
		return &openapi.TypeSpec{Type: "integer"}
	case "integer64", "int64":
		return &openapi.TypeSpec{Type: "integer", Format: "int64"}
	case "float", "float number", "number":
		return &openapi.TypeSpec{Type: "number"}
	case "boolean", "bool":
		return &openapi.TypeSpec{Type: "boolean"}
	case "true":
		// Telegram Bot API special pseudo-type "True" means a boolean literal true
		return &openapi.TypeSpec{Type: "boolean", Default: true}
	}

	name := strings.TrimSpace(raw)

	return &openapi.TypeSpec{Ref: &openapi.TypeRef{Name: name}}
}

// ContainsType reports whether the raw type (including nested arrays/unions)
// references the provided type name (case-insensitive exact match).
func (t *TypeRef) ContainsType(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}

	targetLower := strings.ToLower(target)

	return t.contains(func(raw string) bool {
		return strings.EqualFold(raw, target) || strings.ToLower(raw) == targetLower
	})
}

// ContainsTypeWithPrefix reports whether any referenced type name begins with
// the provided prefix (case-insensitive).
func (t *TypeRef) ContainsTypeWithPrefix(prefix string) bool {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return false
	}

	prefixLower := strings.ToLower(prefix)

	return t.contains(func(raw string) bool {
		return strings.HasPrefix(strings.ToLower(raw), prefixLower)
	})
}

func (t *TypeRef) contains(predicate func(string) bool) bool {
	if t == nil {
		return false
	}

	raw := strings.TrimSpace(t.RawType)
	if raw == "" {
		return false
	}

	// Handle arrays like "Array of X" (possibly nested)
	const arrayPrefix = "Array of "

	lowerRaw := strings.ToLower(raw)
	if strings.HasPrefix(lowerRaw, strings.ToLower(arrayPrefix)) {
		inner := strings.TrimSpace(raw[len(arrayPrefix):])

		return NewTypeRef(inner).contains(predicate)
	}

	// Handle union types via UnionParts
	if parts := t.UnionParts(); parts != nil {
		for _, part := range parts {
			if NewTypeRef(part).contains(predicate) {
				return true
			}
		}

		return false
	}

	return predicate(raw)
}

// TypeDef captures a Telegram object/type definition extracted from the docs.
type TypeDef struct {
	Anchor      string
	Name        string
	Tag         string
	Description []string
	Notes       []string
	Fields      []TypeFieldDef
}

// TypeFieldDef describes an individual field inside a Telegram object schema.
type TypeFieldDef struct {
	Name        string
	TypeRef     *TypeRef
	Required    bool
	Description string
}

// ParseType parses a Telegram type definition starting at the provided anchor
// and returns the structured representation.
//
//nolint:cyclop,funlen // DOM parsing needs branching
func ParseType(doc *goquery.Document, anchor string) (*TypeDef, error) {
	res := &TypeDef{
		Anchor: anchor,
	}

	header := doc.Find("h4").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is(fmt.Sprintf("a.anchor[Name='%s']", anchor))
	})
	if header.Length() == 0 {
		return nil, ErrElementNotFound
	}

	le := header.First()
	res.Name = strings.TrimSpace(le.Text())
	// Determine the tag as the nearest preceding h3 title
	if prevH3 := le.PrevAll().Filter("h3").First(); prevH3.Length() > 0 {
		res.Tag = strings.TrimSpace(prevH3.Text())
	}

	// Limit our search to the section between this header and the next h4
	section := le.NextUntil("h4")

	// Walk siblings preserving order until the first table (fields)
	for sibling := le.Next(); sibling.Length() > 0; sibling = sibling.Next() {
		nodeName := goquery.NodeName(sibling)
		if nodeName == "h4" {
			break
		}

		if nodeName == "table" {
			break
		}

		switch nodeName {
		case "p":
			text := strings.TrimSpace(sibling.Text())
			if text != "" {
				res.Description = append(res.Description, text)
			}
		case "ul", "ol":
			sibling.Find("li").Each(func(i int, li *goquery.Selection) {
				text := strings.TrimSpace(li.Text())
				if text != "" {
					res.Description = append(res.Description, text)
				}
			})
		}
	}

	// Parse fields from the first table in the section
	section.Find("table tbody tr").Each(func(index int, tr *goquery.Selection) {
		fieldDef := TypeFieldDef{}

		tr.Find("td").Each(func(tdIndex int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())

			switch tdIndex {
			case fieldNameColumnIndex:
				fieldDef.Name = text
			case fieldTypeColumnIndex:
				fieldDef.TypeRef = NewTypeRef(text)
			case fieldDescriptionColumnIndex:
				fieldDef.Description = text
			}
		})

		if fieldDef.Name == "" {
			return
		}
		// Force chat_id and user_id to be Integer64 regardless of parsed union or other forms.
		// Priority for chat_id is based on its name.
		// Also force any field mentioning "64-bit integer" in description to be Integer64.
		if fieldDef.Name == "chat_id" || strings.HasSuffix(fieldDef.Name, "_chat_id") ||
			fieldDef.Name == "user_id" || strings.HasSuffix(fieldDef.Name, "_user_id") ||
			strings.Contains(strings.ToLower(fieldDef.Description), "64-bit integer") {
			fieldDef.TypeRef = NewTypeRef("Integer64")
		}

		fieldDef.Required = !isOptionalDescription(fieldDef.Description)
		res.Fields = append(res.Fields, fieldDef)
	})

	// Parse notes inside blockquotes in the section
	section.Find("blockquote p").Each(func(index int, p *goquery.Selection) {
		text := strings.TrimSpace(p.Text())
		if text != "" {
			res.Notes = append(res.Notes, text)
		}
	})

	return res, nil
}

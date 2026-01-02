package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

const (
	paramTypeColumnIndex        = 1
	paramOptionalColumnIndex    = 2
	paramDescriptionColumnIndex = 3
	initialReturnTypesCapacity  = 2
)

// MethodDef holds the structured information extracted for a Telegram API method.
type MethodDef struct {
	Anchor      string
	Name        string
	Tags        []string
	Description []string
	Notes       []string
	Params      map[string]MethodParamDef
	Return      *TypeRef
}

// MethodParamDef describes a single parameter in a Telegram API method table.
type MethodParamDef struct {
	Name        string
	TypeRef     *TypeRef
	Required    bool
	Description string
}

// ParseMethod walks the documentation rooted at the provided anchor and
// produces a MethodDef describing the Telegram API method.
//
//nolint:cyclop,funlen // doc traversal needs multiple branches
func ParseMethod(doc *goquery.Document, anchor string) (*MethodDef, error) {
	res := &MethodDef{
		Anchor: anchor,
		Params: make(map[string]MethodParamDef),
	}

	el := doc.Find("h4").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is(fmt.Sprintf("a.anchor[Name='%s']", anchor))
	})
	// header with anchor not found
	if el.Length() == 0 {
		return nil, fmt.Errorf("parse method at anchor %s: %w", anchor, ErrElementNotFound)
	}

	res.Name = el.Text()
	// Determine the tag as the nearest preceding h3 title
	if prevH3 := el.PrevAll().Filter("h3").First(); prevH3.Length() > 0 {
		res.Tags = []string{strings.TrimSpace(prevH3.Text())}
	}

	// Traverse siblings starting after the header until the next h4,
	// collecting all description paragraphs (even those that appear after tables).
	// We skip tables but do not stop at them, since return descriptions might
	// be placed after the parameters table in some methods.
	sib := el.Next()
	for sib.Length() > 0 {
		if sib.IsMatcher(goquery.Single("h4")) {
			break
		}
		// Skip over tables but keep scanning further siblings
		if sib.IsMatcher(goquery.Single("table")) {
			sib = sib.Next()

			continue
		}

		if sib.IsMatcher(goquery.Single("p")) {
			res.Description = append(res.Description, strings.TrimSpace(sib.Text()))
		}

		sib = sib.Next()
	}

	// Try to extract a return type from the description paragraphs
	rt := extractReturnType(res.Description)
	if strings.TrimSpace(rt) == "" {
		return nil, fmt.Errorf("parse method %s (%s): %w", res.Name, anchor, ErrReturnTypeNotParsed)
	}

	res.Return = NewTypeRef(rt)

	// Limit our search to the section between this header and the next h4
	section := el.NextUntil("h4")

	section.Find("table tbody tr").Each(func(index int, tr *goquery.Selection) {
		name := ""
		def := MethodParamDef{}

		var optionalValue string

		tr.Find("td").Each(func(tdIndex int, td *goquery.Selection) {
			switch tdIndex {
			case 0:
				name = td.Text()
			case paramTypeColumnIndex:
				def.TypeRef = NewTypeRef(td.Text())
			case paramOptionalColumnIndex:
				optionalValue = strings.TrimSpace(td.Text())
			case paramDescriptionColumnIndex:
				def.Description = td.Text()
			}
		})

		// Determine required based on "Required" column OR description starting with Optional
		def.Required = !isOptionalDescription(def.Description) && !strings.EqualFold(optionalValue, "Optional")

		// Force chat_id to be Integer for method parameters as well
		if name == "chat_id" || strings.HasSuffix(name, "_chat_id") {
			def.TypeRef = NewTypeRef("Integer")
		}

		res.Params[name] = def
	})

	section.Find("blockquote p").Each(func(index int, p *goquery.Selection) {
		res.Notes = append(res.Notes, p.Text())
	})

	return res, nil
}

// extractReturnType scans description paragraphs to detect a return type phrase
// like "Returns X on success." or "On success, X is returned." and returns the
// extracted type string, or empty if not found.
func extractReturnType(paragraphs []string) string {
	for _, p := range paragraphs {
		text := strings.TrimSpace(p)
		if text == "" {
			continue
		}

		sentences := splitIntoSentences(text)
		for _, sentence := range sentences {
			if candidate := extractReturnTypeFromSentence(sentence); candidate != "" {
				return candidate
			}
		}
	}

	return ""
}

func splitIntoSentences(text string) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}

	var (
		sentences []string
		current   strings.Builder
	)
	for _, r := range trimmed {
		current.WriteRune(r)

		switch r {
		case '.', '!', '?':
			if s := strings.TrimSpace(current.String()); s != "" {
				sentences = append(sentences, s)
			}

			current.Reset()
		}
	}

	if tail := strings.TrimSpace(current.String()); tail != "" {
		sentences = append(sentences, tail)
	}

	return sentences
}

//nolint:cyclop,nestif // parsing return phrasing requires branching
func extractReturnTypeFromSentence(sentence string) string {
	text := strings.TrimSpace(sentence)
	if text == "" {
		return ""
	}

	lower := strings.ToLower(text)
	if !strings.Contains(lower, "return") {
		return ""
	}

	if idx := strings.Index(lower, "returns "); idx != -1 {
		remainder := text[idx+len("Returns "):]
		if candidate := parseReturnsClause(remainder); candidate != "" {
			return candidate
		}
	}

	if idx := strings.Index(lower, "on success"); idx != -1 {
		remainder := strings.TrimSpace(text[idx+len("on success"):])

		remainder = strings.TrimLeft(remainder, ", ")
		if remainder != "" {
			remainderLower := strings.ToLower(remainder)
			if strings.HasPrefix(remainderLower, "returns ") {
				if candidate := parseReturnsClause(remainder[len("returns "):]); candidate != "" {
					return candidate
				}
			}

			if candidate := parseIsReturnedClause(remainder); candidate != "" {
				return candidate
			}
		}
	}

	if candidate := parseIsReturnedClause(text); candidate != "" {
		return candidate
	}

	return ""
}

//nolint:cyclop // phrase trimming needs multiple delimiters
func parseReturnsClause(remainder string) string {
	trimmed := strings.TrimSpace(remainder)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)

	stop := len(trimmed)
	for _, marker := range []string{" on success", " upon success", " if ", " when ", " otherwise ", ", if", ", when"} {
		if idx := strings.Index(lower, marker); idx != -1 && idx < stop {
			stop = idx
		}
	}

	if idx := strings.IndexAny(trimmed, ".!?;"); idx != -1 && idx < stop {
		stop = idx
	}

	if idx := strings.Index(trimmed, "("); idx != -1 && idx < stop {
		stop = idx
	}

	if idx := strings.Index(trimmed, "["); idx != -1 && idx < stop {
		stop = idx
	}

	if idx := strings.Index(trimmed, ","); idx != -1 && idx < stop {
		stop = idx
	}

	candidate := strings.TrimSpace(trimmed[:stop])
	candidate = strings.TrimSuffix(candidate, ".")
	candidate = strings.TrimSpace(candidate)

	candidate = normalizeReturnTypePhrase(candidate)
	if looksLikeReturnType(candidate) {
		return candidate
	}

	return ""
}

//nolint:cyclop,funlen // handles numerous conjunction patterns
func parseIsReturnedClause(text string) string {
	lower := strings.ToLower(text)
	stopIdx := strings.Index(lower, " is returned")
	keyword := " is returned"

	if stopIdx == -1 {
		stopIdx = strings.Index(lower, " are returned")
		keyword = " are returned"
	}

	if stopIdx == -1 {
		return ""
	}

	prefix := strings.TrimSpace(text[:stopIdx])
	if prefix == "" {
		return ""
	}

	if comma := strings.LastIndex(prefix, ","); comma != -1 {
		prefix = strings.TrimSpace(prefix[comma+1:])
	}

	primary := normalizeReturnTypePhrase(prefix)

	types := make([]string, 0, initialReturnTypesCapacity)
	if looksLikeReturnType(primary) {
		types = append(types, primary)
	}

	rest := text[stopIdx+len(keyword):]

	lowerRest := strings.ToLower(rest)
	if idx := strings.Index(lowerRest, "otherwise "); idx != -1 {
		alt := strings.TrimSpace(rest[idx+len("otherwise "):])
		if strings.HasPrefix(strings.ToLower(alt), "returns ") {
			if val := parseReturnsClause(alt[len("returns "):]); val != "" {
				types = append(types, val)
			}
		} else if val := parseIsReturnedClause(alt); val != "" {
			types = append(types, val)
		}
	}

	if idx := strings.Index(lowerRest, " or "); idx != -1 {
		alt := strings.TrimSpace(rest[idx+len(" or "):])
		if strings.HasPrefix(strings.ToLower(alt), "returns ") {
			if val := parseReturnsClause(alt[len("returns "):]); val != "" {
				types = append(types, val)
			}
		} else if val := parseIsReturnedClause(alt); val != "" {
			types = append(types, val)
		}
	}

	if len(types) == 0 {
		return ""
	}

	seen := make(map[string]struct{}, len(types))

	unique := make([]string, 0, len(types))
	for _, t := range types {
		if _, ok := seen[t]; ok {
			continue
		}

		seen[t] = struct{}{}
		unique = append(unique, t)
	}

	if len(unique) == 1 {
		return unique[0]
	}

	return strings.Join(unique, " or ")
}

func looksLikeReturnType(s string) bool {
	if s == "" {
		return false
	}

	ls := strings.ToLower(s)
	switch ls {
	case "true", "false", "string", "integer", "int", "float", "float number", "number", "boolean", "bool":
		return true
	}

	if strings.HasPrefix(s, "Array of ") || strings.HasPrefix(s, "array of ") {
		return true
	}

	if r, _ := utf8.DecodeRuneInString(s); r != utf8.RuneError && unicode.IsUpper(r) {
		return true
	}

	return false
}

// normalizeReturnTypePhrase cleans up common wording around return type phrases
// from Telegram docs to align with our TypeRef parser expectations.
//
//nolint:cyclop,funlen,gocognit // normalization handles many doc edge cases
func normalizeReturnTypePhrase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	leadingPrefixes := []string{
		"the sent ",
		"the uploaded ",
		"the created ",
		"the edited ",
		"the revoked ",
		"the new ",
		"the ",
		"an ",
		"a ",
		"stopped ",
	}
	stripLeading := func() {
		for s != "" {
			trimmed := false

			ls := strings.ToLower(s)
			for _, pref := range leadingPrefixes {
				if strings.HasPrefix(ls, pref) {
					s = strings.TrimSpace(s[len(pref):])
					trimmed = true

					break
				}
			}

			if !trimmed {
				break
			}
		}
	}
	stripLeading()

	linkPrefixes := []string{
		"created invoice link as ",
		"invoice link as ",
		"new invite link as ",
		"edited invite link as ",
		"revoked invite link as ",
		"invite link as ",
	}

	for {
		ls := strings.ToLower(s)
		trimmed := false

		for _, pref := range linkPrefixes {
			if strings.HasPrefix(ls, pref) {
				s = strings.TrimSpace(s[len(pref):])
				trimmed = true

				break
			}
		}

		if !trimmed {
			break
		}

		stripLeading()
	}

	ls := strings.ToLower(s)
	if strings.Contains(ls, "information about ") {
		if idx := strings.Index(ls, " as a "); idx != -1 {
			s = strings.TrimSpace(s[idx+len(" as a "):])

			stripLeading()

			ls = strings.ToLower(s)
		} else if idx := strings.Index(ls, " as an "); idx != -1 {
			s = strings.TrimSpace(s[idx+len(" as an "):])

			stripLeading()

			ls = strings.ToLower(s)
		}
	}

	if idx := strings.Index(ls, " in form of a "); idx != -1 {
		s = strings.TrimSpace(s[idx+len(" in form of a "):])

		stripLeading()

		ls = strings.ToLower(s)
	}

	if idx := strings.Index(ls, " in form of an "); idx != -1 {
		s = strings.TrimSpace(s[idx+len(" in form of an "):])

		stripLeading()

		ls = strings.ToLower(s)
	}

	if idx := strings.Index(ls, " in the form of a "); idx != -1 {
		s = strings.TrimSpace(s[idx+len(" in the form of a "):])

		stripLeading()

		ls = strings.ToLower(s)
	}

	if idx := strings.Index(ls, " in the form of an "); idx != -1 {
		s = strings.TrimSpace(s[idx+len(" in the form of an "):])

		stripLeading()

		ls = strings.ToLower(s)
	}

	if idx := strings.LastIndex(ls, " as "); idx != -1 {
		before := strings.TrimSpace(ls[:idx])
		if strings.HasSuffix(before, " link") {
			s = strings.TrimSpace(s[idx+len(" as "):])

			stripLeading()

			ls = strings.ToLower(s)
		}
	}
	// Normalize array phrasing
	if strings.HasPrefix(ls, "array of ") {
		s = "Array of " + strings.TrimSpace(s[len("array of "):])

		stripLeading()
	}
	// Drop trailing generic words like "object(s)" when preceded by a type name
	suffixes := []string{
		" of the sent messages",
		" of the sent message",
		" of sent messages",
		" of sent message",
		" that were sent",
		" that was sent",
		" of the message",
		" of messages",
		" objects",
		" object",
		" stories",
		" story",
	}

	for {
		trimmed := false

		ls = strings.ToLower(s)
		for _, suf := range suffixes {
			if strings.HasSuffix(ls, suf) {
				s = strings.TrimSpace(s[:len(s)-len(suf)])
				trimmed = true

				break
			}
		}

		if !trimmed {
			break
		}
	}
	// Capitalize first letter of common scalar names when found alone
	switch ls {
	case "string":
		return "String"
	case "integer":
		return "Integer"
	case "boolean":
		return "Boolean"
	case "true":
		return "True"
	}

	return s
}

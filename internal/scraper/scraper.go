package scraper

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/metalagman/tgbotspec/internal/fetcher"
	"github.com/metalagman/tgbotspec/internal/openapi"
	"github.com/metalagman/tgbotspec/internal/parser"
)

var fetchDocument = fetcher.Document

// Options configures the scraper behavior.
type Options struct {
	MergeUnionTypes bool
}

// Run orchestrates fetching the Telegram Bot API docs, parsing them, and
// rendering the OpenAPI specification to the provided writer.
func Run(w io.Writer, opts Options) error { //nolint:cyclop,funlen,gocognit
	doc, err := fetchDocument()
	if err != nil {
		return fmt.Errorf("fetch document: %w", err)
	}

	apiVersion := extractBotAPIVersion(doc)
	if apiVersion == "" {
		apiVersion = "0.0.0"
	}

	title := extractAPITitle(doc)
	if title == "" {
		title = "Telegram Bot API"
	}

	log.Printf("scraper: detected Telegram Bot API title %q, version %s", title, apiVersion)

	typeTargets, methodTargets := splitTargets(parser.ParseNavLists(doc), doc)

	// Pass 1: Create a map of all types for lookup during merging
	typesMap := make(map[string]parser.TypeDef, len(typeTargets))
	for _, t := range typeTargets {
		typesMap[t.Name] = t
	}

	renderData := openapi.TemplateData{
		Title:   title,
		Version: apiVersion,
	}

	// Pre-populate valid types for union merging validation
	validTypes := make(map[string]struct{}, len(typeTargets))
	for name := range typesMap {
		validTypes[name] = struct{}{}
	}

	validTypes["ResponseParameters"] = struct{}{}

	seenTypes := make(map[string]struct{}, len(typeTargets))

	for _, t := range typeTargets {
		if _, exists := seenTypes[t.Name]; exists {
			continue
		}

		if t.Name == "InputFile" {
			continue
		}

		seenTypes[t.Name] = struct{}{}

		spec := openapi.Type{
			Name:        t.Name,
			Tag:         t.Tag,
			Description: t.Description,
		}
		for _, field := range t.Fields {
			s := field.TypeRef.ToTypeSpec()
			if opts.MergeUnionTypes {
				s = mergeUnionTypes(s, validTypes, typesMap)
			}

			spec.Fields = append(spec.Fields, openapi.TypeField{
				Name:        field.Name,
				Description: field.Description,
				Required:    field.Required,
				Schema:      s.WithDescription(field.Description),
			})
		}

		renderData.Types = append(renderData.Types, spec)
	}

	if _, ok := seenTypes["ResponseParameters"]; !ok {
		renderData.Types = append(renderData.Types, openapi.Type{
			Name: "ResponseParameters",
			Fields: []openapi.TypeField{
				{
					Name:   "migrate_to_chat_id",
					Schema: &openapi.TypeSpec{Type: "integer"},
				},
				{
					Name:   "retry_after",
					Schema: &openapi.TypeSpec{Type: "integer"},
				},
			},
		})
		seenTypes["ResponseParameters"] = struct{}{}
	}

	sort.Slice(renderData.Types, func(i, j int) bool {
		return renderData.Types[i].Name < renderData.Types[j].Name
	})

	for _, m := range methodTargets {
		method := openapi.Method{
			Name:              m.Name,
			Tags:              m.Tags,
			Description:       m.Description,
			SupportsMultipart: false,
		}
		if m.Return != nil {
			method.Return = m.Return.ToTypeSpec()
			if opts.MergeUnionTypes {
				method.Return = mergeUnionTypes(method.Return, validTypes, typesMap)
			}
		}

		paramNames := make([]string, 0, len(m.Params))
		for name := range m.Params {
			paramNames = append(paramNames, name)
		}

		sort.Strings(paramNames)

		for _, name := range paramNames {
			param := m.Params[name]

			s := param.TypeRef.ToTypeSpec()
			if opts.MergeUnionTypes {
				s = mergeUnionTypes(s, validTypes, typesMap)
			}

			method.Params = append(method.Params, openapi.MethodParam{
				Name:        name,
				Description: param.Description,
				Required:    param.Required,
				Schema:      s.WithDescription(param.Description),
			})

			if requiresMultipart(param.TypeRef) {
				method.SupportsMultipart = true
			}
		}

		renderData.Methods = append(renderData.Methods, method)
	}

	if err := openapi.RenderTemplate(w, &renderData); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return nil
}

func mergeUnionTypes(
	spec *openapi.TypeSpec,
	validTypes map[string]struct{},
	typesMap map[string]parser.TypeDef,
) *openapi.TypeSpec {
	if spec == nil {
		return nil
	}

	// Recursive check for Array items
	if spec.Type == "array" && spec.Items != nil {
		spec.Items = mergeUnionTypes(spec.Items, validTypes, typesMap)

		return spec
	}

	elements := spec.AnyOf
	isAnyOf := true

	if len(elements) == 0 {
		elements = spec.OneOf
		isAnyOf = false
	}

	if len(elements) == 0 {
		return spec
	}

	var (
		refs     []openapi.TypeSpec
		others   []openapi.TypeSpec
		refNames []string
	)

	for _, el := range elements {
		if el.Ref != nil && el.Ref.Name != "" {
			refs = append(refs, el)
			refNames = append(refNames, el.Ref.Name)
		} else {
			others = append(others, el)
		}
	}

	// We need at least 2 refs to merge anything meaningfully,
	// unless we want to "upcast" a single ref? No, MinUnionParts=2 usually.
	if len(refs) < parser.MinUnionParts {
		return spec
	}

	// Try merging by common prefix
	if merged := tryMergeByPrefix(spec, refs, others, refNames, isAnyOf, validTypes); merged != nil {
		return merged
	}

	// If it's all refs but no common prefix was found, merge them into a single object with all properties.
	if len(others) == 0 {
		return mergeProperties(refs, typesMap)
	}

	return spec
}

func tryMergeByPrefix(
	spec *openapi.TypeSpec,
	refs, others []openapi.TypeSpec,
	refNames []string,
	isAnyOf bool,
	validTypes map[string]struct{},
) *openapi.TypeSpec {
	common := commonPrefix(refNames)
	if common == "" {
		return nil
	}

	if _, ok := validTypes[common]; !ok {
		return nil
	}

	mergedRef := openapi.TypeSpec{
		Ref: &openapi.TypeRef{Name: common},
	}

	if len(others) == 0 {
		return &mergedRef
	}

	newElements := append([]openapi.TypeSpec{mergedRef}, others...)

	newSpec := *spec
	if isAnyOf {
		newSpec.AnyOf = newElements
		newSpec.OneOf = nil
	} else {
		newSpec.OneOf = newElements
		newSpec.AnyOf = nil
	}

	return &newSpec
}

func mergeProperties(refs []openapi.TypeSpec, typesMap map[string]parser.TypeDef) *openapi.TypeSpec {
	merged := &openapi.TypeSpec{
		Type:       "object",
		Properties: make(map[string]openapi.TypeSpec),
	}

	for _, ref := range refs {
		if td, ok := typesMap[ref.Ref.Name]; ok {
			for _, field := range td.Fields {
				merged.Properties[field.Name] = *field.TypeRef.ToTypeSpec().WithDescription(field.Description)
			}
		}
	}

	if len(merged.Properties) > 0 {
		return merged
	}

	return &openapi.TypeSpec{Type: "object"}
}

func commonPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}

	prefix := names[0]
	for _, name := range names[1:] {
		for !strings.HasPrefix(name, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}

	return prefix
}

func splitTargets(targets []parser.ParseTarget, doc *goquery.Document) ([]parser.TypeDef, []parser.MethodDef) {
	if len(targets) == 0 {
		sections := []string{
			"getting-updates",
			"available-types",
			"available-methods",
			"updating-messages",
			"stickers",
			"inline-mode",
			"payments",
			"telegram-passport",
			"games",
		}
		targets = parser.ParseAllNavs(doc, sections)
	}

	var (
		typeTargets   []parser.TypeDef
		methodTargets []parser.MethodDef
	)

	for _, target := range targets {
		switch target.Mode {
		case parser.ParseModeType:
			td, err := parser.ParseType(doc, target.Anchor)
			if err != nil {
				log.Printf("scraper: parse type %q failed: %v", target.Anchor, err)

				continue
			}

			typeTargets = append(typeTargets, *td)
		case parser.ParseModeMethod:
			md, err := parser.ParseMethod(doc, target.Anchor)
			if err != nil {
				log.Printf("scraper: parse method %q failed: %v", target.Anchor, err)

				continue
			}

			methodTargets = append(methodTargets, *md)
		}
	}

	sort.Slice(typeTargets, func(i, j int) bool {
		return typeTargets[i].Name < typeTargets[j].Name
	})

	sort.Slice(methodTargets, func(i, j int) bool {
		return methodTargets[i].Name < methodTargets[j].Name
	})

	// Ensure method tag lists are sorted for deterministic grouping
	for i := range methodTargets {
		if len(methodTargets[i].Tags) > 1 {
			sort.Strings(methodTargets[i].Tags)
		}
	}

	return typeTargets, methodTargets
}

func extractBotAPIVersion(doc *goquery.Document) string {
	sel := doc.Find("p strong").FilterFunction(func(i int, s *goquery.Selection) bool {
		return strings.HasPrefix(strings.TrimSpace(s.Text()), "Bot API ")
	}).First()

	text := strings.TrimSpace(sel.Text())
	if text == "" {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(text, "Bot API "))
}

func extractAPITitle(doc *goquery.Document) string {
	h1 := strings.TrimSpace(doc.Find("h1").First().Text())
	if h1 != "" {
		return h1
	}

	return strings.TrimSpace(doc.Find("title").First().Text())
}

func requiresMultipart(tr *parser.TypeRef) bool {
	if tr == nil {
		return false
	}

	if tr.ContainsType("InputFile") {
		return true
	}

	if tr.ContainsTypeWithPrefix("InputMedia") {
		return true
	}

	if tr.ContainsTypeWithPrefix("InputSticker") {
		return true
	}

	return false
}

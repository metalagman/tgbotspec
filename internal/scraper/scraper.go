package scraper

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"tgbotspec/internal/fetcher"
	"tgbotspec/internal/openapi"
	"tgbotspec/internal/parser"
)

var fetchDocument = fetcher.Document

// Run orchestrates fetching the Telegram Bot API docs, parsing them, and
// rendering the OpenAPI specification to the provided writer.
//
//nolint:cyclop,funlen // orchestration covers multiple branches
func Run(w io.Writer) error {
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

	renderData := openapi.TemplateData{
		Title:   title,
		Version: apiVersion,
	}

	seenTypes := make(map[string]struct{}, len(typeTargets))

	for _, t := range typeTargets {
		if _, exists := seenTypes[t.Name]; exists {
			continue
		}

		seenTypes[t.Name] = struct{}{}

		spec := openapi.Type{
			Name:        t.Name,
			Tag:         t.Tag,
			Description: t.Description,
		}
		for _, field := range t.Fields {
			spec.Fields = append(spec.Fields, openapi.TypeField{
				Name:        field.Name,
				Description: field.Description,
				Required:    field.Required,
				Schema:      field.TypeRef.ToTypeSpec(),
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
			Name:        m.Name,
			Tags:        m.Tags,
			Description: m.Description,
		}
		if m.Return != nil {
			method.Return = m.Return.ToTypeSpec()
		}

		paramNames := make([]string, 0, len(m.Params))
		for name := range m.Params {
			paramNames = append(paramNames, name)
		}

		sort.Strings(paramNames)

		for _, name := range paramNames {
			param := m.Params[name]
			method.Params = append(method.Params, openapi.MethodParam{
				Name:        name,
				Description: param.Description,
				Required:    param.Required,
				Schema:      param.TypeRef.ToTypeSpec(),
			})
		}

		renderData.Methods = append(renderData.Methods, method)
	}

	if err := openapi.RenderTemplate(w, &renderData); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return nil
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

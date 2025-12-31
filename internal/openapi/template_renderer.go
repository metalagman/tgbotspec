package openapi

import (
	_ "embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/template"
)

//go:embed openapi.yaml.gotmpl
var openapiTemplate []byte

var (
	templateOnce sync.Once
	tmpl         *template.Template
	tmplErr      error
)

// RenderTemplate executes the embedded OpenAPI template against the provided
// data and writes the result to the given writer.
func RenderTemplate(w io.Writer, data *TemplateData) error {
	if data == nil {
		return nil
	}

	t, err := parsedTemplate()
	if err != nil {
		return err
	}

	return t.Execute(w, data)
}

func parsedTemplate() (*template.Template, error) {
	templateOnce.Do(func() {
		if len(openapiTemplate) == 0 {
			tmplErr = fmt.Errorf("openapi template is empty")

			return
		}

		tmpl, tmplErr = template.New("openapi.yaml.gotmpl").Funcs(template.FuncMap{
			"renderSchema":          renderSchema,
			"renderJSONSchema":      renderJSONSchema,
			"renderMultipartSchema": renderMultipartSchema,
			"indent":                indent,
			"isBinary":              isBinary,
			"isNotBinary":           isNotBinary,
			"isPureBinary":          isPureBinary,
		}).Parse(string(openapiTemplate))
	})

	return tmpl, tmplErr
}

func isPureBinary(spec *TypeSpec) bool {
	if spec == nil {
		return false
	}

	return spec.Format == "binary"
}

func isBinary(spec *TypeSpec) bool {
	if spec == nil {
		return false
	}

	if spec.Format == "binary" {
		return true
	}

	for i := range spec.AnyOf {
		if isBinary(&spec.AnyOf[i]) {
			return true
		}
	}

	for i := range spec.OneOf {
		if isBinary(&spec.OneOf[i]) {
			return true
		}
	}

	if spec.Items != nil {
		return isBinary(spec.Items)
	}

	return false
}

func isNotBinary(spec *TypeSpec) bool {
	return !isBinary(spec)
}

func renderJSONSchema(spec *TypeSpec) (string, error) {
	return RenderTypeSpecToYAML(simplifyJSON(spec))
}

func renderMultipartSchema(spec *TypeSpec) (string, error) {
	return RenderTypeSpecToYAML(simplifyMultipart(spec))
}

func simplifyJSON(spec *TypeSpec) *TypeSpec {
	if spec == nil {
		return nil
	}

	if spec.Format == "binary" {
		return nil
	}

	res := *spec

	if len(spec.AnyOf) > 0 {
		var anyOf []TypeSpec

		for i := range spec.AnyOf {
			if s := simplifyJSON(&spec.AnyOf[i]); s != nil {
				anyOf = append(anyOf, *s)
			}
		}

		if len(anyOf) == 1 {
			return &anyOf[0]
		}

		res.AnyOf = anyOf
	}

	if spec.Items != nil {
		res.Items = simplifyJSON(spec.Items)
	}

	return &res
}

func simplifyMultipart(spec *TypeSpec) *TypeSpec {
	if spec == nil {
		return nil
	}

	if isBinary(spec) {
		// If it's an array, we keep it an array but simplify items
		if spec.Type == "array" && spec.Items != nil {
			res := *spec
			res.Items = simplifyMultipart(spec.Items)

			return &res
		}

		return &TypeSpec{
			Type:        "string",
			Format:      "binary",
			Description: spec.Description,
		}
	}

	return spec
}

func renderSchema(spec *TypeSpec) (string, error) {
	if spec == nil {
		return "", nil
	}

	return RenderTypeSpecToYAML(spec)
}

func indent(n int, s string) string {
	if s == "" {
		return s
	}

	pad := strings.Repeat(" ", n)

	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}

		lines[i] = pad + ln
	}

	return strings.Join(lines, "\n")
}

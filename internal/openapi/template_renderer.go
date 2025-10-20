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
			"renderSchema": renderSchema,
			"indent":       indent,
		}).Parse(string(openapiTemplate))
	})
	return tmpl, tmplErr
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

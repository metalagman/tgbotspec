package openapi //nolint:testpackage // tests rely on constructing internal types directly

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestRenderTypeSpecToYAML(t *testing.T) { //nolint:funlen // table-driven cases cover many shapes
	testCases := []struct {
		name         string
		spec         TypeSpec
		expectedYAML string
	}{
		{
			name: "ScalarType",
			spec: TypeSpec{
				Type: "integer",
			},
			expectedYAML: "type: integer\n",
		},
		{
			name: "ObjectType",
			spec: TypeSpec{
				Type: "object",
				Properties: map[string]TypeSpec{
					"id": {
						Type:   "integer",
						Format: "int64",
					},
					"name": {
						Type: "string",
					},
				},
				Required: []string{"id", "name"},
			},
			expectedYAML: "type: object\n" +
				"properties:\n  id:\n    format: int64\n    type: integer\n  name:\n    type: string\n" +
				"required:\n- id\n- name\n",
		},
		{
			name: "ArrayScalarType",
			spec: TypeSpec{
				Type: "array",
				Items: &TypeSpec{
					Type: "integer",
				},
			},
			expectedYAML: "type: array\nitems:\n  type: integer\n",
		},
		{
			name: "ArrayOfObjectType",
			spec: TypeSpec{
				Type: "array",
				Items: &TypeSpec{
					Type: "object",
					Properties: map[string]TypeSpec{
						"id": {
							Type:   "integer",
							Format: "int64",
						},
						"name": {
							Type: "string",
						},
					},
					Required: []string{"id", "name"},
				},
			},
			expectedYAML: "type: array\n" +
				"items:\n" +
				"  type: object\n" +
				"  properties:\n" +
				"    id:\n" +
				"      format: int64\n" +
				"      type: integer\n" +
				"    name:\n" +
				"      type: string\n" +
				"  required:\n" +
				"  - id\n" +
				"  - name\n",
		},
		{
			name: "ArrayOfArrayOfObjectType",
			spec: TypeSpec{
				Type: "array",
				Items: &TypeSpec{
					Type: "array",
					Items: &TypeSpec{
						Type: "object",
						Properties: map[string]TypeSpec{
							"id": {
								Type:   "integer",
								Format: "int64",
							},
							"name": {
								Type: "string",
							},
						},
						Required: []string{"id", "name"},
					},
				},
			},
			expectedYAML: "type: array\n" +
				"items:\n" +
				"  type: array\n" +
				"  items:\n" +
				"    type: object\n" +
				"    properties:\n" +
				"      id:\n" +
				"        format: int64\n" +
				"        type: integer\n" +
				"      name:\n" +
				"        type: string\n" +
				"    required:\n" +
				"    - id\n" +
				"    - name\n",
		},
		{
			name: "AnyOfType",
			spec: TypeSpec{
				AnyOf: []TypeSpec{
					{
						Type: "string",
					},
					{
						Type: "integer",
					},
				},
			},
			expectedYAML: "anyOf:\n- type: string\n- type: integer\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			yamlStr, err := RenderTypeSpecToYAML(&tc.spec)
			if err != nil {
				t.Errorf("Error rendering to YAML: %v", err)

				return
			}

			// Convert expected YAML to JSON
			expectedJSON, err := yaml.YAMLToJSON([]byte(tc.expectedYAML))
			if err != nil {
				t.Errorf("Error converting expected YAML to JSON: %v", err)

				return
			}

			// Convert actual YAML to JSON
			actualJSON, err := yaml.YAMLToJSON([]byte(yamlStr))
			if err != nil {
				t.Errorf("Error converting actual YAML to JSON: %v", err)

				return
			}

			// Compare JSON strings with relaxed equality
			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestRenderTemplateNilData(t *testing.T) {
	var buf bytes.Buffer

	if err := RenderTemplate(&buf, nil); err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("expected no output for nil data, got %q", buf.String())
	}
}

func TestRenderTemplateEmptyTemplate(t *testing.T) {
	originalTemplate := append([]byte(nil), openapiTemplate...)

	openapiTemplate = nil
	templateOnce = sync.Once{}
	tmpl = nil
	tmplErr = nil

	t.Cleanup(func() {
		openapiTemplate = originalTemplate
		templateOnce = sync.Once{}
		tmpl = nil
		tmplErr = nil
	})

	var buf bytes.Buffer

	err := RenderTemplate(&buf, &TemplateData{Title: "Test", Version: "1.0.0"})
	if err == nil {
		t.Fatal("expected error when template is empty")
	}
}

func TestIndent(t *testing.T) {
	input := "line1\n\nline2"
	got := indent(2, input)
	expected := "  line1\n\n  line2"

	if got != expected {
		t.Fatalf("indent result mismatch: got %q want %q", got, expected)
	}

	if indent(4, "") != "" {
		t.Fatal("expected empty string to remain empty")
	}

	if indent(2, strings.Repeat(" ", 3)) != strings.Repeat(" ", 3) {
		t.Fatal("expected whitespace-only lines to remain unchanged")
	}
}

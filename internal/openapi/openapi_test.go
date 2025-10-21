package openapi

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestRenderTypeSpecToYAML(t *testing.T) {
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
			expectedYAML: "type: object\nproperties:\n  id:\n    format: int64\n    type: integer\n  name:\n    type: string\nrequired:\n- id\n- name\n",
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
			expectedYAML: "type: array\nitems:\n  type: object\n  properties:\n    id:\n      format: int64\n      type: integer\n    name:\n      type: string\n  required:\n  - id\n  - name\n",
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
			expectedYAML: "type: array\nitems:\n  type: array\n  items:\n    type: object\n    properties:\n      id:\n        format: int64\n        type: integer\n      name:\n        type: string\n    required:\n    - id\n    - name\n",
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

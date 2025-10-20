package openapi

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func RenderTypeSpecToYAML(spec *TypeSpec) (string, error) {
	yamlData, err := yaml.Marshal(&spec)
	if err != nil {
		return "", err
	}
	// Trim trailing newlines to avoid introducing blank lines between fields
	// when embedding rendered fragments into larger templates.
	return strings.TrimRight(string(yamlData), "\n"), nil
}

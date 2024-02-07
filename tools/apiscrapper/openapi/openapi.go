package openapi

import "gopkg.in/yaml.v3"

func RenderTypeSpecToYAML(spec *TypeSpec) (string, error) {
	yamlData, err := yaml.Marshal(&spec)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

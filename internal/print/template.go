package print

import (
	"bytes"
	"text/template"

	"gopkg.in/yaml.v3"
)

var funcMap = template.FuncMap{
	"toYaml": toYAML,
}

func toYAML(v interface{}) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

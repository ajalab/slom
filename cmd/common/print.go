package common

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

func PrintJSON(v any, stdout io.Writer) error {
	e := json.NewEncoder(stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "    ")

	return e.Encode(v)
}

func PrintYAML(v any, stdout io.Writer) error {
	e := yaml.NewEncoder(stdout)
	defer e.Close()

	return e.Encode(v)
}

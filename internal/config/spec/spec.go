package spec

import (
	"io"

	"gopkg.in/yaml.v3"
)

func ParseSpecConfig(r io.Reader) (*SpecConfig, error) {
	var config SpecConfig

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

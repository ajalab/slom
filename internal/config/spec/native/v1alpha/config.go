package v1alpha

import (
	core "github.com/ajalab/slom/internal/config/spec/core/v1alpha"
)

// SpecConfig is a configuration for an SLO specification.
// An SLO specification typically corresponds to a specific service or user journey and may include multiple SLO entries.
type SpecConfig struct {
	// Name is the name of the SLO specification.
	Name string `yaml:"name"`
	// Labels are the labels of the SLO specification.
	Labels map[string]string `yaml:"labels"`
	// Annotations are the annotations of the SLO specification.
	Annotations map[string]string `yaml:"annotations"`
	// SLOs are SLO configurations.
	SLOs []SLOConfig `yaml:"slos,omitempty"`
}

// SLOConfig is a configuration for SLO.
type SLOConfig struct {
	// Name is the name of the SLO.
	Name string `yaml:"name"`
	// Labels are the labels of the SLO.
	Labels map[string]string `yaml:"labels"`
	// Annotations are the annotations of the SLO.
	Annotations map[string]string `yaml:"annotations"`

	// Objective is the target of the SLO.
	Objective core.ObjectiveConfig `yaml:"objective"`
	// Indicator is the SLI for the SLO.
	Indicator core.IndicatorConfig `yaml:"indicator"`
	// Alerts are alert configurations for the SLO.
	Alerts []core.AlertConfig `yaml:"alerts,omitempty"`
	// Windows are windows used by the SLI and SLO.
	Windows []core.WindowConfig `yaml:"windows,omitempty"`
}

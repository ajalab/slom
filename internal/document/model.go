package document

// Document is a document of an SLO specification.
type Document struct {
	// Name is the name of the SLO specification.
	Name string `yaml:"name" json:"name"`
	// Labels are the labels of the SLO specification.
	Labels map[string]string `yaml:"labels" json:"labels"`
	// Annotations are the annotations of the SLO specification.
	Annotations map[string]string `yaml:"annotations" json:"annotations"`
	// SLOs are SLO configurations.
	SLOs []SLO `yaml:"slos,omitempty" json:"slos,omitempty"`
}

// SLO is a document for SLO.
type SLO struct {
	// Name is the name of the SLO.
	Name string `yaml:"name" json:"name"`
	// Labels are the labels of the SLO.
	Labels map[string]string `yaml:"labels" json:"labels"`
	// Annotations are the annotations of the SLO.
	Annotations map[string]string `yaml:"annotations" json:"annotations"`
	// Objective is the target of the SLO.
	Objective Objective `yaml:"objective" json:"objective"`
	// Indicator is the SLI for the SLO.
	Indicator Indicator `yaml:"indicator" json:"indicator"`
}

// Objective is a document for an SLO target.
type Objective struct {
	// Ratio is the target ratio of the SLO.
	Ratio float64 `yaml:"ratio" json:"ratio"`
	// WindowRef is the window name that refers to a window defined in SLOConfig.Windows.
	Window Window `yaml:"window" json:"window"`
}

// Indicator is a document for a service level indicator (SLI).
type Indicator struct {
	// Source is the metric source of the indicator (e.g., prometheus, ...)
	Source string `yaml:"source" json:"source"`
	// Query is the query of the indicator.
	Query string `yaml:"query" json:"query"`
}

// Window is a document for a window used by SLIs and SLOs.
// Either the Rolling or Calendar field must be specified.
type Window struct {
	// Name is the name of the window.
	Name string `yaml:"name" json:"name"`
	// Type is the type of the window: either rolling or calendar.
	Type string `yaml:"type" json:"type"`
	// Duration is the duration of the window.
	Duration string `yaml:"duration" json:"duration"`
}

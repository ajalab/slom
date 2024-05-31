package spec

// SpecConfig is a top-level configuration for Slogen.
type SpecConfig struct {
	Name string      `yaml:"name"`
	SLOs []SLOConfig `yaml:"slos,omitempty"`
}

// SLOConfig is a configuration for SLO.
type SLOConfig struct {
	Name      string          `yaml:"name"`
	Objective ObjectiveConfig `yaml:"objective"`
	Indicator IndicatorConfig `yaml:"indicator"`
	Alerts    []AlertConfig   `yaml:"alerts,omitempty"`
	Windows   []WindowConfig  `yaml:"windows,omitempty"`
}

// ObjectiveConfig is a configuration for SLO objective.
type ObjectiveConfig struct {
	Ratio     float64 `yaml:"ratio"`
	WindowRef string  `yaml:"windowRef"`
}

// IndicatorConfig is a configuration for a service level indicator (SLI).
type IndicatorConfig struct {
	Prometheus *PrometheusIndicatorConfig `yaml:"prometheus,omitempty"`
}

// PrometheusIndicatorConfig is a configuration for an SLI implemented with Prometheus.
type PrometheusIndicatorConfig struct {
	ErrorRatio string   `yaml:"errorRatio"`
	Level      []string `yaml:"level,omitempty"`
}

// WindowConfig is a configuration for a SLO window.
type WindowConfig struct {
	Name     string                `yaml:"name"`
	Rolling  *RollingWindowConfig  `yaml:"rolling,omitempty"`
	Calendar *CalendarWindowConfig `yaml:"calendar,omitempty"`
}

// RollingWindowConfig is a configuration for a SLO rolling window.
type RollingWindowConfig struct {
	Duration string `yaml:"duration"`
}

// CalendarWindowConfig is a configuration for a SLO rolling window.
type CalendarWindowConfig struct {
	Duration string `yaml:"duration"`
	Start    string `yaml:"start"`
}

// AlertConfig is a configuration for SLO alerts.
type AlertConfig struct {
	Name     string               `yaml:"name"`
	BurnRate *BurnRateAlertConfig `yaml:"burnRate"`
	Breach   *BreachAlertConfig   `yaml:"breach"`
	Alerter  AlerterConfig
}

// BurnRateAlertConfig is a configuration for SLO burn-rate alert.
type BurnRateAlertConfig struct {
	ConsumedBudgetRatio float64                          `yaml:"consumedBudgetRatio"`
	MultiWindows        *MultiWindowsBurnRateAlertConfig `yaml:"multiWindows"`
}

// MultiWindowsBurnRateAlertConfig is a configuration for SLO burn-rate alert with multi (long and short) windows.
type MultiWindowsBurnRateAlertConfig struct {
	ShortWindowRef string `yaml:"shortWindowRef"`
	LongWindowRef  string `yaml:"longWindowRef"`
}

// BreachAlertConfig is a configuration for SLO breach alert.
type BreachAlertConfig struct {
	WindowRef string `yaml:"windowRef"`
}

// AlerterConfig is a configuration for alerter, which fires alerts.
type AlerterConfig struct {
	Prometheus *PrometheusAlerterConfig `yaml:"prometheus,omitempty"`
}

type PrometheusAlerterConfig struct {
	Name        string            `yaml:"name,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

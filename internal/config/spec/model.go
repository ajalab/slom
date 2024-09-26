package spec

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
	Objective ObjectiveConfig `yaml:"objective"`
	// Indicator is the SLI for the SLO.
	Indicator IndicatorConfig `yaml:"indicator"`
	// Alerts are alert configurations for the SLO.
	Alerts []AlertConfig `yaml:"alerts,omitempty"`
	// Windows are windows used by the SLI and SLO.
	Windows []WindowConfig `yaml:"windows,omitempty"`
}

// ObjectiveConfig is a configuration for an SLO target.
type ObjectiveConfig struct {
	// Ratio is the target ratio of the SLO.
	Ratio float64 `yaml:"ratio"`
	// WindowRef is the window name that refers to a window defined in SLOConfig.Windows.
	WindowRef string `yaml:"windowRef"`
}

// IndicatorConfig is a configuration for a service level indicator (SLI).
type IndicatorConfig struct {
	// Prometheus is an SLI implemented with Prometheus.
	Prometheus *PrometheusIndicatorConfig `yaml:"prometheus,omitempty"`
}

// PrometheusIndicatorConfig is a configuration for an SLI implemented with Prometheus.
type PrometheusIndicatorConfig struct {
	// ErrorRatio is a PromQL query that computes the error ratio (0 - 1) of a service.
	ErrorRatio string `yaml:"errorRatio"`
	// Level is the list of Prometheus labels that represent the recording [aggregation level] of the query and appear in the query results.
	//
	// [aggregation level]: https://prometheus.io/docs/practices/rules/#naming
	Level []string `yaml:"level,omitempty"`
}

// WindowConfig is a configuration for a window used by SLIs and SLOs.
// Either the Rolling or Calendar field must be specified.
type WindowConfig struct {
	// Name is the name of the window.
	Name string `yaml:"name"`
	// Rolling specifies the window as a rolling window.
	Rolling *RollingWindowConfig `yaml:"rolling,omitempty"`
	// Calendar specifies the window as a calendar window.
	Calendar *CalendarWindowConfig `yaml:"calendar,omitempty"`
	// Prometheus specifies the evaluation configuration of the window.
	Prometheus *PrometheusWindowConfig `yaml:"prometheus,omitempty"`
}

// RollingWindowConfig is a configuration for an SLO rolling window.
type RollingWindowConfig struct {
	// Duration is the size of the window in [time.Duration] format.
	Duration string `yaml:"duration"`
}

// CalendarWindowConfig is a configuration for an SLO calendar window.
type CalendarWindowConfig struct {
	// Duration is the size of the window in [time.Duration] format.
	Duration string `yaml:"duration"`
	// Start is the starting point of the calendar windows.
	Start string `yaml:"start"`
}

type PrometheusWindowConfig struct {
	// EvaluationInterval represents how often rules associated with this window are evaluated.
	EvaluationInterval string `yaml:"evaluation_interval"`
}

// AlertConfig is a configuration for SLO alerts.
// Either the BurnRate or Breach field must be specified.
type AlertConfig struct {
	// Name is the name of the alert.
	Name string `yaml:"name"`
	// BurnRate specifies the alert as error budget burn rate alert.
	BurnRate *BurnRateAlertConfig `yaml:"burnRate"`
	// Breach specifies the alert as SLO breach alert.
	Breach *BreachAlertConfig `yaml:"breach"`
	// Alerter specifies how alerting is implemented for this alert.
	Alerter AlerterConfig
}

// BurnRateAlertConfig is a configuration for an SLO burn rate alert.
type BurnRateAlertConfig struct {
	// ConsumedBudgetRatio is the alerting threshold based on the ratio of the consumed error budget (0 - 1).
	ConsumedBudgetRatio float64 `yaml:"consumedBudgetRatio"`

	// MultiWindows specifies that the SLO burn rate alert is implemented as a [multiwindow] alert.
	//
	// [multiwindow]: https://sre.google/workbook/alerting-on-slos/#6-multiwindow-multi-burn-rate-alerts
	MultiWindows *MultiWindowsBurnRateAlertConfig `yaml:"multiWindows"`
}

// MultiWindowsBurnRateAlertConfig is a configuration for an SLO burn rate alert implemented as a [multiwindow] alert.
//
// [multiwindow]: https://sre.google/workbook/alerting-on-slos/#6-multiwindow-multi-burn-rate-alerts
type MultiWindowsBurnRateAlertConfig struct {
	// ShortWindowRef is the window name that refers to a window defined in SpecConfig.Windows.
	// The short window is a secondary window
	ShortWindowRef string `yaml:"shortWindowRef"`
	LongWindowRef  string `yaml:"longWindowRef"`
}

// BreachAlertConfig is a configuration for SLO breach alert.
type BreachAlertConfig struct {
	// WindowRef is the window name that referes to a window defined in SpecConfig.Windows.
	WindowRef string `yaml:"windowRef"`
}

// AlerterConfig is a configuration for alert implementation.
type AlerterConfig struct {
	// Prometheus specifies that the alert is implemented with Prometheus.
	Prometheus *PrometheusAlerterConfig `yaml:"prometheus,omitempty"`
}

type PrometheusAlerterConfig struct {
	// Name is the name of the Prometheus alert.
	Name string `yaml:"name,omitempty"`
	// Labels are the labels attached to Prometheus alerts.
	Labels map[string]string `yaml:"labels,omitempty"`
	// Annotations are the annotations attached to Prometheus alerts.
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

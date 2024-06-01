package series

import "time"

type SeriesSetConfig struct {
	Start    time.Time      `yaml:"start"`
	End      time.Time      `yaml:"end"`
	Interval time.Duration  `yaml:"interval"`
	Series   []SeriesConfig `yaml:"series"`
}

type SeriesConfig struct {
	SuccessFailure *SuccessFailureSeriesConfig `yaml:"successFailure"`
}

type SuccessFailureSeriesConfig struct {
	// Constant is the configuration for ConstantSuccessFailureSeriesGenerator.
	Constant *ConstantSuccessFailureSeriesGeneratorConfig `yaml:"constant"`
	// Binomial is the configuration for BinomialSuccessFailureSeriesGenerator.
	Binomial *BinomialSuccessFailureSeriesGeneratorConfig `yaml:"binomial"`
	// MetricFamilyName is the name of the metric family.
	MetricFamilyName string `yaml:"metricFamilyName"`
	// MetricFamilyHelp is the help (description) of the metric family.
	MetricFamilyHelp string `yaml:"metricFamilyHelp"`
	// LabelNameStatus is the name of the label describing success or failure status.
	LabelNameStatus string `yaml:"labelNameStatus"`
	// LabelValueSuccess is the value of the label represending success.
	LabelValueSuccess string `yaml:"labelValueSuccess"`
	// LabelValueFailure is the value of the label represending failure.
	LabelValueFailure string `yaml:"labelValueFailure"`
	// Labels is a set of labels in addition to the above labels.
	Labels map[string]string `yaml:"labels"`
}

type ConstantSuccessFailureSeriesGeneratorConfig struct {
	// Throughput is the throughput of the success counter.
	ThroughputSuccess int `yaml:"throughputSuccess"`
	// Throughput is the throughput of the failure counter.
	ThroughputFailure int `yaml:"throughputFailure"`
	// Overrides are overriding configurations.
	Overrides []ConstantSuccessFailureSeriesOverridesConfig `yaml:"overrides"`
}

type ConstantSuccessFailureSeriesOverridesConfig struct {
	Start             time.Time `yaml:"start"`
	End               time.Time `yaml:"end"`
	ThroughputSuccess int       `yaml:"throughputSuccess"`
	ThroughputFailure int       `yaml:"throughputFailure"`
}

type BinomialSuccessFailureSeriesGeneratorConfig struct {
	// Throughput is the throughput of the counter.
	Throughput int `yaml:"throughput"`
	// BaseErrorRate is the error rate in the period.
	BaseErrorRate float64 `yaml:"baseErrorRate"`
}

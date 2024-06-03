package series

import "time"

type SeriesSetConfig struct {
	Start          time.Time            `yaml:"start"`
	End            time.Time            `yaml:"end"`
	Interval       time.Duration        `yaml:"interval"`
	MetricFamilies []MetricFamilyConfig `yaml:"metricFamilies"`
}

type MetricFamilyConfig struct {
	// Name is the name of the metric family.
	Name string `yaml:"name"`
	// Help is the help (description) of the metric family.
	Help string `yaml:"help"`
	// Series is a list of series configurations.
	Series []SeriesConfig `yaml:"series"`
}

type SeriesConfig struct {
	// SuccessFailure is the configuration for SuccessFailure series.
	SuccessFailure *SuccessFailureSeriesConfig `yaml:"successFailure"`
	// Labels is a set of labels in addition to the above labels.
	Labels map[string]string `yaml:"labels"`
}

type SuccessFailureSeriesConfig struct {
	// Constant is the configuration for ConstantSuccessFailureSeries.
	Constant *ConstantSuccessFailureSeriesConfig `yaml:"constant"`
	// Binomial is the configuration for BinomialSuccessFailureSeries.
	Binomial *BinomialSuccessFailureSeriesConfig `yaml:"binomial"`
	// LabelNameStatus is the name of the label describing success or failure status.
	LabelNameStatus string `yaml:"labelNameStatus"`
	// LabelValueSuccess is the value of the label represending success.
	LabelValueSuccess string `yaml:"labelValueSuccess"`
	// LabelValueFailure is the value of the label represending failure.
	LabelValueFailure string `yaml:"labelValueFailure"`
}

type ConstantSuccessFailureSeriesConfig struct {
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

type BinomialSuccessFailureSeriesConfig struct {
	// Throughput is the throughput of the counter.
	Throughput int `yaml:"throughput"`
	// BaseErrorRate is the error rate in the period.
	BaseErrorRate float64 `yaml:"baseErrorRate"`
}

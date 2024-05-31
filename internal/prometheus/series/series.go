package series

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"time"

	configseries "github.com/ajalab/slogen/internal/config/series"
	"github.com/ajalab/slogen/internal/prometheus"
	"github.com/prometheus/common/model"
	"golang.org/x/exp/rand"
	"gopkg.in/yaml.v3"

	"gonum.org/v1/gonum/stat/distuv"
)

type successFailureSeriesGenerator interface {
	generate(
		successWriter func(v int, t time.Time) error,
		failureWriter func(v int, t time.Time) error,
	) (time.Time, time.Time, error)
}

type constantSuccessFailureSeriesGenerator struct {
	start             time.Time
	end               time.Time
	interval          time.Duration
	throughputSuccess int
	throughputFailure int
	overrides         []constantSuccessFailureSeriesOverride
}

type constantSuccessFailureSeriesOverride struct {
	start             time.Time
	end               time.Time
	throughputSuccess int
	throughputFailure int
}

func (g *constantSuccessFailureSeriesGenerator) generate(
	successWriter func(v int, t time.Time) error,
	failureWriter func(v int, t time.Time) error,
) (time.Time, time.Time, error) {
	t := g.start
	failureTotal := 0
	successTotal := 0
	for t.Before(g.end) {
		failureTotalDelta := g.throughputFailure
		successTotalDelta := g.throughputSuccess
		for i := len(g.overrides) - 1; i >= 0; i-- {
			o := g.overrides[i]
			if (t.Equal(o.start) || t.After(o.start)) && t.Before(o.end) {
				failureTotalDelta = o.throughputFailure
				successTotalDelta = o.throughputSuccess
				break
			}
		}
		failureTotal += failureTotalDelta
		successTotal += successTotalDelta
		if err := successWriter(successTotal, t); err != nil {
			return time.Time{}, time.Time{}, err
		}
		if err := failureWriter(failureTotal, t); err != nil {
			return time.Time{}, time.Time{}, err
		}

		t = t.Add(g.interval)
	}
	return g.start, g.end, nil
}

type binomialSuccessFailureSeriesGenerator struct {
	start         time.Time
	end           time.Time
	interval      time.Duration
	throughput    int
	baseErrorRate float64
	source        rand.Source
}

func (g *binomialSuccessFailureSeriesGenerator) generate(
	successWriter func(v int, t time.Time) error,
	failureWriter func(v int, t time.Time) error,
) (time.Time, time.Time, error) {
	t := g.start
	failureTotal := 0
	successTotal := 0
	baseDist := distuv.Binomial{
		N:   float64(g.throughput),
		P:   g.baseErrorRate,
		Src: rand.New(g.source),
	}
	for t.Before(g.end) {
		failure := int(baseDist.Rand())
		success := g.throughput - failure
		failureTotal += failure
		successTotal += success
		if err := successWriter(successTotal, t); err != nil {
			return time.Time{}, time.Time{}, err
		}
		if err := failureWriter(failureTotal, t); err != nil {
			return time.Time{}, time.Time{}, err
		}

		t = t.Add(g.interval)
	}
	return g.start, g.end, nil
}

type SeriesTestDataGenerator interface {
	GenerateOpenMetrics(
		w io.Writer,
	) (time.Time, time.Time, error)

	GenerateUnitTest(
		ruleFiles []string,
		w io.Writer,
	) (time.Time, time.Time, error)
}

func NewSeriesTestDataGenerator(
	config *configseries.SeriesConfig,
) (SeriesTestDataGenerator, error) {
	start := config.Start
	end := config.End
	interval := config.Interval
	if config := config.SuccessFailure; config != nil {
		var sg successFailureSeriesGenerator
		switch {
		case config.Constant != nil && config.Binomial == nil:
			var overrides []constantSuccessFailureSeriesOverride
			for _, o := range config.Constant.Overrides {
				overrides = append(overrides, constantSuccessFailureSeriesOverride{
					start:             o.Start,
					end:               o.End,
					throughputSuccess: o.ThroughputSuccess,
					throughputFailure: o.ThroughputFailure,
				})
			}
			sg = &constantSuccessFailureSeriesGenerator{
				start:             start,
				end:               end,
				interval:          interval,
				throughputSuccess: config.Constant.ThroughputSuccess,
				throughputFailure: config.Constant.ThroughputFailure,
				overrides:         overrides,
			}
		case config.Binomial != nil && config.Constant == nil:
			sg = &binomialSuccessFailureSeriesGenerator{
				start:         start,
				end:           end,
				interval:      interval,
				throughput:    config.Binomial.Throughput,
				baseErrorRate: config.Binomial.BaseErrorRate,
			}
		default:
			return nil, fmt.Errorf("either constant or binomial generator must be specified")
		}
		return &SuccessFailureSeriesTestDataGenerator{
			name:              config.MetricFamilyName,
			description:       config.MetricFamilyHelp,
			labelNameStatus:   config.LabelNameStatus,
			labelValueSuccess: config.LabelValueSuccess,
			labelValueFailure: config.LabelValueFailure,
			labels:            config.Labels,
			seriesGenerator:   sg,
			interval:          interval,
		}, nil
	}

	return nil, fmt.Errorf("either successFailure or ... must be specified")
}

type SuccessFailureSeriesTestDataGenerator struct {
	name              string
	description       string
	labelNameStatus   string
	labelValueSuccess string
	labelValueFailure string
	labels            map[string]string
	seriesGenerator   successFailureSeriesGenerator
	interval          time.Duration
}

func (g *SuccessFailureSeriesTestDataGenerator) GenerateOpenMetrics(
	w io.Writer,
) (time.Time, time.Time, error) {
	if _, err := fmt.Fprintf(w, "# HELP %[1]s %[2]s\n# TYPE %[1]s counter\n", g.name, g.description); err != nil {
		return time.Time{}, time.Time{}, err
	}

	labelsSuccess := maps.Clone(g.labels)
	labelsSuccess[g.labelNameStatus] = g.labelValueSuccess
	labelsSuccessStr := prometheus.GenerateLabels(labelsSuccess, false)
	labelsFailure := maps.Clone(g.labels)
	labelsFailure[g.labelNameStatus] = g.labelValueFailure
	labelsFailureStr := prometheus.GenerateLabels(labelsFailure, false)

	s, e, err := g.seriesGenerator.generate(
		func(v int, t time.Time) error {
			_, err := fmt.Fprintf(w, "%s%s %d %d\n", g.name, labelsSuccessStr, v, t.Unix())
			return err
		},
		func(v int, t time.Time) error {
			_, err := fmt.Fprintf(w, "%s%s %d %d\n", g.name, labelsFailureStr, v, t.Unix())
			return err
		},
	)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if _, err = fmt.Fprintln(w, "# EOF"); err != nil {
		return time.Time{}, time.Time{}, err
	}
	return s, e, nil
}

func (g *SuccessFailureSeriesTestDataGenerator) GenerateUnitTest(
	ruleFiles []string,
	w io.Writer,
) (time.Time, time.Time, error) {
	var valuesBufSuccess bytes.Buffer
	var valuesBufFailure bytes.Buffer

	cswSuccess := newCompressingSeriesWriter(&valuesBufSuccess)
	cswFailure := newCompressingSeriesWriter(&valuesBufFailure)

	s, e, err := g.seriesGenerator.generate(
		cswSuccess.writerFunc(),
		cswFailure.writerFunc(),
	)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	cswSuccess.Close()
	cswFailure.Close()

	labelsSuccess := maps.Clone(g.labels)
	labelsSuccess[g.labelNameStatus] = g.labelValueSuccess
	labelsSuccessStr := prometheus.GenerateLabels(labelsSuccess, true)
	labelsFailure := maps.Clone(g.labels)
	labelsFailure[g.labelNameStatus] = g.labelValueFailure
	labelsFailureStr := prometheus.GenerateLabels(labelsFailure, true)

	test := testGroup{
		Interval: model.Duration(g.interval),
		InputSeries: []series{
			{
				Series: g.name + labelsSuccessStr,
				Values: valuesBufSuccess.String(),
			},
			{
				Series: g.name + labelsFailureStr,
				Values: valuesBufFailure.String(),
			},
		},
		AlertRuleTests:  []alertTestCase{},
		PromqlExprTests: []promqlTestCase{},
	}

	unitTestFile := unitTestFile{
		RuleFiles:          ruleFiles,
		EvaluationInterval: model.Duration(g.interval),
		Tests:              []testGroup{test},
	}

	encoder := yaml.NewEncoder(w)
	defer encoder.Close()

	if err := encoder.Encode(&unitTestFile); err != nil {
		return time.Time{}, time.Time{}, err
	}

	return s, e, nil
}

// Below part is brought from https://github.com/prometheus/prometheus/blob/main/cmd/promtool/unittest.go
// to produce rule unit test files.
// Start

type unitTestFile struct {
	RuleFiles          []string       `yaml:"rule_files"`
	EvaluationInterval model.Duration `yaml:"evaluation_interval,omitempty"`
	GroupEvalOrder     []string       `yaml:"group_eval_order"`
	Tests              []testGroup    `yaml:"tests"`
}

type testGroup struct {
	Interval        model.Duration   `yaml:"interval"`
	InputSeries     []series         `yaml:"input_series"`
	AlertRuleTests  []alertTestCase  `yaml:"alert_rule_test"`
	PromqlExprTests []promqlTestCase `yaml:"promql_expr_test"`
	// ExternalLabels  labels.Labels    `yaml:"external_labels,omitempty"` Modification: we don't use external labels
	ExternalURL   string `yaml:"external_url,omitempty"`
	TestGroupName string `yaml:"name,omitempty"`
}

type series struct {
	Series string `yaml:"series"`
	Values string `yaml:"values"`
}

type alertTestCase struct {
	EvalTime  model.Duration `yaml:"eval_time"`
	Alertname string         `yaml:"alertname"`
	ExpAlerts []alert        `yaml:"exp_alerts"`
}

type alert struct {
	ExpLabels      map[string]string `yaml:"exp_labels"`
	ExpAnnotations map[string]string `yaml:"exp_annotations"`
}

type promqlTestCase struct {
	Expr       string         `yaml:"expr"`
	EvalTime   model.Duration `yaml:"eval_time"`
	ExpSamples []sample       `yaml:"exp_samples"`
}

type sample struct {
	Labels    string  `yaml:"labels"`
	Value     float64 `yaml:"value"`
	Histogram string  `yaml:"histogram"` // A non-empty string means Value is ignored.
}

// End

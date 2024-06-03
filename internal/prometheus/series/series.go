package series

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	configseries "github.com/ajalab/slogen/internal/config/series"
	"github.com/prometheus/common/model"
	"golang.org/x/exp/rand"
	"gopkg.in/yaml.v3"

	"gonum.org/v1/gonum/stat/distuv"
)

type successFailureMetricPointsGenerator interface {
	generate(
		start time.Time,
		end time.Time,
		interval time.Duration,
		writerSuccess func(v int, t int64) error,
		writerFailure func(v int, t int64) error,
	) error
}

type constantSuccessFailureMetricPointsGenerator struct {
	throughputSuccess int
	throughputFailure int
	overrides         []constantSuccessFailureMetricPointsOverride
}

type constantSuccessFailureMetricPointsOverride struct {
	start             time.Time
	end               time.Time
	throughputSuccess int
	throughputFailure int
}

var _ successFailureMetricPointsGenerator = &constantSuccessFailureMetricPointsGenerator{}

func (g *constantSuccessFailureMetricPointsGenerator) generate(
	start time.Time,
	end time.Time,
	interval time.Duration,
	writerSuccess func(v int, t int64) error,
	writerFailure func(v int, t int64) error,
) error {
	var (
		totalSuccess int
		totalFailure int
	)
	t := start

	for t.Before(end) {
		success := g.throughputSuccess
		failure := g.throughputFailure
		for i := len(g.overrides) - 1; i >= 0; i-- {
			o := g.overrides[i]
			if (t.Equal(o.start) || t.After(o.start)) && t.Before(o.end) {
				success = o.throughputSuccess
				failure = o.throughputFailure
				break
			}
		}
		totalSuccess += success
		totalFailure += failure

		tUnix := t.Unix()
		if err := writerSuccess(totalSuccess, tUnix); err != nil {
			return err
		}
		if err := writerFailure(totalFailure, tUnix); err != nil {
			return err
		}

		t = t.Add(interval)
	}

	return nil
}

type binomialSuccessFailureMetricPointsGenerator struct {
	throughput    int
	baseErrorRate float64
	baseDist      distuv.Binomial
}

func newBinomialSuccessFailureMetricPointsGenerator(
	throughput int,
	baseErrorRate float64,
	source rand.Source,
) *binomialSuccessFailureMetricPointsGenerator {
	baseDist := distuv.Binomial{
		N:   float64(throughput),
		P:   baseErrorRate,
		Src: rand.New(source),
	}
	return &binomialSuccessFailureMetricPointsGenerator{
		throughput:    throughput,
		baseErrorRate: baseErrorRate,
		baseDist:      baseDist,
	}
}

var _ successFailureMetricPointsGenerator = &binomialSuccessFailureMetricPointsGenerator{}

func (g *binomialSuccessFailureMetricPointsGenerator) generate(
	start time.Time,
	end time.Time,
	interval time.Duration,
	writerSuccess func(v int, t int64) error,
	writerFailure func(v int, t int64) error,
) error {
	var (
		totalSuccess int
		totalFailure int
	)
	t := start

	for t.Before(end) {
		failure := int(g.baseDist.Rand())
		success := g.throughput - failure
		totalSuccess += success
		totalFailure += failure

		tUnix := t.Unix()
		if err := writerSuccess(totalSuccess, tUnix); err != nil {
			return err
		}
		if err := writerFailure(totalFailure, tUnix); err != nil {
			return err
		}

		t = t.Add(interval)
	}

	return nil
}

type seriesGenerator interface {
	generateOpenMetricsSeries(
		start time.Time,
		end time.Time,
		interval time.Duration,
		w io.Writer,
	) error

	generateUnitTestSeries(
		start time.Time,
		end time.Time,
		interval time.Duration,
	) []series
}

type successFailureSeriesGenerator struct {
	metricPointsGenerator successFailureMetricPointsGenerator
	name                  string
	labelNameStatus       string
	labelValueSuccess     string
	labelValueFailure     string
	labels                map[string]string
}

func (sg *successFailureSeriesGenerator) generateLabels(commaSpace bool) (string, string) {
	var ls []string
	for name, value := range sg.labels {
		ls = append(ls, fmt.Sprintf("%s=\"%s\"", name, value))
	}

	lsSuccess := append(slices.Clone(ls), fmt.Sprintf("%s=\"%s\"", sg.labelNameStatus, sg.labelValueSuccess))
	lsFailure := append(ls, fmt.Sprintf("%s=\"%s\"", sg.labelNameStatus, sg.labelValueFailure))

	var sep = ","
	if commaSpace {
		sep = ", "
	}

	return "{" + strings.Join(lsSuccess, sep) + "}", "{" + strings.Join(lsFailure, sep) + "}"
}

var _ seriesGenerator = &successFailureSeriesGenerator{}

func (g *successFailureSeriesGenerator) generateOpenMetricsSeries(
	start time.Time,
	end time.Time,
	interval time.Duration,
	w io.Writer,
) error {
	labelsSuccess, labelsFailure := g.generateLabels(false)

	var buf bytes.Buffer
	g.metricPointsGenerator.generate(
		start,
		end,
		interval,
		func(v int, t int64) error {
			_, err := fmt.Fprintf(w, "%s%s %d %d\n", g.name, labelsSuccess, v, t)
			return err
		},
		func(v int, t int64) error {
			_, err := fmt.Fprintf(&buf, "%s%s %d %d\n", g.name, labelsFailure, v, t)
			return err
		},
	)
	if _, err := io.Copy(w, &buf); err != nil {
		return err
	}
	return nil
}

func (g *successFailureSeriesGenerator) generateUnitTestSeries(
	start time.Time,
	end time.Time,
	interval time.Duration,
) []series {
	var valuesBufSuccess bytes.Buffer
	var valuesBufFailure bytes.Buffer

	cswSuccess := newCompressingSeriesWriter(&valuesBufSuccess)
	cswFailure := newCompressingSeriesWriter(&valuesBufFailure)

	g.metricPointsGenerator.generate(start, end, interval, cswSuccess.writerFunc(), cswFailure.writerFunc())
	cswSuccess.Close()
	cswFailure.Close()

	labelsSuccess, labelsFailure := g.generateLabels(true)

	seriesSuccess := series{
		Series: g.name + labelsSuccess,
		Values: valuesBufSuccess.String(),
	}
	seriesFailure := series{
		Series: g.name + labelsFailure,
		Values: valuesBufFailure.String(),
	}
	return []series{seriesSuccess, seriesFailure}
}

type metricFamilyGenerator struct {
	name             string
	help             string
	seriesGenerators []seriesGenerator
}

func (g *metricFamilyGenerator) generateOpenMetricsMetricFamily(
	start time.Time,
	end time.Time,
	interval time.Duration,
	w io.Writer,
) error {
	if _, err := fmt.Fprintf(w, "# HELP %[1]s %[2]s\n# TYPE %[1]s counter\n", g.name, g.help); err != nil {
		return err
	}

	for _, sg := range g.seriesGenerators {
		if err := sg.generateOpenMetricsSeries(start, end, interval, w); err != nil {
			return err
		}
	}

	return nil
}

func (g *metricFamilyGenerator) generateUnitTestSeries(
	start time.Time,
	end time.Time,
	interval time.Duration,
) []series {
	var series []series
	for _, sg := range g.seriesGenerators {
		series = append(series, sg.generateUnitTestSeries(start, end, interval)...)
	}
	return series
}

type SeriesSetGenerator struct {
	start                  time.Time
	end                    time.Time
	interval               time.Duration
	metricFamilyGenerators []*metricFamilyGenerator
}

func NewSeriesSetGenerator(config *configseries.SeriesSetConfig) (*SeriesSetGenerator, error) {
	var metricFamilyGenerators []*metricFamilyGenerator
	for _, metricFamilyConfig := range config.MetricFamilies {
		var seriesGenerators []seriesGenerator
		for _, seriesConfig := range metricFamilyConfig.Series {
			var sg seriesGenerator
			switch {
			case seriesConfig.SuccessFailure != nil:
				c := seriesConfig.SuccessFailure

				var mpg successFailureMetricPointsGenerator
				switch {
				case c.Constant != nil && c.Binomial == nil:
					var overrides []constantSuccessFailureMetricPointsOverride
					for _, o := range c.Constant.Overrides {
						overrides = append(overrides, constantSuccessFailureMetricPointsOverride{
							start:             o.Start,
							end:               o.End,
							throughputSuccess: o.ThroughputSuccess,
							throughputFailure: o.ThroughputFailure,
						})
					}
					mpg = &constantSuccessFailureMetricPointsGenerator{
						throughputSuccess: c.Constant.ThroughputSuccess,
						throughputFailure: c.Constant.ThroughputFailure,
						overrides:         overrides,
					}
				case c.Binomial != nil && c.Constant == nil:
					mpg = newBinomialSuccessFailureMetricPointsGenerator(
						c.Binomial.Throughput,
						c.Binomial.BaseErrorRate,
						rand.NewSource(0),
					)
				default:
					return nil, fmt.Errorf("either constant or binomial generator must be specified")
				}

				sg = &successFailureSeriesGenerator{
					metricPointsGenerator: mpg,
					name:                  metricFamilyConfig.Name,
					labelNameStatus:       c.LabelNameStatus,
					labelValueSuccess:     c.LabelValueSuccess,
					labelValueFailure:     c.LabelValueFailure,
					labels:                seriesConfig.Labels,
				}
			default:
				return nil, fmt.Errorf("success failure generator configuration must be specified")
			}

			seriesGenerators = append(seriesGenerators, sg)
		}

		mfg := &metricFamilyGenerator{
			name:             metricFamilyConfig.Name,
			help:             metricFamilyConfig.Help,
			seriesGenerators: seriesGenerators,
		}
		metricFamilyGenerators = append(metricFamilyGenerators, mfg)
	}

	return &SeriesSetGenerator{
		start:                  config.Start,
		end:                    config.End,
		interval:               config.Interval,
		metricFamilyGenerators: metricFamilyGenerators,
	}, nil
}

func (g *SeriesSetGenerator) GenerateOpenMetrics(w io.Writer) error {
	for _, mfg := range g.metricFamilyGenerators {
		if err := mfg.generateOpenMetricsMetricFamily(g.start, g.end, g.interval, w); err != nil {
			return fmt.Errorf("failed to write OpenMetrics metric family")
		}
	}
	_, err := fmt.Fprintln(w, "# EOF")
	return err
}

func (g *SeriesSetGenerator) GenerateUnitTest(
	ruleFiles []string,
	w io.Writer,
) error {
	var inputSeries []series
	for _, mfg := range g.metricFamilyGenerators {
		series := mfg.generateUnitTestSeries(g.start, g.end, g.interval)
		inputSeries = append(inputSeries, series...)
	}

	test := testGroup{
		Interval:        model.Duration(g.interval),
		InputSeries:     inputSeries,
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

	return encoder.Encode(&unitTestFile)
}

func (g *SeriesSetGenerator) Start() time.Time {
	return g.start
}

func (g *SeriesSetGenerator) End() time.Time {
	return g.end
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

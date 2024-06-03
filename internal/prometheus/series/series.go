package series

import (
	"fmt"
	"io"
	"time"

	configseries "github.com/ajalab/slogen/internal/config/series"
	"github.com/prometheus/common/model"
	"golang.org/x/exp/rand"
	"gopkg.in/yaml.v3"
)

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

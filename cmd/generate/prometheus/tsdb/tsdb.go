package tsdb

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/ajalab/slogen/cmd/common"
	configseries "github.com/ajalab/slogen/internal/config/series"
	configspec "github.com/ajalab/slogen/internal/config/spec"
	"github.com/ajalab/slogen/internal/prometheus/rule"
	"github.com/ajalab/slogen/internal/prometheus/series"
	"github.com/ajalab/slogen/internal/prometheus/tsdb"
	"github.com/ajalab/slogen/internal/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func run(
	ctx context.Context,
	logger *slog.Logger,
	tsdbDirName string,
	seriesFileName string,
	specFileName string,
	start time.Time,
	end time.Time,
	interval time.Duration,
	backfiller string,
	backfillerPrometheus string,
	backfillerPromtool string,
	backfillerImage string,
) error {
	var b tsdb.TSDBBackfiller
	switch backfiller {
	case "default":
		b = tsdb.NewDefaultTSDBBackfiller(logger, backfillerPrometheus, backfillerPromtool)
	case "docker":
		dockerB, err := tsdb.NewDockerTSDBBackfiller(logger, backfillerImage)
		if err != nil {
			return fmt.Errorf("failed to create DockerTSDBBackfiller: %w", err)
		}
		if err := dockerB.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize the TSDB backfiller: %w", err)
		}
		b = dockerB
		defer dockerB.Close()
	default:
		return fmt.Errorf("unknown backfiller: %s", backfiller)
	}

	s := start
	e := end
	if seriesFileName != "" {
		openMetricsSeriesFile, err := os.CreateTemp("", "slogen-generate-prometheus-tsdb-openMetricsSeries-*")
		if err != nil {
			return fmt.Errorf("failed to create a temporary file for OpenMetrics series: %w", err)
		}
		defer os.Remove(openMetricsSeriesFile.Name())

		if s, e, err = writeOpenMetricsSeries(openMetricsSeriesFile, seriesFileName, start, end, interval); err != nil {
			return fmt.Errorf("failed to write OpenMetrics series file from %s: %w", seriesFileName, err)
		}
		logger.Debug("wrote OpenMetrics series file", "openMetricsSeriesFile", openMetricsSeriesFile.Name())

		if err := b.BackfillOpenMetricsSeries(ctx, openMetricsSeriesFile.Name(), tsdbDirName); err != nil {
			return fmt.Errorf("failed to backfill series to TSDB: %w", err)
		}
	}

	if specFileName != "" {
		prometheusRuleFile, err := os.CreateTemp("", "slogen-generate-prometheus-tsdb-prometheusRule-*")
		if err != nil {
			return fmt.Errorf("failed to create a temporary file for Prometheus rules: %w", err)
		}
		defer os.Remove(prometheusRuleFile.Name())

		if err := writePrometheusRule(prometheusRuleFile, specFileName); err != nil {
			return fmt.Errorf("failed to write Prometheus recording rule group file from %s: %w", specFileName, err)
		}
		logger.Debug("wrote rule file", "prometheusRuleFile", prometheusRuleFile.Name())

		if err := b.BackfillRule(ctx, prometheusRuleFile.Name(), s, e, tsdbDirName); err != nil {
			return fmt.Errorf("failed to backfill rule to TSDB: %w", err)
		}
	}
	return nil
}

func writeOpenMetricsSeries(
	w io.Writer,
	seriesFileName string,
	start time.Time,
	end time.Time,
	interval time.Duration,
) (time.Time, time.Time, error) {
	seriesFile, err := os.Open(seriesFileName)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to open %s: %w", seriesFileName, err)
	}
	defer seriesFile.Close()

	seriesConfigParser := configseries.SeriesConfigParser{
		Start:    start,
		End:      end,
		Interval: interval,
	}
	seriesConfig, err := seriesConfigParser.Parse(seriesFile)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse %s: %w", seriesFileName, err)
	}

	g, err := series.NewSeriesSetGenerator(seriesConfig)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to create a generator: %w", err)
	}

	return g.Start(), g.End(), g.GenerateOpenMetrics(w)
}

func writePrometheusRule(
	w io.Writer,
	specFileName string,
) error {
	specFile, err := os.Open(specFileName)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", specFileName, err)
	}
	defer specFile.Close()

	specConfig, err := configspec.ParseSpecConfig(specFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", specFileName, err)
	}
	spec, err := spec.ToSpec(specConfig)
	if err != nil {
		return fmt.Errorf("failed to convert a config %s into spec: %w", specFileName, err)
	}

	g := rule.NewRuleGenerator()
	err = g.GenerateRecordingRules(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}
	recordingRuleGroups := g.RuleGroups()
	prometheusRecordingRuleGroups := recordingRuleGroups.Prometheus()

	e := yaml.NewEncoder(w)
	defer e.Close()

	if err := e.Encode(&prometheusRecordingRuleGroups); err != nil {
		return fmt.Errorf("failed to encode recording rule groups: %w", err)
	}

	return nil
}

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	var seriesFileName string
	var specFileName string
	var start string
	var end string
	var interval string
	var backfiller string
	var backfillerPrometheus string
	var backfillerPromtool string
	var backfillerImage string

	command := &cobra.Command{
		Use:   "prometheus-tsdb [--series seriesFileName] [--spec specFileName] -s startTime -e endTime -i interval tsdbDirName",
		Short: "Generate TSDB data for prometheus",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			logger := common.NewLogger(flags.Debug, cmd.OutOrStdout())

			var err error
			var startTime time.Time
			if start != "" {
				if startTime, err = time.Parse(time.RFC3339, start); err != nil {
					return fmt.Errorf("failed to parse date time in start: %w", err)
				}
			}
			var endTime time.Time
			if end != "" {
				if endTime, err = time.Parse(time.RFC3339, end); err != nil {
					return fmt.Errorf("failed to parse date time in end: %w", err)
				}
			}
			var intervalDuration time.Duration
			if interval != "" {
				if intervalDuration, err = time.ParseDuration(interval); err != nil {
					return fmt.Errorf("failed to parse interval: %w", err)
				}
			}
			return run(
				ctx,
				logger,
				args[0],
				seriesFileName,
				specFileName,
				startTime,
				endTime,
				intervalDuration,
				backfiller,
				backfillerPrometheus,
				backfillerPromtool,
				backfillerImage,
			)
		},
	}
	command.DisableFlagsInUseLine = true
	command.Flags().SortFlags = false
	command.Flags().StringVar(&seriesFileName, "series", "", "path to the series file config")
	command.Flags().StringVar(&specFileName, "spec", "", "path to the spec file config")
	command.Flags().StringVarP(&start, "start", "s", "", "start time of the generated series in RFC3339")
	command.Flags().StringVarP(&end, "end", "e", "", "end time of the generated series in RFC3339")
	command.Flags().StringVarP(&interval, "interval", "i", "", "interval of the generated series")
	command.Flags().StringVar(&backfiller, "backfiller", "default", "tsdb backfiller implementation to use. Either default or docker")
	command.Flags().StringVar(&backfillerPrometheus, "backfiller-prometheus", "prometheus", "path to the prometheus executable to run backfiller (only used in default backfiller")
	command.Flags().StringVar(&backfillerPromtool, "backfiller-promtool", "promtool", "path to the promtool executable to run backfiller (only used in default backfiller")
	command.Flags().StringVar(&backfillerImage, "backfiller-image", "prom/prometheus", "image name of prometheus (only used in docker backfiller)")

	return command
}

package series

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ajalab/slom/cmd/common"
	configseries "github.com/ajalab/slom/internal/config/series"
	"github.com/ajalab/slom/internal/prometheus/series"
	"github.com/spf13/cobra"
)

func run(
	output string,
	ruleFiles []string,
	start time.Time,
	end time.Time,
	interval time.Duration,
	args []string,
	stdout io.Writer,
) error {
	fileName := args[0]
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", fileName, err)
	}
	defer file.Close()

	seriesConfigParser := configseries.SeriesConfigParser{
		Start:    start,
		End:      end,
		Interval: interval,
	}
	seriesConfig, err := seriesConfigParser.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", fileName, err)
	}

	g, err := series.NewSeriesSetGenerator(seriesConfig)
	if err != nil {
		return fmt.Errorf("failed to create a generator: %w", err)
	}

	switch output {
	case "openmetrics":
		return g.GenerateOpenMetrics(stdout)
	case "unittest":
		return g.GenerateUnitTest(ruleFiles, stdout)
	}

	return fmt.Errorf("unsupported output format: %s", output)
}

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	var output string
	var ruleFiles []string
	var start string
	var end string
	var interval string

	command := &cobra.Command{
		Use:   "prometheus-series -s startTime -e endTime -i interval seriesFile",
		Short: "Generate time series for Prometheus-compatible systems",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				output,
				ruleFiles,
				startTime,
				endTime,
				intervalDuration,
				args,
				cmd.OutOrStdout(),
			)
		},
	}
	command.Flags().SortFlags = false

	command.Flags().StringVarP(&output, "output", "o", "openmetrics", "format of the output data. Either \"openmetrics\" for OpenMetrics time series or \"unittest\" for Promtool rule unit tests")
	command.Flags().StringArrayVarP(&ruleFiles, "rule-files", "r", []string{}, "list of rule file names.")
	command.Flags().StringVarP(&start, "start", "s", "", "start time of the generated series in RFC3339")
	command.Flags().StringVarP(&end, "end", "e", "", "end time of the generated series in RFC3339")
	command.Flags().StringVarP(&interval, "interval", "i", "", "interval of the generated series")

	return command
}

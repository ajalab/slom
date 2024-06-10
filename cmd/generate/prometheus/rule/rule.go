package rule

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/ajalab/slogen/cmd/common"
	configspec "github.com/ajalab/slogen/internal/config/spec"
	"github.com/ajalab/slogen/internal/prometheus/rule"
	"github.com/ajalab/slogen/internal/spec"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/spf13/cobra"
)

func run(types []string, output string, args []string, stdout io.Writer) error {
	recordEnabled := slices.Contains(types, "record")
	alertEnabled := slices.Contains(types, "alert")
	if !recordEnabled && !alertEnabled {
		return fmt.Errorf("either \"record\" or \"alert\" must be specified as types")
	}

	fileName := args[0]
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", fileName, err)
	}
	defer file.Close()

	config, err := configspec.ParseSpecConfig(file)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", fileName, err)
	}

	spec, err := spec.ToSpec(config)
	if err != nil {
		return fmt.Errorf("failed to convert a config %s into spec: %w", fileName, err)
	}

	switch output {
	case "json":
		return runJSON(spec, recordEnabled, alertEnabled, stdout)
	case "prometheus":
		return runPrometheus(spec, recordEnabled, alertEnabled, stdout)
	}
	return fmt.Errorf("unsupported format: %s", output)
}

func runJSON(
	spec *spec.Spec,
	recordEnabled, alertEnabled bool,
	stdout io.Writer,
) error {
	if recordEnabled && alertEnabled {
		return fmt.Errorf("either one of \"record\" or \"alert\" must be specified as output types in json format")
	}

	recordingRuleGenerator := rule.RecordingRuleGenerator{}
	recordingRuleGroups, gCtx, err := recordingRuleGenerator.Generate(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}

	if recordEnabled {
		return common.PrintJSON(recordingRuleGroups, stdout)
	}

	alertingRuleGenerator := rule.AlertingRuleGenerator{}
	alertingRuleGroups, err := alertingRuleGenerator.Generate(gCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups")
	}

	return common.PrintJSON(alertingRuleGroups, stdout)
}

func runPrometheus(
	spec *spec.Spec,
	recordEnabled, alertEnabled bool,
	stdout io.Writer,
) error {
	recordingRuleGenerator := rule.RecordingRuleGenerator{}
	recordingRuleGroups, gCtx, err := recordingRuleGenerator.Generate(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}
	prometheusRecordingRuleGroups := recordingRuleGroups.Prometheus()

	if recordEnabled && !alertEnabled {
		return common.PrintYAML(&prometheusRecordingRuleGroups, stdout)
	}

	alertingRuleGenerator := rule.AlertingRuleGenerator{}
	alertingRuleGroups, err := alertingRuleGenerator.Generate(gCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups")
	}
	prometheusAlertingRuleGroups := alertingRuleGroups.Prometheus()

	if !recordEnabled && alertEnabled {
		return common.PrintYAML(&prometheusAlertingRuleGroups, stdout)
	}

	rgs := &rulefmt.RuleGroups{
		Groups: slices.Concat(prometheusRecordingRuleGroups.Groups, prometheusAlertingRuleGroups.Groups),
	}
	return common.PrintYAML(rgs, stdout)
}

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	var types []string
	var output string

	command := &cobra.Command{
		Use:   "prometheus-rule [-t types] [-o output] file",
		Short: "Generate SLI recording or alerting rules for Prometheus-compatible systems",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(types, output, args, cmd.OutOrStdout())
		},
	}
	command.Flags().StringArrayVarP(&types, "types", "t", []string{"record", "alert"}, "rule types to generate. Either \"record\" or \"alert\"")
	command.Flags().StringVarP(&output, "output", "o", "prometheus", "output format of generated rules. Either \"json\" or \"prometheus\"")

	return command
}

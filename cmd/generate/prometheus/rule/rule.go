package rule

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/ajalab/slogen/cmd/common"
	configspec "github.com/ajalab/slogen/internal/config/spec"
	"github.com/ajalab/slogen/internal/print"
	"github.com/ajalab/slogen/internal/prometheus/rule"
	"github.com/ajalab/slogen/internal/spec"
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
		return runJSON(spec, alertEnabled, stdout)
	case "prometheus":
		return runPrometheus(spec, alertEnabled, stdout)
	}
	return fmt.Errorf("unsupported format: %s", output)
}

func runJSON(
	spec *spec.Spec,
	alertEnabled bool,
	stdout io.Writer,
) error {
	recordingRuleGenerator := rule.RecordingRuleGenerator{}
	gCtx, err := recordingRuleGenerator.Generate(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}

	printer := print.NewJSONPrinter(stdout)
	defer printer.Close()

	if !alertEnabled {
		return printer.Print(rule.RuleGroups{Groups: gCtx.RuleGroups()})
	}

	alertingRuleGenerator := rule.AlertingRuleGenerator{}
	err = alertingRuleGenerator.Generate(gCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups")
	}

	return printer.Print(rule.RuleGroups{Groups: gCtx.RuleGroups()})
}

func runPrometheus(
	spec *spec.Spec,
	alertEnabled bool,
	stdout io.Writer,
) error {
	recordingRuleGenerator := rule.RecordingRuleGenerator{}
	gCtx, err := recordingRuleGenerator.Generate(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}

	printer := print.NewYAMLPrinter(stdout)
	defer printer.Close()

	if !alertEnabled {
		recordingRuleGroups := rule.RuleGroups{Groups: gCtx.RuleGroups()}
		prometheusRecordingRuleGroups := recordingRuleGroups.Prometheus()
		return printer.Print(&prometheusRecordingRuleGroups)
	}

	alertingRuleGenerator := rule.AlertingRuleGenerator{}
	err = alertingRuleGenerator.Generate(gCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups")
	}

	ruleGroups := rule.RuleGroups{Groups: gCtx.RuleGroups()}
	prometheusRuleGroups := ruleGroups.Prometheus()
	return printer.Print(&prometheusRuleGroups)
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

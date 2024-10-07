package rule

import (
	"fmt"
	"io"
	"os"

	"github.com/ajalab/slom/cmd/common"
	configspec "github.com/ajalab/slom/internal/config/spec/native/v1alpha"
	"github.com/ajalab/slom/internal/print"
	"github.com/ajalab/slom/internal/prometheus/rule"
	"github.com/ajalab/slom/internal/spec"
	"github.com/spf13/cobra"
)

func run(typ string, output string, args []string, stdout io.Writer) error {
	var alertEnabled bool
	switch typ {
	case "all":
		alertEnabled = true
	case "record":
		alertEnabled = false
	default:
		return fmt.Errorf("either \"all\" or \"record\" must be specified as type")
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
	g := rule.NewRuleGenerator()
	err := g.GenerateRecordingRules(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}

	printer := print.NewJSONPrinter(stdout)
	defer printer.Close()

	if !alertEnabled {
		return printer.Print(g.RuleGroups())
	}

	err = g.GenerateAlertingRules(spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups: %w", err)
	}

	return printer.Print(g.RuleGroups())
}

func runPrometheus(
	spec *spec.Spec,
	alertEnabled bool,
	stdout io.Writer,
) error {
	g := rule.NewRuleGenerator()
	err := g.GenerateRecordingRules(spec)
	if err != nil {
		return fmt.Errorf("failed to generate recording rule groups")
	}

	printer := print.NewYAMLPrinter(stdout)
	defer printer.Close()

	if !alertEnabled {
		recordingRuleGroups := g.RuleGroups()
		prometheusRecordingRuleGroups := recordingRuleGroups.Prometheus()
		return printer.Print(&prometheusRecordingRuleGroups)
	}

	err = g.GenerateAlertingRules(spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups: %w", err)
	}

	ruleGroups := g.RuleGroups()
	prometheusRuleGroups := ruleGroups.Prometheus()
	return printer.Print(&prometheusRuleGroups)
}

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	var typ string
	var output string

	command := &cobra.Command{
		Use:   "prometheus-rule [-t types] [-o output] file",
		Short: "Generate SLI recording or alerting rules for Prometheus-compatible systems",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(typ, output, args, cmd.OutOrStdout())
		},
	}
	command.Flags().StringVarP(&typ, "type", "t", "all", "rule types to generate. Either \"record\" or \"all\"")
	command.Flags().StringVarP(&output, "output", "o", "prometheus", "output format of generated rules. Either \"json\" or \"prometheus\"")

	return command
}

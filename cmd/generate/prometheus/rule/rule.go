package rule

import (
	"encoding/json"
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
	"gopkg.in/yaml.v3"
)

func run(types []string, format string, args []string, stdout io.Writer) error {
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

	switch format {
	case "json":
		return runJSON(spec, recordEnabled, alertEnabled, stdout)
	case "prometheus":
		return runPrometheus(spec, recordEnabled, alertEnabled, stdout)
	}
	return fmt.Errorf("unsupported format: %s", format)
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
		return printJSONRecordingRules(recordingRuleGroups, stdout)
	}

	alertingRuleGenerator := rule.AlertingRuleGenerator{}
	alertingRuleGroups, err := alertingRuleGenerator.Generate(gCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups")
	}

	return printJSONAlertingRules(alertingRuleGroups, stdout)
}

func printJSONRecordingRules(
	rgs *rule.RecordingRuleGroups,
	stdout io.Writer,
) error {
	e := json.NewEncoder(stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "    ")
	return e.Encode(&rgs)
}

func printJSONAlertingRules(
	rgs *rule.AlertingRuleGroups,
	stdout io.Writer,
) error {
	e := json.NewEncoder(stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "    ")
	return e.Encode(&rgs)
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
		return printPrometheusRules(&prometheusRecordingRuleGroups, stdout)
	}

	alertingRuleGenerator := rule.AlertingRuleGenerator{}
	alertingRuleGroups, err := alertingRuleGenerator.Generate(gCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to generate alerting rule groups")
	}
	prometheusAlertingRuleGroups := alertingRuleGroups.Prometheus()

	if !recordEnabled && alertEnabled {
		return printPrometheusRules(&prometheusAlertingRuleGroups, stdout)
	}

	rgs := &rulefmt.RuleGroups{
		Groups: slices.Concat(prometheusRecordingRuleGroups.Groups, prometheusAlertingRuleGroups.Groups),
	}
	return printPrometheusRules(rgs, stdout)
}

func printPrometheusRules(
	rgs *rulefmt.RuleGroups,
	stdout io.Writer,
) error {
	e := yaml.NewEncoder(stdout)
	defer e.Close()

	return e.Encode(rgs)
}

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	var types []string
	var format string

	command := &cobra.Command{
		Use:   "prometheus-rule [-t types] [-f format] file",
		Short: "Generate SLI recording or alerting rules for Prometheus-compatible systems",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(types, format, args, cmd.OutOrStdout())
		},
	}
	command.Flags().StringArrayVarP(&types, "types", "t", []string{"record", "alert"}, "rule types to generate. Either \"record\" or \"alert\"")
	command.Flags().StringVarP(&format, "format", "f", "prometheus", "output format of generated rules. Either \"json\" or \"prometheus\"")

	return command
}

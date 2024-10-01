package rule

import (
	"strings"

	"github.com/ajalab/slom/internal/spec"
)

const labelNameSpec string = "slom_spec"
const labelNameSLO string = "slom_slo"
const labelNameId string = "slom_id"

func sloId(specName string, sloName string) string {
	return specName + "-" + sloName
}

func metricNameErrorRate(
	levels []string,
	duration spec.Duration,
) string {
	return metricNamePrefix(levels) + "slom_error:ratio_rate" + duration.String()
}

func metricNameErrorBudget(
	levels []string,
	duration spec.Duration,
) string {
	return metricNamePrefix(levels) + "slom_error_budget:ratio_rate" + duration.String()
}

func metricNamePrefix(levels []string) string {
	if len(levels) > 0 {
		return strings.Join(levels, "_") + ":"
	}
	return ""
}

const metricNameSLO string = "slom_slo"

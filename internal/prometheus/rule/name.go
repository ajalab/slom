package rule

import (
	"strings"

	"github.com/ajalab/slogen/internal/spec"
)

const labelNameSpec string = "slogen_spec"
const labelNameSLO string = "slogen_slo"
const labelNameId string = "slogen_id"

func sloId(specName string, sloName string) string {
	return specName + "-" + sloName
}

func metricNameErrorRate(
	levels []string,
	duration spec.Duration,
) string {
	return metricNamePrefix(levels) + "slogen_error:ratio_rate" + duration.String()
}

func metricNameErrorBudget(
	levels []string,
	duration spec.Duration,
) string {
	return metricNamePrefix(levels) + "slogen_error_budget:ratio_rate" + duration.String()
}

func metricNamePrefix(levels []string) string {
	if len(levels) > 0 {
		return strings.Join(levels, "_") + ":"
	}
	return ""
}

const metricNameSLO string = "slogen_slo"

package rule

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ajalab/slogen/internal/spec"
)

type RecordingRuleGenerator struct{}

func (g *RecordingRuleGenerator) Generate(
	s *spec.Spec,
) (*PrometheusGeneratorContext, error) {
	gCtx := NewPrometheusGeneratorContext()

	for _, slo := range s.SLOs() {
		err := g.generateRecordingRuleGroups(gCtx, s.Name(), slo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recording rule groups for SLO %s: %w", slo.Name(), err)
		}
	}

	return gCtx, nil
}

func (g *RecordingRuleGenerator) generateRecordingRuleGroups(
	gCtx *PrometheusGeneratorContext,
	specName string,
	slo *spec.SLO,
) error {
	id := sloId(specName, slo.Name())

	labels := map[string]string{
		labelNameSpec: specName,
		labelNameSLO:  slo.Name(),
		labelNameId:   id,
	}

	indicator, ok := slo.Indicator().(*spec.PrometheusIndicator)
	if !ok {
		return fmt.Errorf("no Prometheus indicator found")
	}

	for _, w := range slo.Windows() {
		ruleGroupName := "slogen:" + id + ":default"
		ruleErrorRate := g.generateErrorRateRecordingRule(indicator, w, labels)
		gCtx.addErrorRateRecordingRule(ruleGroupName, slo.Name(), w.Name(), ruleErrorRate)
	}

	ruleErrorBudget, err := g.generateErrorBudgetRecordingRule(gCtx, indicator, slo.Name(), *slo.Objective(), labels)
	if err != nil {
		return fmt.Errorf("failed to generate error budget recording rule: %w", err)
	}
	if ruleErrorBudget != nil {
		ruleGroupNameErrorBudget := "slogen:" + id + ":default"
		gCtx.addErrorBudgetRecordingRule(ruleGroupNameErrorBudget, slo.Name(), *ruleErrorBudget)
	}

	ruleGroupNameMeta := "slogen:" + id + ":meta"
	ruleMeta := &RecordingRule{
		Record: metricNameSLO,
		Expr:   strconv.FormatFloat(slo.Objective().Ratio(), 'f', -1, 64),
		Labels: labels,
	}
	gCtx.addMetaRecordingRule(ruleGroupNameMeta, ruleMeta)

	return nil
}

var reWindow = regexp.MustCompile(`\$window\b`)

func (g *RecordingRuleGenerator) generateErrorRateRecordingRule(
	indicator *spec.PrometheusIndicator,
	window spec.Window,
	labels map[string]string,
) RecordingRule {
	name := metricNameErrorRate(indicator.Level(), window.Duration())
	expr := GenerateErrorRateQuery(indicator, window)

	return RecordingRule{
		Record: name,
		Expr:   expr,
		Labels: labels,
	}
}

func (g *RecordingRuleGenerator) generateErrorBudgetRecordingRule(
	gCtx *PrometheusGeneratorContext,
	indicator *spec.PrometheusIndicator,
	sloName string,
	objective spec.Objective,
	labels map[string]string,
) (*RecordingRule, error) {
	sloWindow := objective.Window()
	if sloWindow == nil {
		return nil, nil
	}

	name := metricNameErrorBudget(indicator.Level(), sloWindow.Duration())
	errorRateRule, err := gCtx.getErrorRateRecordingRule(sloName, sloWindow.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to get error rate recording rule: %w", err)
	}

	var expr string
	switch sloWindow.(type) {
	case *spec.RollingWindow:
		expr = fmt.Sprintf(
			"1 - (%s) / (1 - %g)",
			errorRateRule.Expr,
			// prometheus.GenerateLabels(labels, true),
			objective.Ratio(),
		)

	}
	return &RecordingRule{
		Record: name,
		Expr:   expr,
		Labels: labels,
	}, nil
}

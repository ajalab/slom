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
) (*RecordingRuleGroups, *PrometheusGeneratorContext, error) {
	gCtx := NewPrometheusGeneratorContext()

	var ruleGroups []RecordingRuleGroup

	for _, slo := range s.SLOs() {
		rgs, err := g.generateRecordingRuleGroups(gCtx, s.Name(), slo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate recording rule groups for SLO %s: %w", slo.Name(), err)
		}

		ruleGroups = append(ruleGroups, rgs...)
	}

	return &RecordingRuleGroups{ruleGroups}, gCtx, nil
}

func (g *RecordingRuleGenerator) generateRecordingRuleGroups(
	gCtx *PrometheusGeneratorContext,
	specName string,
	slo *spec.SLO,
) ([]RecordingRuleGroup, error) {
	id := sloId(specName, slo.Name())

	labels := map[string]string{
		labelNameSpec: specName,
		labelNameSLO:  slo.Name(),
		labelNameId:   id,
	}

	indicator, ok := slo.Indicator().(*spec.PrometheusIndicator)
	if !ok {
		return nil, fmt.Errorf("no Prometheus indicator found")
	}

	recordingRuleGroup := RecordingRuleGroup{
		Name: "slogen:" + id + ":recording",
	}
	for _, w := range slo.Windows() {
		ruleErrorRate := g.generateErrorRateRecordingRule(indicator, w, labels)
		recordingRuleGroup.Rules = append(recordingRuleGroup.Rules, ruleErrorRate)
		gCtx.addErrorRateRecordingRule(slo.Name(), w.Name(), ruleErrorRate)
	}
	ruleErrorBudget, err := g.generateErrorBudgetRecordingRule(gCtx, indicator, slo.Name(), *slo.Objective(), labels)
	if err != nil {
		return nil, fmt.Errorf("failed to generate error budget recording rule: %w", err)
	}
	recordingRuleGroup.Rules = append(recordingRuleGroup.Rules, ruleErrorBudget)
	gCtx.addErrorBudgetRecordingRule(slo.Name(), ruleErrorBudget)

	recordingMetaRuleGroup := RecordingRuleGroup{
		Name: "slogen:" + id + ":recording-meta",
		Rules: []RecordingRule{
			{
				Record: metricNameSLO,
				Expr:   strconv.FormatFloat(slo.Objective().Ratio(), 'f', -1, 64),
				Labels: labels,
			},
		},
	}

	return []RecordingRuleGroup{recordingRuleGroup, recordingMetaRuleGroup}, nil
}

var reWindow = regexp.MustCompile(`\$window\b`)

func (g *RecordingRuleGenerator) generateErrorRateRecordingRule(
	indicator *spec.PrometheusIndicator,
	window spec.Window,
	labels map[string]string,
) RecordingRule {
	name := metricNameErrorRate(indicator.Level(), window.Duration())

	var expr string
	switch w := window.(type) {
	case *spec.RollingWindow:
		expr = reWindow.ReplaceAllString(indicator.ErrorRatio(), w.Duration().String())
	case *spec.CalendarWindow:
		expr = ""
	}

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
) (RecordingRule, error) {
	sloWindow := objective.Window()

	name := metricNameErrorBudget(indicator.Level(), sloWindow.Duration())
	errorRateRule, err := gCtx.getErrorRateRecordingRule(sloName, sloWindow.Name())
	if err != nil {
		return RecordingRule{}, fmt.Errorf("failed to get error rate recording rule: %w", err)
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
	return RecordingRule{
		Record: name,
		Expr:   expr,
		Labels: labels,
	}, nil
}

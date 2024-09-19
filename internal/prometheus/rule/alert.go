package rule

import (
	"fmt"

	"github.com/ajalab/slogen/internal/spec"
)

type AlertingRuleGenerator struct{}

func (g *AlertingRuleGenerator) Generate(
	gCtx *PrometheusGeneratorContext,
	s *spec.Spec,
) (*RuleGroups, error) {
	var ruleGroups []RuleGroup

	for _, slo := range s.SLOs() {
		rgs, err := g.generateAlertingRuleGroups(gCtx, s.Name(), slo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate alerting rule groups for SLO %s: %w", slo.Name(), err)
		}

		ruleGroups = append(ruleGroups, rgs...)
	}

	return &RuleGroups{ruleGroups}, nil
}

func (g *AlertingRuleGenerator) generateAlertingRuleGroups(
	gCtx *PrometheusGeneratorContext,
	specName string,
	slo *spec.SLO,
) ([]RuleGroup, error) {
	if len(slo.Alerts()) == 0 {
		return nil, nil
	}

	id := sloId(specName, slo.Name())

	ruleGroup := RuleGroup{
		Name: "slogen:" + id + ":alerting",
	}

	for _, a := range slo.Alerts() {
		var rule AlertingRule
		var err error
		switch a := a.(type) {
		case *spec.BurnRateAlert:
			rule, err = g.generateBurnRateAlertingRule(gCtx, specName, slo.Name(), slo.Objective(), a)
		case *spec.BreachAlert:
			rule, err = g.generateBreachAlertingRule(gCtx, specName, slo.Name(), a)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to generate alerting rule for alert %s: %w", a.Name(), err)
		}
		ruleGroup.Rules = append(ruleGroup.Rules, &rule)
	}

	return []RuleGroup{ruleGroup}, nil
}

func (g *AlertingRuleGenerator) generateBurnRateAlertingRule(
	gCtx *PrometheusGeneratorContext,
	specName string,
	sloName string,
	objective *spec.Objective,
	a *spec.BurnRateAlert,
) (AlertingRule, error) {
	id := sloId(specName, sloName)

	alerter, ok := a.Alerter().(*spec.PrometheusAlerter)
	if !ok {
		return AlertingRule{}, fmt.Errorf("only prometheus alerter is supported")
	}

	burnRateThreshold := a.ConsumedBudgetRatio() * float64(objective.Window().Duration()) / float64(a.Window().Window().Duration())
	errorRateThreshold := fmt.Sprintf("%g * %g", burnRateThreshold, 1-objective.Ratio())

	var expr string
	switch w := a.Window().(type) {
	case *spec.BurnRateAlertMultiWindows:
		errorRateRuleShort := gCtx.errorRateRecordingRules[sloName][w.ShortWindow().Name()]
		errorRateRuleLong := gCtx.errorRateRecordingRules[sloName][w.LongWindow().Name()]

		errorRateQueryShort := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRuleShort.Record, labelNameId, id)
		errorRateQueryLong := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRuleLong.Record, labelNameId, id)

		expr = fmt.Sprintf("%[1]s > %[3]s and %[2]s > %[3]s", errorRateQueryLong, errorRateQueryShort, errorRateThreshold)
	}

	return AlertingRule{
		Alert:       alerter.Name(),
		Expr:        expr,
		Labels:      alerter.Labels(),
		Annotations: alerter.Annotations(),
	}, nil
}

func (g *AlertingRuleGenerator) generateBreachAlertingRule(
	gCtx *PrometheusGeneratorContext,
	specName string,
	sloName string,
	a *spec.BreachAlert,
) (AlertingRule, error) {
	id := sloId(specName, sloName)

	alerter, ok := a.Alerter().(*spec.PrometheusAlerter)
	if !ok {
		return AlertingRule{}, fmt.Errorf("only prometheus alerter is supported")
	}

	errorBudgetRule, err := gCtx.getErrorBudgetRecordingRule(sloName)
	if err != nil {
		return AlertingRule{}, err
	}
	expr := fmt.Sprintf("%s{%s=\"%s\"} <= 0", errorBudgetRule.Record, labelNameId, id)

	return AlertingRule{
		Alert:       alerter.Name(),
		Expr:        expr,
		Labels:      alerter.Labels(),
		Annotations: alerter.Annotations(),
	}, nil
}

package rule

import (
	"fmt"
	"strconv"

	"github.com/ajalab/slogen/internal/spec"
)

type RuleGenerator struct {
	ruleGroups       []*RuleGroup
	ruleGroupsByName map[string]*RuleGroup

	// sloName → windowName → rule
	errorRateRecordingRules map[string]map[string]*RecordingRule

	// sloName → rule
	errorBudgetRecordingRules map[string]*RecordingRule
}

func NewRuleGenerator() *RuleGenerator {
	return &RuleGenerator{
		ruleGroups:                nil,
		ruleGroupsByName:          map[string]*RuleGroup{},
		errorRateRecordingRules:   map[string]map[string]*RecordingRule{},
		errorBudgetRecordingRules: map[string]*RecordingRule{},
	}
}

func (g *RuleGenerator) getOrCreateRuleGroup(
	ruleGroupName string,
) *RuleGroup {
	ruleGroup, ok := g.ruleGroupsByName[ruleGroupName]
	if !ok {
		ruleGroup = &RuleGroup{
			Name: ruleGroupName,
		}
		g.ruleGroupsByName[ruleGroupName] = ruleGroup
		g.ruleGroups = append(g.ruleGroups, ruleGroup)
	}

	return ruleGroup
}

func (g *RuleGenerator) GenerateRecordingRules(
	s *spec.Spec,
) error {
	for _, slo := range s.SLOs() {
		err := g.generateRecordingRules(s.Name(), slo)
		if err != nil {
			return fmt.Errorf("failed to generate recording rules for SLO %s: %w", slo.Name(), err)
		}
	}
	return nil
}

func (g *RuleGenerator) generateRecordingRules(
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
		return fmt.Errorf("only prometheus indicator is supported")
	}

	for _, w := range slo.Windows() {
		ruleGroupName := "slogen:" + id + ":default"
		ruleErrorRate := g.generateErrorRateRecordingRule(indicator, w, labels)
		g.addErrorRateRecordingRule(ruleGroupName, slo.Name(), w.Name(), ruleErrorRate)
	}

	sloWindow := slo.Objective().Window()
	if sloWindow != nil {
		ruleErrorBudget, err := g.generateErrorBudgetRecordingRule(indicator, slo.Name(), slo.Objective(), labels)
		if err != nil {
			return fmt.Errorf("failed to generate error budget recording rule: %w", err)
		}
		ruleGroupNameErrorBudget := "slogen:" + id + ":default"
		g.addErrorBudgetRecordingRule(ruleGroupNameErrorBudget, slo.Name(), ruleErrorBudget)
	}

	ruleGroupNameMeta := "slogen:" + id + ":meta"
	ruleMeta := &RecordingRule{
		Record: metricNameSLO,
		Expr:   strconv.FormatFloat(slo.Objective().Ratio(), 'f', -1, 64),
		Labels: labels,
	}
	g.addMetaRecordingRule(ruleGroupNameMeta, ruleMeta)

	return nil
}

func (g *RuleGenerator) generateErrorRateRecordingRule(
	indicator *spec.PrometheusIndicator,
	window spec.Window,
	labels map[string]string,
) *RecordingRule {
	name := metricNameErrorRate(indicator.Level(), window.Duration())
	expr := generateErrorRateQuery(indicator, window)

	return &RecordingRule{
		Record: name,
		Expr:   expr,
		Labels: labels,
	}
}

func (g *RuleGenerator) addErrorRateRecordingRule(
	ruleGroupName string,
	sloName string,
	windowName string,
	r *RecordingRule,
) {
	ruleGroup := g.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, r)

	rules, ok := g.errorRateRecordingRules[sloName]
	if !ok {
		rules = make(map[string]*RecordingRule)
		g.errorRateRecordingRules[sloName] = rules
	}

	rules[windowName] = r
}

func (g *RuleGenerator) getErrorRateRecordingRule(
	sloName string,
	windowName string,
) (*RecordingRule, error) {
	rules, ok := g.errorRateRecordingRules[sloName]
	if !ok {
		return nil, fmt.Errorf("recording rules for SLO %s were not generated", sloName)
	}
	rule, ok := rules[windowName]
	if !ok {
		return nil, fmt.Errorf("recording rule with windowName %s for SLO %s was not generated", windowName, sloName)
	}
	return rule, nil
}

func (g *RuleGenerator) generateErrorBudgetRecordingRule(
	indicator *spec.PrometheusIndicator,
	sloName string,
	objective *spec.Objective,
	labels map[string]string,
) (*RecordingRule, error) {
	sloWindow := objective.Window()
	name := metricNameErrorBudget(indicator.Level(), sloWindow.Duration())

	errorRateRule, err := g.getErrorRateRecordingRule(sloName, sloWindow.Name())
	if err != nil {
		return nil, fmt.Errorf("could not find an error rate recording rule for error budget recording rule: %w", err)
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

func (g *RuleGenerator) addErrorBudgetRecordingRule(
	ruleGroupName string,
	sloName string,
	r *RecordingRule,
) {
	ruleGroup := g.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, r)

	g.errorBudgetRecordingRules[sloName] = r
}

func (g *RuleGenerator) getErrorBudgetRecordingRule(
	sloName string,
) (*RecordingRule, error) {
	rule, ok := g.errorBudgetRecordingRules[sloName]
	if !ok {
		return nil, fmt.Errorf("error budget recording rule with for SLO %s was not generated", sloName)
	}
	return rule, nil
}

func (g *RuleGenerator) addMetaRecordingRule(
	ruleGroupName string,
	r *RecordingRule,
) {
	ruleGroup := g.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, r)
}

func (g *RuleGenerator) GenerateAlertingRules(
	s *spec.Spec,
) error {
	for _, slo := range s.SLOs() {
		err := g.generateAlertingRules(s.Name(), slo)
		if err != nil {
			return fmt.Errorf("failed to generate alerting rules for SLO %s: %w", slo.Name(), err)
		}
	}
	return nil
}

func (g *RuleGenerator) generateAlertingRules(
	specName string,
	slo *spec.SLO,
) error {
	if len(slo.Alerts()) == 0 {
		return nil
	}

	id := sloId(specName, slo.Name())

	for _, a := range slo.Alerts() {
		var rule *AlertingRule
		var err error
		switch a := a.(type) {
		case *spec.BurnRateAlert:
			rule, err = g.generateBurnRateAlertingRule(specName, slo.Name(), slo.Objective(), a)
		case *spec.BreachAlert:
			rule, err = g.generateBreachAlertingRule(specName, slo.Name(), a)
		}

		if err != nil {
			return fmt.Errorf("failed to generate alerting rule for alert %s: %w", a.Name(), err)
		}
		ruleGroupName := "slogen:" + id + ":default"
		g.addAlertingRule(ruleGroupName, rule)
	}

	return nil
}

func (g *RuleGenerator) generateBurnRateAlertingRule(
	specName string,
	sloName string,
	objective *spec.Objective,
	a *spec.BurnRateAlert,
) (*AlertingRule, error) {
	id := sloId(specName, sloName)

	alerter, ok := a.Alerter().(*spec.PrometheusAlerter)
	if !ok {
		return nil, fmt.Errorf("only prometheus alerter is supported")
	}

	burnRateThreshold := a.ConsumedBudgetRatio() * float64(objective.Window().Duration()) / float64(a.Window().Window().Duration())
	errorRateThreshold := fmt.Sprintf("%g * %g", burnRateThreshold, 1-objective.Ratio())

	var expr string
	switch w := a.Window().(type) {
	case *spec.BurnRateAlertMultiWindows:
		shortWindowName := w.ShortWindow().Name()
		longWindowName := w.LongWindow().Name()
		errorRateRuleShort, err := g.getErrorRateRecordingRule(sloName, shortWindowName)
		if err != nil {
			return nil, fmt.Errorf("could not find an error rate recording rule for the short window of the burn rate alert rule: %w", err)
		}
		errorRateRuleLong, err := g.getErrorRateRecordingRule(sloName, longWindowName)
		if err != nil {
			return nil, fmt.Errorf("could not find an error rate recording rule for the long window of the burn rate alert rule: %w", err)
		}

		errorRateQueryShort := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRuleShort.Record, labelNameId, id)
		errorRateQueryLong := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRuleLong.Record, labelNameId, id)

		expr = fmt.Sprintf("%[1]s > %[3]s and %[2]s > %[3]s", errorRateQueryLong, errorRateQueryShort, errorRateThreshold)
	}

	return &AlertingRule{
		Alert:       alerter.Name(),
		Expr:        expr,
		Labels:      alerter.Labels(),
		Annotations: alerter.Annotations(),
	}, nil
}

func (g *RuleGenerator) generateBreachAlertingRule(
	specName string,
	sloName string,
	a *spec.BreachAlert,
) (*AlertingRule, error) {
	id := sloId(specName, sloName)

	alerter, ok := a.Alerter().(*spec.PrometheusAlerter)
	if !ok {
		return nil, fmt.Errorf("only prometheus alerter is supported")
	}

	errorBudgetRule, err := g.getErrorBudgetRecordingRule(sloName)
	if err != nil {
		return nil, fmt.Errorf("could not find an error budget recording rule for the breach alert rule: %w", err)
	}
	expr := fmt.Sprintf("%s{%s=\"%s\"} <= 0", errorBudgetRule.Record, labelNameId, id)

	return &AlertingRule{
		Alert:       alerter.Name(),
		Expr:        expr,
		Labels:      alerter.Labels(),
		Annotations: alerter.Annotations(),
	}, nil
}

func (g *RuleGenerator) addAlertingRule(
	ruleGroupName string,
	r *AlertingRule,
) {
	ruleGroup := g.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, r)
}

func (g *RuleGenerator) RuleGroups() *RuleGroups {
	return &RuleGroups{Groups: g.ruleGroups}
}

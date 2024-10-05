package rule

import (
	"fmt"
	"strconv"

	"github.com/ajalab/slom/internal/spec"
)

type ruleGroupKind int

const (
	ruleGroupRecord = iota + 1
	ruleGroupAlert
	ruleGroupMeta
)

type RuleGenerator struct {
	ruleGroups       []*RuleGroup
	ruleGroupsByName map[string]*RuleGroup

	// sloId → windowName → rule
	errorRateRecordingRules map[string]map[string]*RecordingRule

	// sloId → rule
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
	sloId string,
	ruleGroupKind ruleGroupKind,
	evaluationInterval spec.Duration,
) *RuleGroup {
	var ruleGroupName string
	if ruleGroupKind == ruleGroupMeta {
		ruleGroupName = fmt.Sprintf("slom:%s:meta", sloId)
	} else if evaluationInterval == 0 {
		ruleGroupName = fmt.Sprintf("slom:%s:default", sloId)
	} else {
		ruleGroupName = fmt.Sprintf("slom:%s:%s", sloId, evaluationInterval.String())
	}

	ruleGroup, ok := g.ruleGroupsByName[ruleGroupName]
	if !ok {
		ruleGroup = &RuleGroup{
			Name:     ruleGroupName,
			Interval: evaluationInterval,
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
		ruleErrorRate := g.generateErrorRateRecordingRule(indicator, w, labels)
		g.addErrorRateRecordingRule(id, w.Name(), ruleErrorRate, w.Prometheus().EvaluationInterval())
	}

	sloWindow := slo.Objective().Window()
	if sloWindow != nil {
		ruleErrorBudget, err := g.generateErrorBudgetRecordingRule(indicator, id, slo.Objective(), labels)
		if err != nil {
			return fmt.Errorf("failed to generate error budget recording rule: %w", err)
		}
		g.addErrorBudgetRecordingRule(id, ruleErrorBudget, sloWindow.Prometheus().EvaluationInterval())
	}

	ruleMeta := &RecordingRule{
		Record: metricNameSLO,
		Expr:   strconv.FormatFloat(slo.Objective().Ratio(), 'f', -1, 64),
		Labels: labels,
	}
	g.addMetaRecordingRule(id, ruleMeta, spec.Duration(0))

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
	sloId string,
	windowName string,
	r *RecordingRule,
	evaluationInterval spec.Duration,
) {
	ruleGroup := g.getOrCreateRuleGroup(sloId, ruleGroupRecord, evaluationInterval)
	ruleGroup.Rules = append(ruleGroup.Rules, r)

	rules, ok := g.errorRateRecordingRules[sloId]
	if !ok {
		rules = make(map[string]*RecordingRule)
		g.errorRateRecordingRules[sloId] = rules
	}

	rules[windowName] = r
}

func (g *RuleGenerator) getErrorRateRecordingRule(
	sloId string,
	windowName string,
) (*RecordingRule, error) {
	rules, ok := g.errorRateRecordingRules[sloId]
	if !ok {
		return nil, fmt.Errorf("recording rules for SLO %s were not generated", sloId)
	}
	rule, ok := rules[windowName]
	if !ok {
		return nil, fmt.Errorf("recording rule with windowName %s for SLO %s was not generated", windowName, sloId)
	}
	return rule, nil
}

func (g *RuleGenerator) generateErrorBudgetRecordingRule(
	indicator *spec.PrometheusIndicator,
	sloId string,
	objective *spec.Objective,
	labels map[string]string,
) (*RecordingRule, error) {
	sloWindow := objective.Window()
	name := metricNameErrorBudget(indicator.Level(), sloWindow.Duration())

	errorRateRule, err := g.getErrorRateRecordingRule(sloId, sloWindow.Name())
	if err != nil {
		return nil, fmt.Errorf("could not find an error rate recording rule for error budget recording rule: %w", err)
	}

	var expr string
	switch sloWindow.(type) {
	case *spec.RollingWindow:
		expr = fmt.Sprintf(
			"1 - %s{%s=\"%s\"} / (1 - %g)",
			errorRateRule.Record,
			labelNameId,
			sloId,
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
	sloId string,
	r *RecordingRule,
	evaluationInterval spec.Duration,
) {
	ruleGroup := g.getOrCreateRuleGroup(sloId, ruleGroupRecord, evaluationInterval)
	ruleGroup.Rules = append(ruleGroup.Rules, r)

	g.errorBudgetRecordingRules[sloId] = r
}

func (g *RuleGenerator) getErrorBudgetRecordingRule(
	sloId string,
) (*RecordingRule, error) {
	rule, ok := g.errorBudgetRecordingRules[sloId]
	if !ok {
		return nil, fmt.Errorf("error budget recording rule with for SLO %s was not generated", sloId)
	}
	return rule, nil
}

func (g *RuleGenerator) addMetaRecordingRule(
	sloId string,
	r *RecordingRule,
	evaluationInterval spec.Duration,
) {
	ruleGroup := g.getOrCreateRuleGroup(sloId, ruleGroupMeta, evaluationInterval)
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
		var window spec.Window
		switch a := a.(type) {
		case *spec.BurnRateAlert:
			rule, err = g.generateBurnRateAlertingRule(id, slo.Objective(), a)
			window = a.Window().Window()
		case *spec.ErrorBudgetAlert:
			rule, err = g.generateErrorBudgetAlertingRule(id, a)
			window = slo.Objective().Window()
		}

		if err != nil {
			return fmt.Errorf("failed to generate alerting rule for alert %s: %w", a.Name(), err)
		}
		g.addAlertingRule(id, rule, window.Prometheus().EvaluationInterval())
	}

	return nil
}

func (g *RuleGenerator) generateBurnRateAlertingRule(
	sloId string,
	objective *spec.Objective,
	a *spec.BurnRateAlert,
) (*AlertingRule, error) {
	alerter, ok := a.Alerter().(*spec.PrometheusAlerter)
	if !ok {
		return nil, fmt.Errorf("only prometheus alerter is supported")
	}

	sloWindow := objective.Window()
	if sloWindow == nil {
		return nil, fmt.Errorf("SLO window is not defined")
	}

	burnRateThreshold := a.ConsumedBudgetRatio() * float64(sloWindow.Duration()) / float64(a.Window().Window().Duration())
	errorRateThreshold := fmt.Sprintf("%g * %g", burnRateThreshold, 1-objective.Ratio())

	var expr string
	switch w := a.Window().(type) {
	case *spec.BurnRateAlertSingleWindow:
		windowName := w.Window().Name()
		errorRateRule, err := g.getErrorRateRecordingRule(sloId, windowName)
		if err != nil {
			return nil, fmt.Errorf("could not find an error rate recording rule for the window of the burn rate alert rule: %w", err)
		}

		errorRateQuery := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRule.Record, labelNameId, sloId)
		expr = fmt.Sprintf("%s > %s", errorRateQuery, errorRateThreshold)

	case *spec.BurnRateAlertMultiWindows:
		shortWindowName := w.ShortWindow().Name()
		longWindowName := w.LongWindow().Name()
		errorRateRuleShort, err := g.getErrorRateRecordingRule(sloId, shortWindowName)
		if err != nil {
			return nil, fmt.Errorf("could not find an error rate recording rule for the short window of the burn rate alert rule: %w", err)
		}
		errorRateRuleLong, err := g.getErrorRateRecordingRule(sloId, longWindowName)
		if err != nil {
			return nil, fmt.Errorf("could not find an error rate recording rule for the long window of the burn rate alert rule: %w", err)
		}

		errorRateQueryShort := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRuleShort.Record, labelNameId, sloId)
		errorRateQueryLong := fmt.Sprintf("%s{%s=\"%s\"}", errorRateRuleLong.Record, labelNameId, sloId)
		expr = fmt.Sprintf("%[1]s > %[3]s and %[2]s > %[3]s", errorRateQueryLong, errorRateQueryShort, errorRateThreshold)
	}

	return &AlertingRule{
		Alert:       alerter.Name(),
		Expr:        expr,
		Labels:      alerter.Labels(),
		Annotations: alerter.Annotations(),
	}, nil
}

func (g *RuleGenerator) generateErrorBudgetAlertingRule(
	sloId string,
	a *spec.ErrorBudgetAlert,
) (*AlertingRule, error) {
	alerter, ok := a.Alerter().(*spec.PrometheusAlerter)
	if !ok {
		return nil, fmt.Errorf("only prometheus alerter is supported")
	}

	errorBudgetRule, err := g.getErrorBudgetRecordingRule(sloId)
	if err != nil {
		return nil, fmt.Errorf("could not find an error budget recording rule for the error budget alert rule: %w", err)
	}
	expr := fmt.Sprintf("%s{%s=\"%s\"} <= 1 - %g", errorBudgetRule.Record, labelNameId, sloId, a.ConsumedBudgetRatio())

	return &AlertingRule{
		Alert:       alerter.Name(),
		Expr:        expr,
		Labels:      alerter.Labels(),
		Annotations: alerter.Annotations(),
	}, nil
}

func (g *RuleGenerator) addAlertingRule(
	sloId string,
	r *AlertingRule,
	evaluationInterval spec.Duration,
) {
	ruleGroup := g.getOrCreateRuleGroup(sloId, ruleGroupAlert, evaluationInterval)
	ruleGroup.Rules = append(ruleGroup.Rules, r)
}

func (g *RuleGenerator) RuleGroups() *RuleGroups {
	return &RuleGroups{Groups: g.ruleGroups}
}

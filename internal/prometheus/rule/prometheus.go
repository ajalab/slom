package rule

import "fmt"

type PrometheusGeneratorContext struct {
	ruleGroups       []*RuleGroup
	ruleGroupsByName map[string]*RuleGroup

	// sloName → windowName → rule
	errorRateRecordingRules map[string]map[string]RecordingRule

	// sloName → rule
	errorBudgetRecordingRules map[string]RecordingRule
}

func NewPrometheusGeneratorContext() *PrometheusGeneratorContext {
	return &PrometheusGeneratorContext{
		ruleGroups:                nil,
		ruleGroupsByName:          map[string]*RuleGroup{},
		errorRateRecordingRules:   map[string]map[string]RecordingRule{},
		errorBudgetRecordingRules: map[string]RecordingRule{},
	}
}

func (gCtx *PrometheusGeneratorContext) getOrCreateRuleGroup(
	ruleGroupName string,
) *RuleGroup {
	ruleGroup, ok := gCtx.ruleGroupsByName[ruleGroupName]
	if !ok {
		ruleGroup = &RuleGroup{
			Name: ruleGroupName,
		}
		gCtx.ruleGroupsByName[ruleGroupName] = ruleGroup
		gCtx.ruleGroups = append(gCtx.ruleGroups, ruleGroup)
	}

	return ruleGroup
}

func (gCtx *PrometheusGeneratorContext) addErrorRateRecordingRule(
	ruleGroupName string,
	sloName string,
	windowName string,
	r RecordingRule,
) {
	ruleGroup := gCtx.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, &r)

	rules, ok := gCtx.errorRateRecordingRules[sloName]
	if !ok {
		rules = make(map[string]RecordingRule)
		gCtx.errorRateRecordingRules[sloName] = rules
	}

	rules[windowName] = r
}

func (gCtx *PrometheusGeneratorContext) getErrorRateRecordingRule(
	sloName string,
	windowName string,
) (RecordingRule, error) {
	rules, ok := gCtx.errorRateRecordingRules[sloName]
	if !ok {
		return RecordingRule{}, fmt.Errorf("recording rules for SLO %s were not generated", sloName)
	}
	rule, ok := rules[windowName]
	if !ok {
		return RecordingRule{}, fmt.Errorf("recording rule with windowName %s for SLO %s was not generated", windowName, sloName)
	}
	return rule, nil
}

func (gCtx *PrometheusGeneratorContext) addErrorBudgetRecordingRule(
	ruleGroupName string,
	sloName string,
	r RecordingRule,
) {
	ruleGroup := gCtx.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, &r)

	gCtx.errorBudgetRecordingRules[sloName] = r
}

func (gCtx *PrometheusGeneratorContext) getErrorBudgetRecordingRule(
	sloName string,
) (RecordingRule, error) {
	rule, ok := gCtx.errorBudgetRecordingRules[sloName]
	if !ok {
		return RecordingRule{}, fmt.Errorf("error budget recording rule with for SLO %s was not generated", sloName)
	}
	return rule, nil
}

func (gCtx *PrometheusGeneratorContext) addMetaRecordingRule(
	ruleGroupName string,
	r *RecordingRule,
) {
	ruleGroup := gCtx.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, r)
}

func (gCtx *PrometheusGeneratorContext) RuleGroups() []*RuleGroup {
	return gCtx.ruleGroups
}

func (gCtx *PrometheusGeneratorContext) addAlertingRule(
	ruleGroupName string,
	r *AlertingRule,
) {
	ruleGroup := gCtx.getOrCreateRuleGroup(ruleGroupName)
	ruleGroup.Rules = append(ruleGroup.Rules, r)
}

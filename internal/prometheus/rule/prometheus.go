package rule

import "fmt"

type PrometheusGeneratorContext struct {
	// sloName → windowName → rule
	errorRateRecordingRules map[string]map[string]RecordingRule

	// sloName → rule
	errorBudgetRecordingRules map[string]RecordingRule
}

func NewPrometheusGeneratorContext() *PrometheusGeneratorContext {
	return &PrometheusGeneratorContext{
		errorRateRecordingRules:   map[string]map[string]RecordingRule{},
		errorBudgetRecordingRules: map[string]RecordingRule{},
	}
}

func (gCtx *PrometheusGeneratorContext) addErrorRateRecordingRule(
	sloName string,
	windowName string,
	r RecordingRule,
) {
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
	sloName string,
	r RecordingRule,
) {
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

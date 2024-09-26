package rule

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

type RuleGroups struct {
	Groups []*RuleGroup `json:"groups" yaml:"rules"`
}

func (rgs RuleGroups) Prometheus() rulefmt.RuleGroups {
	ruleGroups := rulefmt.RuleGroups{}
	for _, rg := range rgs.Groups {
		ruleGroups.Groups = append(ruleGroups.Groups, rg.Prometheus())
	}
	return ruleGroups
}

type RuleGroup struct {
	Name     string         `json:"name" yaml:"name"`
	Interval model.Duration `json:"interval" yaml:"interval"`
	Rules    []Rule         `json:"rules" yaml:"rules"`
}

func (rg RuleGroup) Prometheus() rulefmt.RuleGroup {
	ruleGroup := rulefmt.RuleGroup{
		Name:     rg.Name,
		Interval: rg.Interval,
	}
	for _, r := range rg.Rules {
		ruleGroup.Rules = append(ruleGroup.Rules, r.Prometheus())
	}
	return ruleGroup
}

type Rule interface {
	Prometheus() rulefmt.RuleNode
}

type RecordingRule struct {
	Record string            `json:"record" yaml:"record"`
	Expr   string            `json:"expr" yaml:"expr"`
	Labels map[string]string `json:"labels" yaml:"labels"`
}

var _ Rule = &RecordingRule{}

func (r *RecordingRule) Prometheus() rulefmt.RuleNode {
	return rulefmt.RuleNode{
		Record: yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: r.Record,
		},
		Expr: yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: r.Expr,
		},
		Labels: r.Labels,
	}
}

type AlertingRule struct {
	Alert       string            `json:"alert" yaml:"alert"`
	Expr        string            `json:"expr" yaml:"expr"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

func (r *AlertingRule) Prometheus() rulefmt.RuleNode {
	return rulefmt.RuleNode{
		Alert: yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: r.Alert,
		},
		Expr: yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: r.Expr,
		},
		Labels:      r.Labels,
		Annotations: r.Annotations,
	}
}

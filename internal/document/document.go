package document

import (
	"github.com/ajalab/slogen/internal/prometheus/rule"
	"github.com/ajalab/slogen/internal/spec"
)

func ToDocument(spec *spec.Spec) *Document {
	var slos []SLO
	for _, s := range spec.SLOs() {
		slos = append(slos, toSLO(s))
	}

	return &Document{
		Name:        spec.Name(),
		Labels:      spec.Labels(),
		Annotations: spec.Annotations(),
		SLOs:        slos,
	}
}

func toSLO(slo *spec.SLO) SLO {
	return SLO{
		Name:        slo.Name(),
		Labels:      slo.Labels(),
		Annotations: slo.Annotations(),
		Objective:   toObjective(slo.Objective()),
		Indicator:   toIndicator(slo.Indicator(), slo.Objective().Window()),
	}
}

func toObjective(objective *spec.Objective) Objective {
	return Objective{
		Ratio:  objective.Ratio(),
		Window: toWindow(objective.Window()),
	}
}

func toIndicator(indicator spec.Indicator, window spec.Window) Indicator {
	var source string
	var query string
	switch i := indicator.(type) {
	case *spec.PrometheusIndicator:
		source = "prometheus"
		query = rule.GenerateErrorRateQuery(i, window)
	default:
		panic("not implemented")
	}

	return Indicator{
		Source: source,
		Query:  query,
	}
}

func toWindow(window spec.Window) Window {
	var ty string
	switch window.(type) {
	case *spec.RollingWindow:
		ty = "rolling"
	case *spec.CalendarWindow:
		ty = "calendar"
	default:
		panic("unknown window type")
	}

	return Window{
		Name:     window.Name(),
		Type:     ty,
		Duration: window.Duration().String(),
	}
}

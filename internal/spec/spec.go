package spec

import (
	"fmt"
	"time"

	configspec "github.com/ajalab/slogen/internal/config/spec"
	"github.com/prometheus/common/model"
)

type specContext struct {
	windows       []Window
	windowsByName map[string]Window
	alerts        []Alert
	alertNames    map[string]struct{}
}

func (sc *specContext) addWindow(window Window) error {
	if _, ok := sc.windowsByName[window.Name()]; ok {
		return fmt.Errorf("windows have the same name \"%s\"", window.Name())
	}

	sc.windowsByName[window.Name()] = window
	sc.windows = append(sc.windows, window)
	return nil
}

func (sc *specContext) addAlert(alert Alert) error {
	if _, ok := sc.alertNames[alert.Name()]; ok {
		return fmt.Errorf("alerts have the same name \"%s\"", alert.Name())
	}

	sc.alertNames[alert.Name()] = struct{}{}
	sc.alerts = append(sc.alerts, alert)
	return nil
}

func ToSpec(c *configspec.SpecConfig) (*Spec, error) {
	var slos []*SLO
	for _, s := range c.SLOs {
		slo, err := toSLO(&s)
		if err != nil {
			return nil, fmt.Errorf("failed to convert an SLO config \"%s\" into spec: %w", s.Name, err)
		}

		slos = append(slos, slo)
	}

	return &Spec{
		name:        c.Name,
		labels:      ensureMapNotNil(c.Labels),
		annotations: ensureMapNotNil(c.Annotations),
		slos:        slos,
	}, nil
}

func toSLO(slo *configspec.SLOConfig) (*SLO, error) {
	sc := specContext{
		windowsByName: make(map[string]Window),
		alertNames:    make(map[string]struct{}),
	}

	for _, w := range slo.Windows {
		window, err := toWindow(&w)
		if err != nil {
			return nil, fmt.Errorf("failed to convert a window config \"%s\" into spec: %w", w.Name, err)
		}

		if err := sc.addWindow(window); err != nil {
			return nil, fmt.Errorf("failed to add a window \"%s\": %w", w.Name, err)
		}
	}

	objective, err := toObjective(&sc, &slo.Objective)
	if err != nil {
		return nil, fmt.Errorf("failed to convert an objective config s to spec: %w", err)
	}

	indicator, err := toIndicator(&slo.Indicator)
	if err != nil {
		return nil, fmt.Errorf("failed to convert an indicator config s to spec: %w", err)
	}

	for _, a := range slo.Alerts {
		alert, err := toAlert(&sc, &a)
		if err != nil {
			return nil, fmt.Errorf("failed to convert an alert config \"%s\" to spec: %w", a.Name, err)
		}

		if err := sc.addAlert(alert); err != nil {
			return nil, err
		}
	}

	return &SLO{
		name:        slo.Name,
		labels:      ensureMapNotNil(slo.Labels),
		annotations: ensureMapNotNil(slo.Annotations),
		objective:   objective,
		indicator:   indicator,
		windows:     sc.windows,
		alerts:      sc.alerts,
	}, err
}

func toObjective(
	sc *specContext,
	objective *configspec.ObjectiveConfig,
) (*Objective, error) {
	var window Window
	if objective.WindowRef != "" {
		var ok bool
		window, ok = sc.windowsByName[objective.WindowRef]
		if !ok {
			return nil, fmt.Errorf("could not find a window from windowRef \"%s\"", objective.WindowRef)
		}
	}

	return &Objective{
		ratio:  objective.Ratio,
		window: window,
	}, nil
}

func toIndicator(indicator *configspec.IndicatorConfig) (Indicator, error) {
	if indicator.Prometheus != nil {
		return &PrometheusIndicator{
			errorRatio: indicator.Prometheus.ErrorRatio,
			level:      indicator.Prometheus.Level,
		}, nil
	}

	return nil, fmt.Errorf("either one of indicator types must be implemented")
}

func toWindow(window *configspec.WindowConfig) (Window, error) {
	prometheus, err := toPrometheusWindow(window.Prometheus)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a prometheus window config: %w", err)
	}

	if window.Rolling != nil && window.Calendar == nil {
		duration, err := model.ParseDuration(window.Rolling.Duration)
		if err != nil {
			return nil, fmt.Errorf("failed to parse a duration \"%s\": %w", window.Rolling.Duration, err)
		}

		return &RollingWindow{
			name:       window.Name,
			duration:   Duration(duration),
			prometheus: prometheus,
		}, nil
	}

	if window.Calendar != nil && window.Rolling == nil {
		duration, err := model.ParseDuration(window.Calendar.Duration)
		if err != nil {
			return nil, fmt.Errorf("failed to parse a duration \"%s\": %w", window.Calendar.Duration, err)
		}

		start, err := time.Parse(time.DateTime, window.Calendar.Start)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time in start: %w", err)
		}

		return &CalendarWindow{
			name:       window.Name,
			duration:   Duration(duration),
			start:      start,
			prometheus: prometheus,
		}, nil
	}

	return nil, fmt.Errorf("either one of windows must be implemented in window %#v", window)
}

func toPrometheusWindow(pw *configspec.PrometheusWindowConfig) (*PrometheusWindow, error) {
	if pw == nil {
		return &PrometheusWindow{
			evaluationInterval: Duration(0),
		}, nil
	}

	evaluationInterval, err := model.ParseDuration(pw.EvaluationInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse an evaluation interval \"%s\": %w", pw.EvaluationInterval, err)
	}

	return &PrometheusWindow{
		evaluationInterval: Duration(evaluationInterval),
	}, nil
}

func toAlert(sc *specContext, alert *configspec.AlertConfig) (Alert, error) {
	alerter, err := toAlerter(&alert.Alerter)
	if err != nil {
		return nil, fmt.Errorf("failed to convert alerter config to spec: %w", err)
	}

	if alert.BurnRate != nil && alert.Breach == nil {
		window, err := toBurnRateAlertWindow(sc, alert.BurnRate)
		if err != nil {
			return nil, fmt.Errorf("failed to convert burn rate alert window config to spec: %w", err)
		}
		return &BurnRateAlert{
			name:                alert.Name,
			consumedBudgetRatio: alert.BurnRate.ConsumedBudgetRatio,
			window:              window,
			alerter:             alerter,
		}, nil
	}

	if alert.Breach != nil && alert.BurnRate == nil {
		windowRef := alert.Breach.WindowRef
		window, ok := sc.windowsByName[alert.Breach.WindowRef]
		if !ok {
			return nil, fmt.Errorf("could not find a window from windowRef \"%s\"", windowRef)
		}
		return &BreachAlert{
			name:    alert.Name,
			window:  window,
			alerter: alerter,
		}, nil
	}
	return nil, fmt.Errorf("either one of alert types must be implemented")
}

func toAlerter(alerter *configspec.AlerterConfig) (Alerter, error) {
	if alerter.Prometheus != nil {
		return &PrometheusAlerter{
			name:        alerter.Prometheus.Name,
			labels:      alerter.Prometheus.Labels,
			annotations: alerter.Prometheus.Annotations,
		}, nil
	}
	return nil, fmt.Errorf("either one of alerter types must be implemented")
}

func toBurnRateAlertWindow(sc *specContext, a *configspec.BurnRateAlertConfig) (BurnRateAlertWindow, error) {
	if a.SingleWindow != nil && a.MultiWindows == nil {
		windowRef := a.SingleWindow.WindowRef
		window, ok := sc.windowsByName[windowRef]
		if !ok {
			return nil, fmt.Errorf("could not find a window from windowRef \"%s\"", windowRef)
		}

		return &BurnRateAlertSingleWindow{
			window: window,
		}, nil
	} else if a.MultiWindows != nil && a.SingleWindow == nil {
		shortWindowRef := a.MultiWindows.ShortWindowRef
		shortWindow, ok := sc.windowsByName[shortWindowRef]
		if !ok {
			return nil, fmt.Errorf("could not find a window from shortWindowRef \"%s\"", shortWindowRef)
		}
		longWindowRef := a.MultiWindows.LongWindowRef
		longWindow, ok := sc.windowsByName[longWindowRef]
		if !ok {
			return nil, fmt.Errorf("could not find a window from longWindowRef \"%s\"", longWindowRef)
		}

		return &BurnRateAlertMultiWindows{
			shortWindow: shortWindow,
			longWindow:  longWindow,
		}, nil
	}
	return nil, fmt.Errorf("either one of burn rate alert windows must be implemented")
}

func ensureMapNotNil(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

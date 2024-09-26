package spec

import (
	"time"

	"github.com/prometheus/common/model"
)

type Duration = model.Duration

type Spec struct {
	name        string
	labels      map[string]string
	annotations map[string]string
	slos        []*SLO
}

func (s *Spec) Name() string {
	return s.name
}

func (s *Spec) Labels() map[string]string {
	return s.labels
}

func (s *Spec) Annotations() map[string]string {
	return s.annotations
}

func (s *Spec) SLOs() []*SLO {
	return s.slos
}

type SLO struct {
	name        string
	labels      map[string]string
	annotations map[string]string
	objective   *Objective
	indicator   Indicator
	alerts      []Alert
	windows     []Window
}

func (s *SLO) Name() string {
	return s.name
}

func (s *SLO) Labels() map[string]string {
	return s.labels
}

func (s *SLO) Annotations() map[string]string {
	return s.annotations
}

func (s *SLO) Objective() *Objective {
	return s.objective
}

func (s *SLO) Indicator() Indicator {
	return s.indicator
}

func (s *SLO) Alerts() []Alert {
	return s.alerts
}

func (s *SLO) Windows() []Window {
	return s.windows
}

type Objective struct {
	ratio  float64
	window Window
}

func (o *Objective) Ratio() float64 {
	return o.ratio
}

func (o *Objective) Window() Window {
	return o.window
}

type Indicator interface {
}

type PrometheusIndicator struct {
	errorRatio string
	level      []string
}

func (pi *PrometheusIndicator) ErrorRatio() string {
	return pi.errorRatio
}

func (pi *PrometheusIndicator) Level() []string {
	return pi.level
}

type PrometheusWindow struct {
	evaluationInterval Duration
}

func (pw *PrometheusWindow) EvaluationInterval() Duration {
	return pw.evaluationInterval
}

type Window interface {
	Name() string
	Duration() Duration
	Prometheus() *PrometheusWindow
}

type RollingWindow struct {
	name       string
	duration   Duration
	prometheus *PrometheusWindow
}

var _ Window = &RollingWindow{}

func (w *RollingWindow) Name() string {
	return w.name
}

func (w *RollingWindow) Duration() Duration {
	return w.duration
}

func (w *RollingWindow) Prometheus() *PrometheusWindow {
	return w.prometheus
}

type CalendarWindow struct {
	name       string
	duration   Duration
	start      time.Time
	prometheus *PrometheusWindow
}

var _ Window = &CalendarWindow{}

func (w *CalendarWindow) Name() string {
	return w.name
}

func (w *CalendarWindow) Duration() Duration {
	return w.duration
}

func (w *CalendarWindow) Start() time.Time {
	return w.start
}

func (w *CalendarWindow) Prometheus() *PrometheusWindow {
	return w.prometheus
}

type Alert interface {
	Name() string
	Alerter() Alerter
}

type BurnRateAlert struct {
	name                string
	consumedBudgetRatio float64
	window              BurnRateAlertWindow
	alerter             Alerter
}

func (a *BurnRateAlert) Name() string {
	return a.name
}

func (a *BurnRateAlert) ConsumedBudgetRatio() float64 {
	return a.consumedBudgetRatio
}

func (a *BurnRateAlert) Window() BurnRateAlertWindow {
	return a.window
}

func (a *BurnRateAlert) Alerter() Alerter {
	return a.alerter
}

type BurnRateAlertWindow interface {
	Window() Window
}

type BurnRateAlertMultiWindows struct {
	shortWindow Window
	longWindow  Window
}

func (w *BurnRateAlertMultiWindows) Window() Window {
	return w.longWindow
}

func (w *BurnRateAlertMultiWindows) ShortWindow() Window {
	return w.shortWindow
}

func (w *BurnRateAlertMultiWindows) LongWindow() Window {
	return w.longWindow
}

type BreachAlert struct {
	name    string
	window  Window
	alerter Alerter
}

func (a *BreachAlert) Name() string {
	return a.name
}

func (a *BreachAlert) Window() Window {
	return a.window
}

func (a *BreachAlert) Alerter() Alerter {
	return a.alerter
}

type Alerter interface {
}

type PrometheusAlerter struct {
	name        string
	labels      map[string]string
	annotations map[string]string
}

func (a *PrometheusAlerter) Name() string {
	return a.name
}

func (a *PrometheusAlerter) Labels() map[string]string {
	return a.labels
}

func (a *PrometheusAlerter) Annotations() map[string]string {
	return a.annotations
}

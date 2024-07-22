package prometheus

import (
	"strings"

	"github.com/bcds/go-hpc-dagbft/protocol/observer"
	prom "github.com/prometheus/client_golang/prometheus"
	promHttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

// the separator of name prefix, should be unified with Prometheus server.
const (
	separator          = "_"
	namespaceLabelName = "ns"
)

// DefaultHandler is the default prometheus handler used to handle
// prometheus http requests.
var DefaultHandler = promHttp.Handler()

var _ observer.MetricProvider = (*Provider)(nil)

// Provider is the Prometheus provider instance.
type Provider struct {
	// Name is the optionally prefix of all metrics registered under
	// this Provider.
	// If it is specified, users need only specify `Name` field when
	// register metrics and the final metric name will be the
	// concatenation of Provider name and metric name.
	// If it is not specified (an empty string), the final metric name
	// will be the concatenation of metric Namespace, Subsystem and
	// metric Name.
	Name string

	// Namespace is the optionally namespace of all metrics registered under
	// this Provider.
	// If it is specified, users need not specify `Namespace` field when
	// register metrics and the final metric namespace will be the
	// displayed as a label which key is "ns".
	Namespace string
}

// SubProvider creates a new instance of Provider with given suffix.
// suffixes will be concatenated by separator `_`.
func (p *Provider) SubProvider(suffix string) observer.MetricProvider {
	var subName string
	if p.Name == "" {
		subName = suffix
	} else {
		subName = strings.Join([]string{p.Name, suffix}, separator)
	}

	return &Provider{
		Name:      subName,
		Namespace: p.Namespace,
	}
}

// SetNamespace set given Namespace to a provider.
func (p *Provider) SetNamespace(namespace string) {
	p.Namespace = namespace
}

// NewCounter constructs and registers a Prometheus CounterVec.
func (p *Provider) NewCounter(o observer.CounterOpts) (observer.Counter, error) {
	// convert user options into prometheus CounterOpts
	promCounterOpts := prom.CounterOpts{
		Namespace: o.Namespace,
		Subsystem: o.Subsystem,
		Name:      o.Name,
		Help:      o.Help,
	}

	// use pre-defined prefix if has.
	if p.Name != "" {
		promCounterOpts.Namespace = ""
		promCounterOpts.Subsystem = ""
		promCounterOpts.Name = strings.Join([]string{p.Name, o.Name}, separator)
	}

	// use pre-defined namespace as const label if has.
	if p.Namespace != "" {
		promCounterOpts.ConstLabels = prom.Labels{namespaceLabelName: p.Namespace}
	}

	counterVec := prom.NewCounterVec(promCounterOpts, o.LabelNames)
	// register this counterVec into DefaultRegisterer
	err := prom.Register(counterVec)
	if err != nil {
		return nil, err
	}

	return &Counter{
		promCounter: counterVec,
	}, nil
}

// NewGauge constructs and registers a Prometheus GaugeVec.
func (p *Provider) NewGauge(o observer.GaugeOpts) (observer.Gauge, error) {
	// convert user options into prometheus GaugeOpts
	promGaugeOpts := prom.GaugeOpts{
		Namespace: o.Namespace,
		Subsystem: o.Subsystem,
		Name:      o.Name,
		Help:      o.Help,
	}

	// use pre-defined prefix if has.
	if p.Name != "" {
		promGaugeOpts.Namespace = ""
		promGaugeOpts.Subsystem = ""
		promGaugeOpts.Name = strings.Join([]string{p.Name, o.Name}, separator)
	}

	// use pre-defined namespace as const label if has.
	if p.Namespace != "" {
		promGaugeOpts.ConstLabels = prom.Labels{namespaceLabelName: p.Namespace}
	}

	gaugeVec := prom.NewGaugeVec(promGaugeOpts, o.LabelNames)
	// register this gaugeVec into DefaultRegisterer
	err := prom.Register(gaugeVec)
	if err != nil {
		return nil, err
	}

	return &Gauge{
		promGauge: gaugeVec,
	}, nil
}

// NewHistogram constructs and registers a Prometheus HistogramVec.
func (p *Provider) NewHistogram(o observer.HistogramOpts) (observer.Histogram, error) {
	// convert user options into prometheus HistogramOpts
	promHistogramOpts := prom.HistogramOpts{
		Namespace: o.Namespace,
		Subsystem: o.Subsystem,
		Name:      o.Name,
		Help:      o.Help,
		Buckets:   o.Buckets,
	}

	// use pre-defined prefix if has.
	if p.Name != "" {
		promHistogramOpts.Namespace = ""
		promHistogramOpts.Subsystem = ""
		promHistogramOpts.Name = strings.Join([]string{p.Name, o.Name}, separator)
	}

	// use pre-defined namespace as const label if has.
	if p.Namespace != "" {
		promHistogramOpts.ConstLabels = prom.Labels{namespaceLabelName: p.Namespace}
	}

	histogramVec := prom.NewHistogramVec(promHistogramOpts, o.LabelNames)
	// register this histogramVec into DefaultRegisterer
	err := prom.Register(histogramVec)
	if err != nil {
		return nil, err
	}

	return &Histogram{
		promHistogram: histogramVec,
	}, nil
}

// NewSummary constructs and registers a Prometheus SummaryVec.
func (p *Provider) NewSummary(o observer.SummaryOpts) (observer.Summary, error) {
	// convert user options into prometheus SummaryOpts
	promSummaryOpts := prom.SummaryOpts{
		Namespace:  o.Namespace,
		Subsystem:  o.Subsystem,
		Name:       o.Name,
		Help:       o.Help,
		Objectives: o.Objectives,
		MaxAge:     o.MaxAge,
		AgeBuckets: o.AgeBuckets,
		BufCap:     o.BufCap,
	}

	// use pre-defined prefix if has.
	if p.Name != "" {
		promSummaryOpts.Namespace = ""
		promSummaryOpts.Subsystem = ""
		promSummaryOpts.Name = strings.Join([]string{p.Name, o.Name}, separator)
	}

	// use pre-defined namespace as const label if has.
	if p.Namespace != "" {
		promSummaryOpts.ConstLabels = prom.Labels{namespaceLabelName: p.Namespace}
	}

	summaryVec := prom.NewSummaryVec(promSummaryOpts, o.LabelNames)
	// register this summaryVec into DefaultRegisterer
	err := prom.Register(summaryVec)
	if err != nil {
		return nil, err
	}

	return &Summary{
		promSummary: summaryVec,
	}, nil
}

// Counter is the wrapper of Prometheus CounterVec.
type Counter struct {
	promCounter *prom.CounterVec
	labelValues prom.Labels
}

// With implements the Counter interface.
func (c *Counter) With(labelValues ...string) observer.Counter {
	newLVS := mergeLabelValues(c.labelValues, labelValues...)
	return &Counter{
		promCounter: c.promCounter,
		labelValues: newLVS,
	}
}

// Add implements the Counter interface.
func (c *Counter) Add(delta float64) {
	c.promCounter.With(c.labelValues).Add(delta)
}

// Unregister implements the Counter interface.
func (c *Counter) Unregister() {
	prom.Unregister(c.promCounter)
}

// Gauge is the wrapper of Prometheus GaugeVec.
type Gauge struct {
	promGauge   *prom.GaugeVec
	labelValues prom.Labels
}

// With implements the Gauge interface.
func (g *Gauge) With(labelValues ...string) observer.Gauge {
	newLVS := mergeLabelValues(g.labelValues, labelValues...)
	return &Gauge{
		promGauge:   g.promGauge,
		labelValues: newLVS,
	}
}

// Add implements the Gauge interface.
func (g *Gauge) Add(delta float64) {
	g.promGauge.With(g.labelValues).Add(delta)
}

// Set implements the Gauge interface.
func (g *Gauge) Set(value float64) {
	g.promGauge.With(g.labelValues).Set(value)
}

// Unregister implements the Gauge interface.
func (g *Gauge) Unregister() {
	prom.Unregister(g.promGauge)
}

// Histogram is the wrapper of Prometheus HistogramVec.
type Histogram struct {
	promHistogram *prom.HistogramVec
	labelValues   prom.Labels
}

// With implements the Histogram interface.
func (h *Histogram) With(labelValues ...string) observer.Histogram {
	newLVS := mergeLabelValues(h.labelValues, labelValues...)
	return &Histogram{
		promHistogram: h.promHistogram,
		labelValues:   newLVS,
	}
}

// Observe implements the Histogram interface.
func (h *Histogram) Observe(value float64) {
	h.promHistogram.With(h.labelValues).Observe(value)
}

// Unregister implements the Histogram interface.
func (h *Histogram) Unregister() {
	prom.Unregister(h.promHistogram)
}

// Summary is the wrapper of Prometheus SummaryVec.
type Summary struct {
	promSummary *prom.SummaryVec
	labelValues prom.Labels
}

// With implements the Summary interface.
func (s *Summary) With(labelValues ...string) observer.Summary {
	newLVS := mergeLabelValues(s.labelValues, labelValues...)
	return &Summary{
		promSummary: s.promSummary,
		labelValues: newLVS,
	}
}

// Observe implements the Summary interface.
func (s *Summary) Observe(value float64) {
	s.promSummary.With(s.labelValues).Observe(value)
}

// Unregister implements the Summary interface.
func (s *Summary) Unregister() {
	prom.Unregister(s.promSummary)
}

// mergeLabelValues merges an origin prom.Labels with some new labelValues.
func mergeLabelValues(originLVS prom.Labels, labelValues ...string) prom.Labels {
	newLVS := prom.Labels{}

	// merge origin labelValues
	for label, value := range originLVS {
		newLVS[label] = value
	}

	// fill new labelValues to even count as labelValues should be
	// paired with <label, value>
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, "unknown")
	}

	// merge new labelValues
	for i := 0; i < len(labelValues); i += 2 {
		newLVS[labelValues[i]] = labelValues[i+1]
	}

	return newLVS
}

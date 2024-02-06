package wsprometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var DefaultManager = NewManager()

type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
	Summary
)

type Metric struct {
	Type        MetricType
	Name        string
	Description string
	Labels      []string
	Buckets     []float64
	Objectives  map[float64]float64
	collector   prometheus.Collector
}

//func (m *Metric) Inc(labelValues []string) (err error) {
//	switch m.Type {
//	case Counter:
//		m.collector.(*prometheus.CounterVec).WithLabelValues(labelValues...).Inc()
//		return
//	case Gauge:
//		m.collector.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Inc()
//		return
//	default:
//		return fmt.Errorf("metric '%s' not Gauge or Counter type", m.Name)
//	}
//}

func (m *Metric) Add(labelValues []string, value float64) (err error) {
	switch m.Type {
	case Counter:
		m.collector.(*prometheus.CounterVec).WithLabelValues(labelValues...).Add(value)
		return
	case Gauge:
		m.collector.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Add(value)
		return
	default:
		return fmt.Errorf("metric '%s' not Gauge or Counter type", m.Name)
	}
}

func (m *Metric) Observe(labelValues []string, value float64) (err error) {
	switch m.Type {
	case Histogram:
		m.collector.(*prometheus.HistogramVec).WithLabelValues(labelValues...).Observe(value)
		return
	case Summary:
		m.collector.(*prometheus.SummaryVec).WithLabelValues(labelValues...).Observe(value)
		return
	default:
		return fmt.Errorf("metric '%s' not Histogram or Summary type", m.Name)
	}
}

type Manager struct {
	metrics map[string]*Metric
}

func NewManager() *Manager {
	return &Manager{
		metrics: make(map[string]*Metric),
	}
}

func (m *Manager) Add(metric *Metric) error {
	if _, ok := m.metrics[metric.Name]; ok {
		return fmt.Errorf("metric '%s' already existed", metric.Name)
	}
	switch metric.Type {
	case Counter:
		metric.collector = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: metric.Name, Help: metric.Description},
			metric.Labels)
	case Gauge:
		metric.collector = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: metric.Name, Help: metric.Description},
			metric.Labels,
		)
	case Histogram:
		if metric.Buckets == nil {
			metric.Buckets = prometheus.DefBuckets
		}
		metric.collector = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: metric.Name, Help: metric.Description, Buckets: metric.Buckets},
			metric.Labels,
		)
	case Summary:
		metric.collector = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{Name: metric.Name, Help: metric.Description, Objectives: metric.Objectives},
			metric.Labels,
		)
	}
	prometheus.MustRegister(metric.collector)
	m.metrics[metric.Name] = metric
	return nil
}

func (m *Manager) Get(name string) *Metric {
	return m.metrics[name]
}

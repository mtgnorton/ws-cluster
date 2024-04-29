package wsprometheus

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	MetricRequestTotal    = "request_total"
	MetricRequestURLTotal = "request_url_total"
	MetricRequestDuration = "request_duration"
	MetricWsConnection    = "ws_connection"

	MetricQueueHandleDuration = "queue_handle_duration"
	MetricQueueHandleTotal    = "queue_handle_total"
)

var DefaultPrometheus = New()

var once sync.Once

type Prometheus struct {
	opts Options
}

func New(opts ...Option) *Prometheus {
	return &Prometheus{opts: NewOptions(opts...)}
}
func (p *Prometheus) Init() {
	if !p.isEnable() {
		return
	}
	once.Do(func() {
		p.init()
	})
}

func (p *Prometheus) GetAdd(metric string, labelValues []string, value float64) (err error) {
	if !p.isEnable() {
		return
	}
	return p.opts.MetricManager.Get(metric).Add(labelValues, value)
}

func (p *Prometheus) GetObserve(metric string, labelValues []string, value float64) (err error) {
	if !p.isEnable() {
		return
	}
	return p.opts.MetricManager.Get(metric).Observe(labelValues, value)
}

func (p *Prometheus) Options() Options {
	return p.opts
}

func (p *Prometheus) init() {

	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MetricRequestTotal,
		Description: "all the server received request num.",
		Labels:      nil,
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MetricRequestURLTotal,
		Description: "all the server received request url num.",
		Labels:      []string{"url", "code"},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricRequestDuration,
		Description: "all the server received request duration.",
		Labels:      []string{"url"},
		Buckets:     []float64{0.1, 0.3, 0.5, 1, 2, 3, 5, 10},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Gauge,
		Name:        MetricWsConnection,
		Description: "current ws connection num.",
	})

	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricQueueHandleDuration,
		Description: "queue handle msg duration.",
		Labels:      []string{"queue_type"},
		Buckets:     []float64{10, 30, 60, 100, 200, 500, 1000},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MetricQueueHandleTotal,
		Description: "queue handle msg  num",
		Labels:      []string{"queue_type"},
	})

	http.Handle(p.opts.Config.Values().Prometheus.Path, promhttp.Handler())
	go func() {
		err := http.ListenAndServe(p.opts.Config.Values().Prometheus.Addr, nil)
		if err != nil {
			panic(err)
		}
	}()
}

func (p *Prometheus) isEnable() bool {
	return p.opts.Config.Values().Prometheus.Enable
}

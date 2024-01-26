package wsprometheus

import (
	"net/http"
	"sync"

	"github.com/mtgnorton/ws-cluster/tools/wsprometheus/metric"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	MetricRequestTotal    = "request_total"
	MetricRequestURLTotal = "request_url_total"
	MetricRequestDuration = "request_duration"
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
	once.Do(func() {
		p.init()
	})
}

func (p *Prometheus) init() {

	if !p.opts.Config.Values().Prometheus.Enable {
		return
	}

	_ = p.opts.MetricManager.Add(&metric.Metric{
		Type:        metric.Counter,
		Name:        MetricRequestTotal,
		Description: "all the server received request num.",
		Labels:      nil,
	})
	_ = p.opts.MetricManager.Add(&metric.Metric{
		Type:        metric.Counter,
		Name:        MetricRequestURLTotal,
		Description: "all the server received request url num.",
		Labels:      []string{"url", "code"},
	})
	_ = p.opts.MetricManager.Add(&metric.Metric{
		Type:        metric.Histogram,
		Name:        MetricRequestDuration,
		Description: "all the server received request duration.",
		Labels:      []string{"url"},
		Buckets:     []float64{0.1, 0.3, 0.5, 1, 2, 3, 5, 10},
	})

	http.Handle(p.opts.Config.Values().Prometheus.Path, promhttp.Handler())
	go func() {
		err := http.ListenAndServe(p.opts.Config.Values().Prometheus.Addr, nil)
		if err != nil {
			panic(err)
		}
	}()
}

func (p *Prometheus) Options() Options {
	return p.opts
}

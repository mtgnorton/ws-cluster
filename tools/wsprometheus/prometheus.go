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

	MerticQueueEnter = "queue_enter" // 统计进入队列的消息数量
	MetricQueueOut   = "queue_out"   // 统计出队列的消息数量
	MetricQueueDrop  = "queue_drop"  // 统计丢弃的消息数量

	MertricQueueEnterDuration = "queue_enter_duration" // 统计进入队列的消息时间

	MetricQueueHandleDuration      = "queue_handle_duration"       // 统计队列处理消息的时间
	MetricQueuePublishWaitDuration = "queue_publish_wait_duration" // 统计写入本地发布缓冲区等待时间
	MetricQueueLagDuration         = "queue_lag_duration"          // 统计消息进入redis后到被消费的等待时间
	MetricQueueDispatchDuration    = "queue_dispatch_duration"     // 统计单条消息分发处理时间

	MetricClientSendDrop              = "client_send_drop"                // 统计客户端发送队列丢弃次数
	MetricClientSendQueueWaitDuration = "client_send_queue_wait_duration" // 统计客户端发送队列等待时间
	MetricClientWriteDuration         = "client_write_duration"           // 统计websocket写入耗时
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

func (p *Prometheus) Get(metric string) *Metric {
	if !p.isEnable() {
		return nil
	}
	return p.opts.MetricManager.Get(metric)
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
		Labels:      []string{"node", "ip"},
		Description: "current ws connection num.",
	})

	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MetricQueueOut,
		Description: "queue handle msg  num",
		Labels:      []string{"node", "ip"},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MerticQueueEnter,
		Description: "queue enter msg  num",
		Labels:      []string{"node", "ip"},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MetricQueueDrop,
		Description: "queue drop msg num.",
		Labels:      []string{"node", "ip"},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MertricQueueEnterDuration,
		Description: "queue enter msg duration.",
		Labels:      []string{"node", "ip"},
		Buckets:     []float64{10, 30, 60, 100, 200, 500, 1000},
	})

	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricQueueHandleDuration,
		Description: "queue handle msg duration.",
		Labels:      []string{"node", "ip"},
		Buckets:     []float64{10, 30, 60, 100, 200, 500, 1000},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricQueuePublishWaitDuration,
		Description: "queue publish wait duration.",
		Labels:      []string{"node", "ip"},
		Buckets:     []float64{1, 5, 10, 20, 50, 100, 500, 1000, 5000},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricQueueLagDuration,
		Description: "queue lag duration between redis append and consume.",
		Labels:      []string{"node", "ip", "type"},
		Buckets:     []float64{1, 5, 10, 20, 50, 100, 200, 500, 1000, 3000, 5000, 10000},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricQueueDispatchDuration,
		Description: "queue single message dispatch duration.",
		Labels:      []string{"node", "ip", "type"},
		Buckets:     []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Counter,
		Name:        MetricClientSendDrop,
		Description: "client send queue drop count.",
		Labels:      []string{"node", "ip", "client_type"},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricClientSendQueueWaitDuration,
		Description: "client send queue wait duration.",
		Labels:      []string{"node", "ip", "client_type"},
		Buckets:     []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 3000, 5000},
	})
	_ = p.opts.MetricManager.Add(&Metric{
		Type:        Histogram,
		Name:        MetricClientWriteDuration,
		Description: "client websocket write duration.",
		Labels:      []string{"node", "ip", "client_type"},
		Buckets:     []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000},
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

package server

import (
	"context"

	"github.com/mtgnorton/ws-cluster/core/queue"
	"github.com/mtgnorton/ws-cluster/logger"

	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"
)

type Option func(*Options)

type Options struct {
	ctx        context.Context
	config     config.Config
	logger     logger.Logger
	prometheus *wsprometheus.Prometheus
	queue      queue.Queue
	port       int
}

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:        context.Background(),
		config:     config.DefaultConfig,
		logger:     logger.DefaultLogger,
		prometheus: wsprometheus.DefaultPrometheus,
		queue:      queue.GetQueueInstance(config.DefaultConfig),
		port:       config.DefaultConfig.Values().HttpServer.Port,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

func WithConfig(c config.Config) Option {
	return func(o *Options) {
		o.config = c
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}
func WithPrometheus(p *wsprometheus.Prometheus) Option {
	return func(o *Options) {
		o.prometheus = p
	}
}

func WithQueue(q queue.Queue) Option {
	return func(o *Options) {
		o.queue = q
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

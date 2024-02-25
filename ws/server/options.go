package server

import (
	"context"

	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/core/manager"
	"github.com/mtgnorton/ws-cluster/logger"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"
	"github.com/mtgnorton/ws-cluster/ws/handler"
)

type Option func(*Options)
type Options struct {
	ctx        context.Context
	config     config.Config
	manager    manager.Manager
	handler    handler.Handle
	logger     logger.Logger
	prometheus *wsprometheus.Prometheus
	port       int
}

func NewOptions(opts ...Option) Options {

	options := Options{
		ctx:        context.TODO(),
		config:     config.DefaultConfig,
		manager:    manager.DefaultManager,
		handler:    handler.DefaultHandler,
		logger:     logger.DefaultLogger,
		prometheus: wsprometheus.DefaultPrometheus,
	}
	for _, o := range opts {
		o(&options)
	}

	c := options.config
	options.port = c.Values().WsServer.Port
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

func WithHandler(h handler.Handle) Option {
	return func(o *Options) {
		o.handler = h
	}
}

func WithManager(m manager.Manager) Option {
	return func(o *Options) {
		o.manager = m
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

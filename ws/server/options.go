package server

import (
	"context"

	"ws-cluster/config"
	"ws-cluster/core/checking"
	"ws-cluster/core/manager"
	"ws-cluster/logger"
	"ws-cluster/tools/wsprometheus"
	"ws-cluster/ws/handler"
)

type Option func(*Options)
type Options struct {
	ctx        context.Context
	config     config.Config
	manager    manager.Manager
	handler    handler.Handle
	logger     logger.Logger
	prometheus *wsprometheus.Prometheus
	checking   *checking.Checking
	port       int
}

func NewOptions(opts ...Option) Options {

	options := Options{
		ctx:        context.Background(),
		config:     config.DefaultConfig,
		manager:    manager.DefaultManager,
		handler:    handler.DefaultHandler,
		logger:     logger.DefaultLogger,
		prometheus: wsprometheus.DefaultPrometheus,
		checking:   checking.DefaultChecking,
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

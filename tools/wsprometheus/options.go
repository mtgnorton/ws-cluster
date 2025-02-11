package wsprometheus

import (
	"context"

	"ws-cluster/logger"

	"ws-cluster/config"
)

type Option func(*Options)

type Options struct {
	// Other options for implementations of the interface
	// can be stored in a Context
	Context       context.Context
	Config        config.Config
	Logger        logger.Logger
	MetricManager *Manager
}

func NewOptions(opts ...Option) Options {
	options := Options{
		Context:       context.Background(),
		Config:        config.DefaultConfig,
		MetricManager: DefaultManager,
		Logger:        logger.DefaultLogger,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

func WithConfig(c config.Config) Option {
	return func(o *Options) {
		o.Config = c
	}
}

func WithManager(m *Manager) Option {
	return func(o *Options) {
		o.MetricManager = m
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

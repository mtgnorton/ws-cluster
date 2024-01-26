package wsprometheus

import (
	"context"

	"github.com/mtgnorton/ws-cluster/logger"

	"go.uber.org/zap"

	"github.com/mtgnorton/ws-cluster/tools/wsprometheus/metric"

	"github.com/mtgnorton/ws-cluster/config"
)

type Option func(*Options)

type Options struct {
	// Other options for implementations of the interface
	// can be stored in a Context
	Context       context.Context
	Config        config.Config
	Logger        *zap.SugaredLogger
	MetricManager *metric.Manager
}

func NewOptions(opts ...Option) Options {
	options := Options{
		Context:       context.Background(),
		Config:        config.DefaultConfig,
		MetricManager: metric.DefaultManager,
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

func WithManager(m *metric.Manager) Option {
	return func(o *Options) {
		o.MetricManager = m
	}
}

func WithLogger(l *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

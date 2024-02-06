package manager

import (
	"context"

	"github.com/mtgnorton/ws-cluster/logger"

	"github.com/mtgnorton/ws-cluster/config"
)

type Options struct {
	ctx    context.Context
	config config.Config
	logger logger.Logger
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:    context.Background(),
		config: config.DefaultConfig,
		logger: logger.DefaultLogger,
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

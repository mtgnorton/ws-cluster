package manager

import (
	"context"

	"ws-cluster/logger"
)

type Options struct {
	ctx    context.Context
	logger logger.Logger
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:    context.Background(),
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

func WithLogger(l logger.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

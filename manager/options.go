package manager

import (
	"context"
	"github.com/mtgnorton/ws-cluster/config"
	"go.uber.org/zap"
)

type Options struct {
	ctx    context.Context
	config config.Config
	logger *zap.SugaredLogger
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:    context.Background(),
		config: config.DefaultConfig,
		logger: zap.NewNop().Sugar(),
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

func WithLogger(l *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

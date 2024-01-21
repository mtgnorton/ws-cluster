package sentry_instance

import (
	"context"

	"github.com/mtgnorton/ws-cluster/config"
)

type Option func(*Options)

type Options struct {
	ctx   context.Context
	env   config.Env
	debug bool
}

func NewOptions(opts ...Option) Options {

	options := Options{
		ctx: context.TODO(),
		env: config.DefaultConfig.Values().Env,
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

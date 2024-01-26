package wssentry

import (
	"context"

	"github.com/mtgnorton/ws-cluster/config"
)

type Option func(*Options)

type Options struct {
	ctx              context.Context
	env              config.Env
	debug            bool
	dsn              string
	tracesSampleRate float64
}

func NewOptions(opts ...Option) Options {
	c := config.DefaultConfig.Values()
	options := Options{
		ctx:              context.TODO(),
		env:              c.Env,
		debug:            c.Env == config.Local,
		dsn:              c.Sentry.DSN,
		tracesSampleRate: c.Sentry.TracesSampleRate,
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

func WithEnv(env config.Env) Option {
	return func(o *Options) {
		o.env = env
	}
}

func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.debug = debug
	}
}

func WithDSN(dsn string) Option {
	return func(o *Options) {
		o.dsn = dsn
	}
}

func WithTracesSampleRate(tracesSampleRate float64) Option {
	return func(o *Options) {
		o.tracesSampleRate = tracesSampleRate
	}
}

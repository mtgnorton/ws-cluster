package server

import (
	"context"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/shared"
)

type Option func(*Options)

type Options struct {
	ctx    context.Context
	config config.Config
	shared *shared.Shared
	port   int
}

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:    context.Background(),
		config: config.DefaultConfig,
		shared: shared.DefaultShared,
		port:   config.DefaultConfig.Values().HttpServer.Port,
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

func WithShared(s *shared.Shared) Option {
	return func(o *Options) {
		o.shared = s
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

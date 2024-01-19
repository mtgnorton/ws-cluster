package server

import (
	"context"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/core/manager"
	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/ws/handler"
	"github.com/mtgnorton/ws-cluster/ws/message"
)

type Option func(*Options)
type Options struct {
	ctx     context.Context
	config  config.Config
	shared  *shared.Shared
	manager manager.Manager
	parser  message.Parse
	handler handler.Handle
	port    int
}

func NewOptions(opts ...Option) Options {

	options := Options{
		ctx:     context.TODO(),
		config:  config.DefaultConfig,
		shared:  shared.DefaultShared,
		manager: manager.DefaultManager,
		parser:  message.DefaultParser,
		handler: handler.DefaultHandler,
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
func WithShared(s *shared.Shared) Option {
	return func(o *Options) {
		o.shared = s
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

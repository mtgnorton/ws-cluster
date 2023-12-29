package server

import (
	"context"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/manager"
	"github.com/mtgnorton/ws-cluster/message"
	"github.com/mtgnorton/ws-cluster/queue"
	"github.com/mtgnorton/ws-cluster/shared"
)

type Option func(*Options)
type Options struct {
	ctx     context.Context
	manager manager.Manager
	config  config.Config
	shared  *shared.Shared
	queue   queue.Queue
	port    int
}

func newOptions(opts ...Option) Options {

	options := Options{
		ctx:     context.TODO(),
		config:  config.DefaultConfig,
		shared:  shared.DefaultShared,
		manager: manager.DefaultManager,
		queue:   queue.DefaultQueue,
	}
	for _, o := range opts {
		o(&options)
	}
	c := options.config
	options.port = c.Values().Server.Port
	return options
}
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

func WithManager(m manager.Manager) Option {
	return func(o *Options) {
		o.manager = m
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

func WithQueue(q queue.Queue) Option {
	return func(o *Options) {
		o.queue = q
	}
}
func WithParser(p message.WsParser) Option {
	return func(o *Options) {
		o.parser = p
	}
}

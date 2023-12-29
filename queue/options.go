package queue

import (
	"context"
	"github.com/mtgnorton/ws-cluster/dispatcher"
	"github.com/mtgnorton/ws-cluster/message"
	"github.com/mtgnorton/ws-cluster/shared"
)

type Options struct {
	ctx        context.Context
	shared     *shared.Shared
	dispatcher dispatcher.Dispatcher
}

type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		ctx:        context.Background(),
		shared:     shared.DefaultShared,
		dispatcher: dispatcher.DefaultDispatcher,
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
func WithShared(s *shared.Shared) Option {
	return func(o *Options) {
		o.shared = s
	}
}

func WithParser(p message.WsParser) Option {
	return func(o *Options) {
		o.parser = p
	}
}

func WithHandler(h handle.Handler) Option {
	return func(o *Options) {
		o.handler = h
	}
}

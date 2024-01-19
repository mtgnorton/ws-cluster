package queue

import (
	"context"
	"github.com/mtgnorton/ws-cluster/http/handler"
	"github.com/mtgnorton/ws-cluster/shared"
)

type Options struct {
	ctx     context.Context
	shared  *shared.Shared
	handler handler.Handle
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:     context.Background(),
		shared:  shared.DefaultShared,
		handler: handler.DefaultPushHandler,
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

func WithHandler(h handler.Handle) Option {
	return func(o *Options) {
		o.handler = h
	}
}

package option

import (
	"context"

	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/queue/qtype"

	"github.com/mtgnorton/ws-cluster/logger"

	"github.com/mtgnorton/ws-cluster/config"

	"github.com/mtgnorton/ws-cluster/core/queue/handler"
)

type Options struct {
	Ctx        context.Context
	Config     config.Config
	Topic      string
	Logger     logger.Logger
	Handlers   map[clustermessage.Type]handler.Handle
	Prometheus *wsprometheus.Prometheus
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Ctx:        context.Background(),
		Config:     config.DefaultConfig,
		Topic:      qtype.TopicDefault,
		Logger:     logger.DefaultLogger,
		Handlers:   make(map[clustermessage.Type]handler.Handle),
		Prometheus: wsprometheus.DefaultPrometheus,
	}

	userHandler := handler.NewUserHandler()
	serverHandler := handler.NewServerHandler()

	options.Handlers = map[clustermessage.Type]handler.Handle{
		clustermessage.TypePush:       serverHandler,
		clustermessage.TypeRequest:    userHandler,
		clustermessage.TypeConnect:    userHandler,
		clustermessage.TypeDisconnect: userHandler,
	}
	for _, o := range opts {
		o(&options)
	}

	return options
}

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Ctx = ctx
	}
}

func WithConfig(c config.Config) Option {
	return func(o *Options) {
		o.Config = c
	}
}

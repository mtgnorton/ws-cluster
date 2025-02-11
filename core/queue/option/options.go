package option

import (
	"context"

	"ws-cluster/tools/wsprometheus"

	"ws-cluster/clustermessage"
	"ws-cluster/core/queue/qtype"

	"ws-cluster/logger"

	"ws-cluster/config"

	"ws-cluster/core/queue/handler"
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

	sendToServerHandler := handler.NewSendToServerHandler()
	sendToUserHandler := handler.NewSendToUserHandler()

	options.Handlers = map[clustermessage.Type]handler.Handle{
		clustermessage.TypePush:          sendToUserHandler,
		clustermessage.TypeRequest:       sendToServerHandler,
		clustermessage.TypeConnect:       sendToServerHandler,
		clustermessage.TypeDisconnect:    sendToServerHandler,
		clustermessage.TypeOnlineClients: sendToServerHandler,
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

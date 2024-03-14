package option

import (
	"context"

	"github.com/mtgnorton/ws-cluster/core/queue/qtype"

	"github.com/mtgnorton/ws-cluster/logger"

	"github.com/mtgnorton/ws-cluster/config"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"

	"github.com/mtgnorton/ws-cluster/core/queue/handler"
)

type Options struct {
	Ctx              context.Context
	Config           config.Config
	Topic            string
	Logger           logger.Logger
	Handlers         map[queuemessage.Type]handler.Handle
	MessageProcessor queuemessage.Processor
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Ctx:              context.Background(),
		Config:           config.DefaultConfig,
		Topic:            qtype.TopicDefault,
		Logger:           logger.DefaultLogger,
		Handlers:         make(map[queuemessage.Type]handler.Handle),
		MessageProcessor: queuemessage.DefaultProcessor,
	}

	options.Handlers = map[queuemessage.Type]handler.Handle{
		// queuemessage.TypeSubscribe:   handler.NewSubHandler(),
		queuemessage.TypeRequest: handler.NewReqHandler(),
		// queuemessage.TypeUnsubscribe: handler.NewUnSubHandler(),
		queuemessage.TypePush:       handler.NewPushHandler(),
		queuemessage.TypeConnect:    handler.NewConnectHandler(),
		queuemessage.TypeDisconnect: handler.NewDisconnectHandler(),
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

func WithHandler(t queuemessage.Type, handle handler.Handle) Option {
	return func(o *Options) {
		o.Handlers[t] = handle
	}
}

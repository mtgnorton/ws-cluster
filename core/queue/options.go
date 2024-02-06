package queue

import (
	"context"

	"github.com/mtgnorton/ws-cluster/logger"

	"github.com/mtgnorton/ws-cluster/config"

	"github.com/go-redis/redis/v8"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"

	"github.com/mtgnorton/ws-cluster/core/queue/handler"
)

type Options struct {
	ctx              context.Context
	config           config.Config
	logger           logger.Logger
	queueRedis       *redis.Client
	handlers         map[queuemessage.Type]handler.Handle
	messageProcessor queuemessage.Processor
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		ctx:              context.Background(),
		config:           config.DefaultConfig,
		logger:           logger.DefaultLogger,
		handlers:         make(map[queuemessage.Type]handler.Handle),
		messageProcessor: queuemessage.DefaultProcessor,
	}

	options.handlers = map[queuemessage.Type]handler.Handle{
		queuemessage.TypeSubscribe:   handler.NewSubHandler(),
		queuemessage.TypeRequest:     handler.NewReqHandler(),
		queuemessage.TypeUnsubscribe: handler.NewUnSubHandler(),
		queuemessage.TypePush:        handler.NewPushHandler(),
		queuemessage.TypeConnect:     handler.NewConnectHandler(),
		queuemessage.TypeDisconnect:  handler.NewDisconnectHandler(),
	}
	for _, o := range opts {
		o(&options)
	}
	c := options.config
	options.queueRedis = redis.NewClient(&redis.Options{Addr: c.Values().RedisQueue.Addr, Password: c.Values().RedisQueue.Password, Username: c.Values().RedisQueue.User, DB: c.Values().RedisQueue.DB})

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

func WithHandler(t queuemessage.Type, handle handler.Handle) Option {
	return func(o *Options) {
		o.handlers[t] = handle
	}
}

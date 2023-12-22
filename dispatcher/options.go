package dispatcher

import (
	"context"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/handler"
	"github.com/mtgnorton/ws-cluster/logger"
	"github.com/mtgnorton/ws-cluster/message"
	"github.com/mtgnorton/ws-cluster/parse"
	"go.uber.org/zap"
)

type Options struct {
	ctx      context.Context
	config   config.Config
	logger   *zap.SugaredLogger
	handlers map[message.Type]handler.Handler
	parser   parse.Parser
}
type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		ctx:    context.Background(),
		config: config.DefaultConfig,
		logger: logger.NewZapLogger(config.DefaultConfig),
		handlers: map[message.Type]handler.Handler{
			message.TypeSubscribe:   handler.NewSubscribeHandler(),
			message.TypeUnsubscribe: handler.NewUnSubscribeHandler(),
			message.TypePush:        handler.NewPushHandler(),
		},
		parser: parse.DefaultParser,
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

func WithLogger(l *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

func WithHandlers(hs ...handler.Handler) Option {
	return func(o *Options) {
		for _, h := range hs {
			o.handlers[h.Type()] = h
		}
	}
}

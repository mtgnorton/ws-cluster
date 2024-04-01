package handler

import (
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/core/manager"
	"github.com/mtgnorton/ws-cluster/core/queue"
	"github.com/mtgnorton/ws-cluster/logger"
)

type Option func(*Options)

type Options struct {
	manager manager.Manager
	logger  logger.Logger
	queue   queue.Queue
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		manager: manager.DefaultManager,
		logger:  logger.DefaultLogger,
		queue:   queue.GetQueueInstance(config.DefaultConfig),
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func WithManager(m manager.Manager) Option {
	return func(o *Options) {
		o.manager = m
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

func WithQueue(q queue.Queue) Option {
	return func(o *Options) {
		o.queue = q
	}
}

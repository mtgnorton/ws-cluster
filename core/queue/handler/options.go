package handler

import (
	"github.com/mtgnorton/ws-cluster/core/manager"
	"github.com/mtgnorton/ws-cluster/logger"
)

type Option func(*Options)

type Options struct {
	manager manager.Manager
	logger  logger.Logger
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		manager: manager.DefaultManager,
		logger:  logger.DefaultLogger,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

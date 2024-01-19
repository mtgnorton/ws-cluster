package handler

import (
	"github.com/mtgnorton/ws-cluster/core/manager"
	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/ws/message"
	"go.uber.org/zap"
)

type Option func(*Options)

type Options struct {
	manager manager.Manager
	logger  *zap.SugaredLogger
	parser  message.Parse
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		manager: manager.DefaultManager,
		logger:  shared.DefaultShared.Logger,
		parser:  message.DefaultParser,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

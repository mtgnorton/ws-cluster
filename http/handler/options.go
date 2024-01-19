package handler

import (
	"github.com/mtgnorton/ws-cluster/core/manager"
	"github.com/mtgnorton/ws-cluster/shared"
	"go.uber.org/zap"
)

type Option func(*Options)

type Options struct {
	manager manager.Manager
	logger  *zap.SugaredLogger
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		manager: manager.DefaultManager,
		logger:  shared.DefaultShared.Logger,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

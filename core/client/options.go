package client

import (
	"context"

	"ws-cluster/logger"
)

type Device string

const (
	DeviceAndroid Device = "android"
	DeviceIOS     Device = "ios"
	DeviceWeb     Device = "web"
)

type DeviceInfo struct {
	Number  string `json:"number"`      // 设备号
	Type    Device `json:"device_type"` // 设备类型 android ios web
	Lang    string `json:"lang"`        // 语言
	Country string `json:"country"`     // 国家
	IP      string `json:"ip"`          // ip
}

type Option func(o *Options)

type Options struct {
	ctx    context.Context
	logger logger.Logger
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		ctx:    context.Background(),
		logger: logger.DefaultLogger,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

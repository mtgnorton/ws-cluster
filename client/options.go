package client

import (
	"github.com/bwmarrin/snowflake"
	"github.com/mtgnorton/ws-cluster/shared"
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
	SnowflakeNode *snowflake.Node
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		SnowflakeNode: shared.DefaultShared.SnowflakeNode,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func WithSnowflakeNode(s *snowflake.Node) Option {
	return func(o *Options) {
		o.SnowflakeNode = s
	}
}

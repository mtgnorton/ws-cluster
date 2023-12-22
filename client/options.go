package client

import (
	"github.com/bwmarrin/snowflake"
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

func WithSnowflakeNode(s *snowflake.Node) Option {
	return func(o *Options) {
		o.SnowflakeNode = s
	}
}

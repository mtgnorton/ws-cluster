package logger

import "github.com/mtgnorton/ws-cluster/config"

type Options struct {
	configLog    config.Log
	configSentry config.Sentry
}

type Option func(*Options)

func NewOptions(opts ...Option) *Options {
	o := &Options{
		configLog:    config.DefaultConfig.Values().Log,
		configSentry: config.DefaultConfig.Values().Sentry,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithConfigLog(configLog config.Log) Option {
	return func(o *Options) {
		o.configLog = configLog
	}
}

func WithConfigSentry(configSentry config.Sentry) Option {
	return func(o *Options) {
		o.configSentry = configSentry
	}
}

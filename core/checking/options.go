package checking

import (
	"context"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Ctx      context.Context
	Interval time.Duration
	Redis    *redis.Client
}

func NewOptions(opts ...Option) *Options {
	opt := &Options{
		Ctx:      context.Background(),
		Interval: 10 * time.Second,
		Redis:    shared.GetRedis(),
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

type Option func(*Options)

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Ctx = ctx
	}
}

func WithInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.Interval = interval
	}
}

func WithRedis(redis *redis.Client) Option {
	return func(o *Options) {
		o.Redis = redis
	}
}

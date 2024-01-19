package shared

import (
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/logger"
	"go.uber.org/zap"
)

var DefaultShared = NewShared(config.DefaultConfig)

type Shared struct {
	Config        config.Config
	Logger        *zap.SugaredLogger
	SnowflakeNode *snowflake.Node
	Redis         *redis.Client
	QueueRedis    *redis.Client
}

func NewShared(c config.Config) (s *Shared) {
	snowflakeNode, err := snowflake.NewNode(c.Values().Node)
	if err != nil {
		panic(err)
	}
	defer func() {
		s.Logger.Debugf("configs:%+v", c.Values())
	}()
	return &Shared{
		Config:        c,
		Logger:        logger.NewZapLogger(c),
		SnowflakeNode: snowflakeNode,
		Redis:         redis.NewClient(&redis.Options{Addr: c.Values().Redis.Addr, Password: c.Values().Redis.Password, Username: c.Values().Redis.User, DB: c.Values().Redis.DB}),
		QueueRedis:    redis.NewClient(&redis.Options{Addr: c.Values().RedisQueue.Addr, Password: c.Values().RedisQueue.Password, Username: c.Values().RedisQueue.User, DB: c.Values().RedisQueue.DB}),
	}
}

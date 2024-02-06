package shared

import (
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/shared/jwtws"
)

var DefaultRedis *redis.Client = redis.NewClient(&redis.Options{Addr: config.DefaultConfig.Values().Redis.Addr, Password: config.DefaultConfig.Values().Redis.Password, Username: config.DefaultConfig.Values().Redis.User, DB: config.DefaultConfig.Values().Redis.DB})

var SnowflakeNode *snowflake.Node
var DefaultJwtWs *jwtws.JwtWs = jwtws.NewJwtWs(config.DefaultConfig)

func InitSnowflakeRedisJwt(c config.Config) {
	var err error
	SnowflakeNode, err = snowflake.NewNode(c.Values().Node)
	if err != nil {
		panic(err)
	}

	DefaultRedis = redis.NewClient(&redis.Options{Addr: c.Values().Redis.Addr, Password: c.Values().Redis.Password, Username: c.Values().Redis.User, DB: c.Values().Redis.DB})

	DefaultJwtWs = jwtws.NewJwtWs(c)
}

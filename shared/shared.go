package shared

import (
	"fmt"
	"sync"

	"ws-cluster/config"
	"ws-cluster/shared/kit"

	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
)

var DefaultRedis *redis.Client = redis.NewClient(&redis.Options{Addr: config.DefaultConfig.Values().Redis.Addr, Password: config.DefaultConfig.Values().Redis.Password, Username: config.DefaultConfig.Values().Redis.User, DB: config.DefaultConfig.Values().Redis.DB})

var SnowflakeNode *snowflake.Node

var NodeIDWorker *kit.NodeIDWorker

var ServerIP string = "unknown"
var PublicIP string = "unknown"

var ipOnce sync.Once

var NodeID int64
var nodeIDSyncOnce sync.Once

func InitIP() {
	var err error
	ServerIP, err = kit.GetServerIP()
	if err != nil {
		fmt.Printf("获取本机IP失败:%v\n", err)
	}
	PublicIP, err = kit.GetPublicIP()
	if err != nil {
		fmt.Printf("获取公网IP失败:%v\n", err)
	}
}

func GetIP() (string, string) {
	ipOnce.Do(InitIP)
	return ServerIP, PublicIP

}

// var DefaultJwtWs *jwtws.JwtWs = jwtws.NewJwtWs(config.DefaultConfig)

func InitRedis(c config.Config) {

	DefaultRedis = redis.NewClient(&redis.Options{Addr: c.Values().Redis.Addr, Password: c.Values().Redis.Password, Username: c.Values().Redis.User, DB: c.Values().Redis.DB})
	// DefaultJwtWs = jwtws.NewJwtWs(c)
}

func GetNodeID(c ...config.Config) int64 {
	var (
		nodeID int64
		err    error
	)
	nodeIDSyncOnce.Do(func() {
		if len(c) > 0 {
			nodeID = c[0].Values().Node
		}
		if nodeID > 0 {
			NodeID = nodeID
			SnowflakeNode, err = snowflake.NewNode(nodeID)
			if err != nil {
				panic(err)
			}
			return
		}
		NodeIDWorker, err = kit.NewNodeIDWorker(DefaultRedis)
		if err != nil {
			panic(err)
		}
		nodeID, err = NodeIDWorker.Acquire()
		if err != nil {
			panic(err)
		}
		NodeID = nodeID
		SnowflakeNode, err = snowflake.NewNode(nodeID)
		if err != nil {
			panic(err)
		}
		fmt.Println("使用redis 动态获取nodeID:", nodeID)
	})

	return NodeID
}

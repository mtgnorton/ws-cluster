package shared

import (
	"fmt"
	"sync"

	"ws-cluster/config"
	"ws-cluster/shared/kit"

	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
)

var defaultRedis *redis.Client

var redisOnce sync.Once

func setRedis() {
	redisOnce.Do(func() {

		c := config.DefaultConfig

		defaultRedis = redis.NewClient(&redis.Options{Addr: c.Values().Redis.Addr, Password: c.Values().Redis.Password, Username: c.Values().Redis.User, DB: c.Values().Redis.DB})
		if defaultRedis == nil {
			panic("redis not set")
		}
	})
}
func GetRedis() *redis.Client {
	setRedis()
	return defaultRedis
}

var defaultRedisQueue *redis.Client

var defaultRedisQueueOnce sync.Once

func setDefaultRedisQueue() {
	defaultRedisQueueOnce.Do(func() {
		c := config.DefaultConfig
		defaultRedisQueue = redis.NewClient(&redis.Options{Addr: c.Values().Queue.Redis.Addr, Password: c.Values().Queue.Redis.Password, Username: c.Values().Queue.Redis.User, DB: c.Values().Queue.Redis.DB})
		if defaultRedisQueue == nil {
			panic("redis not set")
		}
	})
}

func GetDefaultRedisQueue() *redis.Client {
	setDefaultRedisQueue()
	return defaultRedisQueue
}

var internalIP string = "unknown"
var publicIP string = "unknown"
var iPOnce sync.Once

func setIP() {
	iPOnce.Do(func() {
		var err error
		internalIP, err = kit.GetServerIP()
		if err != nil {
			fmt.Printf("获取本机IP失败:%v\n", err)
		}
		publicIP, err = kit.GetPublicIP()
		if err != nil {
			fmt.Printf("获取公网IP失败:%v\n", err)
		}
	})
}

func GetInternalIP() string {
	setIP()
	return internalIP
}
func GetPublicIP() string {
	setIP()
	return publicIP
}

var nodeIDWorker *kit.NodeIDWorker

var nodeID int64
var nodeIDSyncOnce sync.Once

func setNodeIDAndWorker(c ...config.Config) {
	nodeIDSyncOnce.Do(func() {
		var err error
		setRedis()
		if len(c) > 0 {
			nodeID = c[0].Values().Node
		}
		if nodeID > 0 {
			return
		}

		nodeIDWorker, err = kit.NewNodeIDWorker(defaultRedis)
		if err != nil {
			panic(err)
		}
		nodeID, err = nodeIDWorker.Acquire()
		if err != nil {
			panic(err)
		}
		fmt.Println("使用redis 动态获取nodeID:", nodeID)
	})
}

func GetNodeID() int64 {
	setNodeIDAndWorker()
	return nodeID
}

func GetNodeIDWorker() *kit.NodeIDWorker {
	setNodeIDAndWorker()
	return nodeIDWorker
}

var snowflakeNode *snowflake.Node
var snowflakeNodeOnce sync.Once

func setSnowflakeNode() {
	snowflakeNodeOnce.Do(func() {
		var err error
		setNodeIDAndWorker()
		snowflakeNode, err = snowflake.NewNode(nodeID)
		if err != nil {
			panic(err)
		}
	})
}

func GetSnowflakeNode() *snowflake.Node {
	setSnowflakeNode()
	return snowflakeNode
}

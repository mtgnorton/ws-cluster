package shared

import (
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/v2/util/guid"
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

type timeoutDetection struct {
	callIDs map[string]struct{}
	deadlock.Mutex
}

var TimeoutDetection = &timeoutDetection{callIDs: make(map[string]struct{})}

// Do 超时检测
// duration 超时时间
// timeoutOperate 超时操作
// begin 操作开始时执行
// end 操作结束时执行
func (t *timeoutDetection) Do(duration time.Duration, timeoutOperate func()) (end func()) {
	// 获取协程id
	callID := guid.S()

	t.Lock()
	t.callIDs[callID] = struct{}{}
	t.Unlock()
	go func() {
		time.AfterFunc(duration, func() {
			t.Lock()
			defer t.Unlock()
			if _, ok := t.callIDs[callID]; ok {
				timeoutOperate()
				delete(t.callIDs, callID)
			}
		})
	}()

	end = func() {
		t.Lock()
		delete(t.callIDs, callID)
		t.Unlock()
	}
	return
}

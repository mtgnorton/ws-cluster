package queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/go-redis/redis/v8"
	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/queue/option"
)

type redisQueue struct {
	opts            option.Options
	redisClient     *redis.Client
	groupName       string
	consumerName    string
	lastReceiveTime time.Time
	mu              deadlock.RWMutex
}

func NewRedisQueue(opts ...option.Option) (q Queue) {
	defer func() {
		go func() {
			_ = q.Consume(q.Options().Ctx, nil)
		}()
	}()
	options := option.NewOptions(opts...)

	c := options.Config
	redisClient := redis.NewClient(&redis.Options{Addr: c.Values().Queue.Redis.Addr, Password: c.Values().Queue.Redis.Password, Username: c.Values().Queue.Redis.User, DB: c.Values().Queue.Redis.DB})

	return &redisQueue{
		opts:         options,
		redisClient:  redisClient,
		groupName:    "group-" + fmt.Sprint(options.Config.Values().Node),
		consumerName: "Redis-Consumer-" + fmt.Sprint(options.Config.Values().Node),
	}
}

func (q *redisQueue) Options() option.Options {
	return q.opts
}
func (q *redisQueue) Publish(ctx context.Context, m *clustermessage.AffairMsg) error {
	messageBytes, err := clustermessage.PackAffair(m)
	if err != nil {
		return err
	}
	topic := q.opts.Topic
	if len(messageBytes) > 1000 {
		// q.opts.Logger.Debugf(ctx, "publish topic:%s,m:%s", topic, messageBytes[:100])

	} else {
		// q.opts.Logger.Debugf(ctx, "publish topic:%s,m:%s", topic, messageBytes)
	}

	return q.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: string(topic),
		Values: map[string]interface{}{
			"m": string(messageBytes),
		},
	}).Err()
}

// Consume 开启一个协程，不断地从redis中读取消息
func (q *redisQueue) Consume(ctx context.Context, _ interface{}) (err error) {
	var (
		queueRedis = q.redisClient
		logger     = q.opts.Logger
		topic      = q.opts.Topic
	)
	r1, err := queueRedis.XGroupCreateMkStream(ctx, topic, q.groupName, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logger.Warnf(ctx, "Redis-Consume failed to create group:%s,err:%v", q.groupName, err)
		return err
	}
	logger.Debugf(ctx, "create group:%s,rs:%v,err:%v", q.groupName, r1, err)

	//defer func() {
	//	if r := recover(); r != nil {
	//		logger.Warnf(ctx, "queue-redis consumer panic:%v", r)
	//	}
	//	logger.Infof(ctx, "queue-redis consumer end")
	//}()

	// 假设宕机了10分钟，那么当再次启动时，这10分钟内的消息直接标记为已完成，不再消费
	lastMessages := queueRedis.XRevRangeN(ctx, topic, "+", "-", 1).Val()
	if len(lastMessages) > 0 {
		logger.Debugf(ctx, "mark lastMessages msg:%v", lastMessages[0].ID)
		_, err = queueRedis.XGroupSetID(ctx, topic, q.groupName, lastMessages[0].ID).Result()
		if err != nil {
			logger.Warnf(ctx, "Redis-Consume failed to ack lastMessages msg:%s,err:%v", lastMessages[0].ID, err)
			return err
		}
	}

	go q.xTrimLoop(ctx)

	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(ctx)
	go q.consume(ctx)

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			q.mu.RLock()
			if time.Since(q.lastReceiveTime) < 20*time.Second {
				q.mu.RUnlock()
				continue
			}
			q.mu.RUnlock()
			logger.Debugf(ctx, "Redis-Consume beyond lastReceiveTime:%v", q.lastReceiveTime)
			panic("Redis-Consume beyond lastReceiveTime")
			cancel()
			ctx, cancel = context.WithCancel(ctx)
			go q.consume(ctx)
		}
	}

}

func (q *redisQueue) consume(ctx context.Context) {
	var (
		queueRedis = q.redisClient
		logger     = q.opts.Logger
		topic      = q.opts.Topic
	)

	f := func() {

		beginTime := time.Now()
		q.mu.Lock()
		q.lastReceiveTime = time.Now()
		q.mu.Unlock()

		var currentID = ">"
		streams, err := queueRedis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    q.groupName,
			Consumer: q.consumerName,
			Streams:  []string{string(topic), currentID},
			Block:    time.Millisecond * 100,
			Count:    50,
		}).Result()

		logger.Debugf(ctx, "Redis-Consume streams msg length:%v,err:%v", len(streams[0].Messages), err)

		if err == redis.Nil {
			// Logger.Debugf(Ctx, "Redis-Consume no msg")
			return
		}
		if err != nil {
			logger.Warnf(ctx, "Redis-Consume failed to read group:%s,err:%v", q.groupName, err)
			return
		}

		//end := shared.TimeoutDetection.Do(time.Second*3, func() {
		//	logger.Errorf(ctx, "Redis-Consume execut  timeout,msg:%+v", streams[0].Messages)
		//})
		defer func() {
			//end()
			logger.Infof(ctx, "Redis-Consume e msg length:%v, exec time %v", len(streams[0].Messages), time.Since(beginTime))
		}()

		for _, msg := range streams[0].Messages {
			concreteMsgString := msg.Values["m"].(string)
			concreteMsg, err := clustermessage.ParseAffair([]byte(concreteMsgString))
			if err != nil {
				logger.Warnf(ctx, "Redis-Consume failed to decode msg: %s,err:%v", concreteMsgString, err)
				continue
			}
			if _, ok := q.opts.Handlers[concreteMsg.Type]; !ok {
				logger.Warnf(ctx, "Redis-Consume failed to find handler for msg: %s", concreteMsgString)
				continue
			}

			if isAck := q.opts.Handlers[concreteMsg.Type].Handle(ctx, concreteMsg); isAck {
				_, err := queueRedis.XAck(ctx, string(topic), q.groupName, msg.ID).Result()
				if err != nil {
					logger.Warnf(ctx, "Redis-Consume failed to ack msg: %s,err:%v", msg.Values["m"].(string), err)
					continue
				}
			} else {
				logger.Warnf(ctx, "Redis-Consume msg: %s,not ack", msg.Values["m"].(string))
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "Redis-Consume exit")
			return
		default:
			f()
		}
	}

}

func (q *redisQueue) xTrimLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c, err := q.redisClient.XTrimMaxLenApprox(ctx, q.opts.Topic, 30000, 500).Result()
			if err != nil {
				q.opts.Logger.Warnf(ctx, "xTrimLoop failed to trim err:%v", err)
				continue
			}
			q.opts.Logger.Debugf(ctx, "xTrimLoop trim count:%d", c)
		}
	}
}

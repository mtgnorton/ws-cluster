package queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ws-cluster/tools/wsprometheus"

	"github.com/sasha-s/go-deadlock"

	"ws-cluster/clustermessage"
	"ws-cluster/core/queue/option"

	"github.com/redis/go-redis/v9"
)

type redisGroupQueue struct {
	opts            option.Options
	redisClient     *redis.Client
	groupName       string
	consumerName    string
	lastReceiveTime time.Time
	mu              deadlock.RWMutex
}

func NewRedisGroupQueue(opts ...option.Option) (q Queue) {
	defer func() {
		go func() {
			_ = q.Consume(q.Options().Ctx, nil)
		}()
	}()
	options := option.NewOptions(opts...)

	c := options.Config
	redisClient := redis.NewClient(&redis.Options{Addr: c.Values().Queue.Redis.Addr, Password: c.Values().Queue.Redis.Password, Username: c.Values().Queue.Redis.User, DB: c.Values().Queue.Redis.DB})

	return &redisGroupQueue{
		opts:         options,
		redisClient:  redisClient,
		groupName:    "group-" + fmt.Sprint(options.Config.Values().Node),
		consumerName: "Redis-Consumer-" + fmt.Sprint(options.Config.Values().Node),
	}
}

func (q *redisGroupQueue) Options() option.Options {
	return q.opts
}
func (q *redisGroupQueue) Publish(ctx context.Context, m *clustermessage.AffairMsg) error {
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

	err = q.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: string(topic),
		Values: map[string]interface{}{
			"m": string(messageBytes),
		},
	}).Err()
	if err != nil {
		q.opts.Logger.Warnf(ctx, "Redis-Publish failed to publish msg:%s,err:%v", messageBytes, err)
		return err
	}
	return nil
}

// Consume 开启一个协程，不断地从redis中读取消息
func (q *redisGroupQueue) Consume(ctx context.Context, _ interface{}) (err error) {
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

func (q *redisGroupQueue) consume(ctx context.Context) {
	var (
		queueRedis = q.redisClient
		logger     = q.opts.Logger
		topic      = q.opts.Topic
		p          = q.opts.Prometheus
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
			Block:    time.Millisecond * 10,
			Count:    100,
		}).Result()

		if len(streams) > 1 {
			logger.Debugf(ctx, "Redis-Consume streams msg length:%v,err:%v", len(streams[0].Messages), err)
		}

		if err == redis.Nil {
			// Logger.Debugf(Ctx, "Redis-Consume no msg")
			return
		}
		if err != nil {
			logger.Warnf(ctx, "Redis-Consume failed to read group:%s,err:%v", q.groupName, err)
			return
		}

		defer func() {
			//end()
			logger.Infof(ctx, "Redis-Consume  msg length:%v, exec time %v ms", len(streams[0].Messages), time.Since(beginTime).Milliseconds())

			averageTime := time.Since(beginTime).Milliseconds() / int64(len(streams[0].Messages))
			if averageTime == 0 {
				averageTime = 1
			}
			_ = p.GetObserve(wsprometheus.MetricQueueHandleDuration, []string{topic}, float64(averageTime))

			_ = q.opts.Prometheus.GetAdd(wsprometheus.MetricQueueHandleTotal, []string{"redis"}, float64(len(streams[0].Messages)))
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

func (q *redisGroupQueue) xTrimLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			beginTime := time.Now()
			c, err := q.redisClient.XTrimMaxLenApprox(ctx, q.opts.Topic, 30000, 5000).Result()
			if err != nil {
				q.opts.Logger.Warnf(ctx, "xTrimLoop failed to trim err:%v", err)
				continue
			}
			xLen := q.redisClient.XLen(ctx, q.opts.Topic).Val()

			q.opts.Logger.Debugf(ctx, "xTrimLoop trim count:%d,remain %d,consume :%v ms", c, xLen, time.Since(beginTime).Milliseconds())
		}
	}
}

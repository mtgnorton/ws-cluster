package queue

import (
	"context"
	"strconv"
	"time"

	"ws-cluster/shared/kit"
	"ws-cluster/tools/wsprometheus"

	"ws-cluster/clustermessage"
	"ws-cluster/core/queue/option"

	"github.com/redis/go-redis/v9"
)

// 使用xread实现的队列
type redisQueue struct {
	opts        option.Options
	redisClient *redis.Client
	startTime   time.Time
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
		opts:        options,
		redisClient: redisClient,
		startTime:   time.Now(),
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

type RedisSamplingData struct {
	Msgs        []redis.XMessage
	ConsumeTime time.Duration
}

// Consume 开启一个协程，不断地从redis中读取消息
func (q *redisQueue) Consume(ctx context.Context, _ interface{}) (err error) {
	var (
		queueRedis = q.redisClient
		logger     = q.opts.Logger
		topic      = q.opts.Topic
		p          = q.opts.Prometheus
	)
	// logger.Debugf(ctx, "ServerIP:%s,PublicIP:%s,Redis general consume has started", shared.ServerIP, shared.PublicIP)

	go q.xTrimLoop(ctx)

	var currentID = "$"
	ch := make(chan RedisSamplingData)

	go kit.Sampling(ch, time.Second*10, 0, func(v RedisSamplingData) {
		logger.Infof(ctx, "Redis-Consume Has started %v,msg length:%v,consume time:%v ms", time.Since(q.startTime), len(v.Msgs), v.ConsumeTime.Milliseconds())
	})

	f := func() {
		beginTime := time.Now()
		streams, err := queueRedis.XRead(ctx, &redis.XReadArgs{
			Streams: []string{string(topic), currentID},
			Count:   500,
			Block:   time.Millisecond * 10,
			ID:      currentID,
		}).Result()

		// 没有消息
		if err == redis.Nil {
			// Logger.Debugf(Ctx, "Redis-Consume no msg")
			return
		}
		if len(streams) == 0 {
			logger.Debugf(ctx, "Redis-Consume  stream length is 0")
			time.Sleep(time.Second)
			return
		}

		if err != nil {
			logger.Warnf(ctx, "Redis-Consume failed to read:%v", err)
			return
		}

		defer func() {
			ch <- RedisSamplingData{
				Msgs:        streams[0].Messages,
				ConsumeTime: time.Since(beginTime),
			}

			averageTime := time.Since(beginTime).Milliseconds() / int64(len(streams[0].Messages))
			if averageTime == 0 {
				averageTime = 1
			}
			if p.Get(wsprometheus.MetricQueueHandleDuration) != nil {
				_ = p.GetObserve(wsprometheus.MetricQueueHandleDuration, []string{topic}, float64(averageTime))
			}

			if p.Get(wsprometheus.MetricQueueHandleTotal) != nil {
				_ = p.GetAdd(wsprometheus.MetricQueueHandleTotal, []string{"redis"}, float64(len(streams[0].Messages)))
			}
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
			q.opts.Handlers[concreteMsg.Type].Handle(ctx, concreteMsg)
			currentID = msg.ID
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "Redis-Consume exit")
			return
		default:
			clear := kit.DoWithTimeout(time.Second*10, func() {
				panic("redis consume timeout")
			})

			f()
			clear()
		}
	}
}

// xTrimLoop 每隔30秒执行一次，删除10分钟前的消息
func (q *redisQueue) xTrimLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			beginTime := time.Now()
			// 计算10分钟前的时间戳
			minTime := time.Now().Add(-10 * time.Minute).UnixMilli()
			c, err := q.redisClient.XTrimMinID(ctx, q.opts.Topic, strconv.FormatInt(minTime, 10)).Result()
			if err != nil {
				q.opts.Logger.Warnf(ctx, "xTrimLoop failed to trim err:%v", err)
				continue
			}
			xLen := q.redisClient.XLen(ctx, q.opts.Topic).Val()

			q.opts.Logger.Debugf(ctx, "xTrimLoop trim count:%d,remain %d,consume :%v ms", c, xLen, time.Since(beginTime).Milliseconds())
		}
	}
}

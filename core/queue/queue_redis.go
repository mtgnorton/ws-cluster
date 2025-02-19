package queue

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"ws-cluster/shared"
	"ws-cluster/shared/kit"
	"ws-cluster/tools/wsprometheus"

	"ws-cluster/clustermessage"
	"ws-cluster/core/queue/option"

	"github.com/redis/go-redis/v9"
)

// 使用xread实现的队列
type redisQueue struct {
	opts         option.Options
	redisClient  *redis.Client
	startTime    time.Time
	publishTimes atomic.Int64
	consumeTimes atomic.Int64
	nodeID       int64
	nodeIP       string
}

func NewRedisQueue(opts ...option.Option) (q Queue) {
	defer func() {
		go func() {
			_ = q.Consume(q.Options().Ctx, nil)
		}()
	}()
	options := option.NewOptions(opts...)

	c := options.Config
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Values().Queue.Redis.Addr,
		Password: c.Values().Queue.Redis.Password,
		Username: c.Values().Queue.Redis.User,
		DB:       c.Values().Queue.Redis.DB,
		Protocol: 3,
		// PoolSize:     200,              // 根据并发量调整（建议为最大并发数的 1.5 倍）
		// MinIdleConns: 50,               // 维持最小空闲连接数
		// PoolTimeout:  30 * time.Second, // 等待连接池连接的超时时间
	})

	ip, _ := shared.GetIP()

	rq := &redisQueue{
		opts:         options,
		redisClient:  redisClient,
		startTime:    time.Now(),
		publishTimes: atomic.Int64{},
		consumeTimes: atomic.Int64{},
		nodeID:       shared.GetNodeID(),
		nodeIP:       ip,
	}
	go rq.monitor(options.Ctx)
	go rq.xTrimLoop(options.Ctx)
	return rq
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
		return kit.TransmitError(err)
	}
	q.publishTimes.Add(1)
	_ = q.opts.Prometheus.GetAdd(wsprometheus.MerticQueueEnter, []string{strconv.FormatInt(q.nodeID, 10), q.nodeIP}, 1)

	return nil
}

type RedisSamplingData struct {
	Msgs         []redis.XMessage
	ConsumeTime  time.Duration
	PublishTimes *atomic.Int64
	ConsumeTimes *atomic.Int64
}

// Consume 开启一个协程，不断地从redis中读取消息
func (q *redisQueue) Consume(ctx context.Context, _ interface{}) (err error) {
	var (
		queueRedis = q.redisClient
		logger     = q.opts.Logger
		topic      = q.opts.Topic
		p          = q.opts.Prometheus
	)

	var currentID = "$"
	ch := make(chan RedisSamplingData)

	go kit.Sampling(ch, time.Second*2, 0, func(v RedisSamplingData) {
		if v.ConsumeTime.Milliseconds() > 1000 {
			logger.Warnf(ctx, "Redis-Consume Has started %v,msg length:%v,consume time:%v ms,publish times:%v,consume times:%v", time.Since(q.startTime), len(v.Msgs), v.ConsumeTime.Milliseconds(), v.PublishTimes.Load(), v.ConsumeTimes.Load())
			return
		}
		logger.Infof(ctx, "Redis-Consume Has started %v,msg length:%v,consume time:%v ms,publish times:%v,consume times:%v", time.Since(q.startTime), len(v.Msgs), v.ConsumeTime.Milliseconds(), v.PublishTimes.Load(), v.ConsumeTimes.Load())
	})

	f := func() {
		streams, err := queueRedis.XRead(ctx, &redis.XReadArgs{
			Streams: []string{string(topic), currentID},
			Count:   500,
			Block:   time.Millisecond * 10,
			ID:      currentID,
		}).Result()
		// 没有消息
		if err == redis.Nil {
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
		beginTime := time.Now()

		defer func() {
			ch <- RedisSamplingData{
				Msgs:         streams[0].Messages,
				ConsumeTime:  time.Since(beginTime),
				PublishTimes: &q.publishTimes,
				ConsumeTimes: &q.consumeTimes,
			}
			var averageTime int64
			if len(streams[0].Messages) > 0 {
				averageTime = time.Since(beginTime).Milliseconds() / int64(len(streams[0].Messages))
			}
			q.consumeTimes.Add(int64(len(streams[0].Messages)))
			_ = p.GetObserve(wsprometheus.MetricQueueHandleDuration, []string{strconv.FormatInt(q.nodeID, 10), q.nodeIP}, float64(averageTime))

			_ = p.GetAdd(wsprometheus.MetricQueueOut, []string{strconv.FormatInt(q.nodeID, 10), q.nodeIP}, float64(len(streams[0].Messages)))

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

func (q *redisQueue) monitor(ctx context.Context) error {
	// 定期打印连接池状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			stats := q.redisClient.PoolStats()
			q.opts.Logger.Infof(ctx, "Redis pool stats: TotalConns=%d, IdleConns=%d Hits=%d,Misses=%d Timeouts=%d",
				stats.TotalConns, stats.IdleConns, stats.Hits, stats.Misses, stats.Timeouts)
		case <-ctx.Done():
			return nil
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

			q.opts.Logger.Infof(ctx, "xTrimLoop trim count:%d,remain %d,consume :%v ms", c, xLen, time.Since(beginTime).Milliseconds())
		}
	}
}

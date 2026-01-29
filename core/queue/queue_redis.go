package queue

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/shared/kit"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/queue/option"

	"github.com/redis/go-redis/v9"
)

// 使用xread实现的队列
type redisQueue struct {
	opts         option.Options
	startTime    time.Time
	publishTimes atomic.Int64
	consumeTimes atomic.Int64
	nodeID       int64
	nodeIP       string
	timeoutTimes atomic.Int64
	msgCh        chan *clustermessage.AffairMsg
}

func NewRedisQueue(opts ...option.Option) (q Queue) {
	defer func() {
		go func() {
			_ = q.Consume(q.Options().Ctx, nil)
		}()
	}()
	options := option.NewOptions(opts...)

	ip := shared.GetInternalIP()

	rq := &redisQueue{
		opts:         options,
		startTime:    time.Now(),
		publishTimes: atomic.Int64{},
		consumeTimes: atomic.Int64{},
		nodeID:       shared.GetNodeID(),
		nodeIP:       ip,
		msgCh:        make(chan *clustermessage.AffairMsg, 100000),
	}
	go rq.monitor(options.Ctx)
	go rq.xTrimLoop(options.Ctx)
	for i := 0; i < options.PublishWorkerCount; i++ {
		go rq.publishLoop(options.Ctx, i)
	}
	return rq
}

func (q *redisQueue) Options() option.Options {
	return q.opts
}
func (q *redisQueue) Publish(ctx context.Context, m *clustermessage.AffairMsg) error {
	timer := time.NewTimer(time.Second * 5)
	defer timer.Stop()
	select {
	case q.msgCh <- m:
		return nil
	case <-ctx.Done():
		q.opts.Logger.Warnf(ctx, "Redis-Publish canceled, drop msg:%+v", m)
		return nil
	case <-timer.C:
		q.opts.Logger.Warnf(ctx, "Redis-Publish timeout, drop msg:%+v", m)
		return nil
	}
}

func (q *redisQueue) publishLoop(ctx context.Context, workerID int) {

	logger := q.opts.Logger
	batchSize := q.opts.PublishBatchSize
	tickerMs := q.opts.PublishTickerMs
	cache := make([]*clustermessage.AffairMsg, 0, batchSize)
	ticker := time.NewTicker(tickerMs)
	defer ticker.Stop()

	flush := func() {
		if len(cache) > 0 {
			q.publish(ctx, cache)
			cache = cache[:0]
		}
	}

	for {
		select {
		case <-ctx.Done():
			if len(cache) > 0 {
				flushCtx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				q.publish(flushCtx, cache)
				cancel()
			}
			logger.Infof(ctx, "Redis-publishLoop worker-%d exit", workerID)
			return
		case m := <-q.msgCh:
			cache = append(cache, m)
			for len(cache) < batchSize {
				select {
				case msg := <-q.msgCh:
					cache = append(cache, msg)
				default:
					goto batchReady
				}
			}
		batchReady:
			if len(cache) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (q *redisQueue) publish(ctx context.Context, msgs []*clustermessage.AffairMsg) {
	logger := q.opts.Logger
	pipe := q.opts.RedisClient.Pipeline()
	topic := q.opts.Topic
	validCount := 0

	for _, m := range msgs {
		messageBytes, err := clustermessage.PackAffair(m)
		if err != nil {
			logger.Infof(ctx, "Redis-publish msg:%+v packAffair failed,error: %v", m, err)
			continue
		}
		_ = pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: string(topic),
			Values: map[string]interface{}{
				"m": string(messageBytes),
			},
		})
		validCount++
	}

	if validCount == 0 {
		return
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		logger.Warnf(ctx, "Redis-publish pipe.Exec failed, error:%v", err)
		return
	}
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			logger.Warnf(ctx, "Redis-publish exec cmd xadd failed, error:%v", cmd.Err())
		}
	}
	q.publishTimes.Add(int64(validCount))
	_ = q.opts.Prometheus.GetAdd(wsprometheus.MerticQueueEnter, []string{strconv.FormatInt(q.nodeID, 10), q.nodeIP}, float64(validCount))

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
		queueRedis = q.opts.RedisClient
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
			timeoutDetect := kit.DoWithTimeout(time.Second*10, func() {
				panic("timeout")
			})
			f()
			timeoutDetect()
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
			stats := q.opts.RedisClient.PoolStats()
			q.opts.Logger.Infof(ctx, "Redis pool stats: TotalConns=%d, IdleConns=%d Hits=%d,Misses=%d Timeouts=%d, msgCh len=%d cap=%d",
				stats.TotalConns, stats.IdleConns, stats.Hits, stats.Misses, stats.Timeouts, len(q.msgCh), cap(q.msgCh))
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
			c, err := q.opts.RedisClient.XTrimMinID(ctx, q.opts.Topic, strconv.FormatInt(minTime, 10)).Result()
			if err != nil {
				q.opts.Logger.Warnf(ctx, "xTrimLoop failed to trim err:%v", err)
				continue
			}
			xLen := q.opts.RedisClient.XLen(ctx, q.opts.Topic).Val()

			q.opts.Logger.Infof(ctx, "xTrimLoop trim count:%d,remain %d,consume :%v ms", c, xLen, time.Since(beginTime).Milliseconds())
		}
	}
}

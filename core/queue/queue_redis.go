package queue

import (
	"context"
	"strconv"
	"strings"
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
	metricLabels []string
	msgCh        chan *clustermessage.AffairMsg
	lastSlowLog  atomic.Int64
}

func NewRedisQueue(opts ...option.Option) (q Queue) {
	defer func() {
		go func() {
			_ = q.Consume(q.Options().Ctx, nil)
		}()
	}()
	options := option.NewOptions(opts...)

	ip := shared.GetInternalIP()
	nodeID := shared.GetNodeID()

	rq := &redisQueue{
		opts:         options,
		startTime:    time.Now(),
		publishTimes: atomic.Int64{},
		consumeTimes: atomic.Int64{},
		nodeID:       nodeID,
		nodeIP:       ip,
		metricLabels: []string{strconv.FormatInt(nodeID, 10), ip},
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
	beginTime := time.Now()
	select {
	case q.msgCh <- m:
		_ = q.opts.Prometheus.GetObserve(wsprometheus.MetricQueuePublishWaitDuration, q.metricLabels, float64(time.Since(beginTime).Microseconds())/1000.0)
		return nil
	default:
	}

	timer := time.NewTimer(time.Second * 5)
	defer timer.Stop()
	select {
	case q.msgCh <- m:
		waitMs := float64(time.Since(beginTime).Microseconds()) / 1000.0
		_ = q.opts.Prometheus.GetObserve(wsprometheus.MetricQueuePublishWaitDuration, q.metricLabels, waitMs)
		if waitMs >= 100 && kit.AllowByInterval(&q.lastSlowLog, 2*time.Second) {
			q.opts.Logger.Warnf(ctx, "Redis-Publish local queue wait=%0.2fms,len=%d,cap=%d,type=%s,payload=%s", waitMs, len(q.msgCh), cap(q.msgCh), m.Type, kit.LogSnippet(m.Payload, 240))
		}
		return nil
	case <-ctx.Done():
		q.opts.Logger.Warnf(ctx, "Redis-Publish canceled, drop msg:%+v", m)
		_ = q.opts.Prometheus.GetAdd(wsprometheus.MetricQueueDrop, q.metricLabels, 1)
		return nil
	case <-timer.C:
		q.opts.Logger.Warnf(ctx, "Redis-Publish timeout, drop msg:%+v", m)
		_ = q.opts.Prometheus.GetAdd(wsprometheus.MetricQueueDrop, q.metricLabels, 1)
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
	beginTime := time.Now()
	pipe := q.opts.RedisClient.Pipeline()
	topic := string(q.opts.Topic)
	validCount := 0

	for _, m := range msgs {
		messageBytes, err := clustermessage.PackAffair(m)
		if err != nil {
			logger.Infof(ctx, "Redis-publish msg:%+v packAffair failed,error: %v", m, err)
			continue
		}
		_ = pipe.Do(ctx, "XADD", topic, "*", "m", messageBytes)
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
	_ = q.opts.Prometheus.GetAdd(wsprometheus.MerticQueueEnter, q.metricLabels, float64(validCount))
	_ = q.opts.Prometheus.GetObserve(
		wsprometheus.MertricQueueEnterDuration,
		q.metricLabels,
		float64(time.Since(beginTime).Microseconds())/1000.0/float64(validCount),
	)

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

	f := func() {
		streams, err := queueRedis.XRead(ctx, &redis.XReadArgs{
			Streams: []string{topic},
			Count:   500,
			Block:   time.Millisecond * 10,
			ID:      currentID,
		}).Result()
		// 没有消息
		if err == redis.Nil {
			return
		}
		if err != nil {
			logger.Warnf(ctx, "Redis-Consume failed to read:%v", err)
			return
		}
		if len(streams) == 0 {
			return
		}
		beginTime := time.Now()
		messageCount := len(streams[0].Messages)
		batchSamples := make([]string, 0, 3)

		defer func() {
			var averageTime int64
			if messageCount > 0 {
				averageTime = time.Since(beginTime).Milliseconds() / int64(messageCount)
			}
			q.consumeTimes.Add(int64(messageCount))
			_ = p.GetObserve(wsprometheus.MetricQueueHandleDuration, q.metricLabels, float64(averageTime))
			_ = p.GetAdd(wsprometheus.MetricQueueOut, q.metricLabels, float64(messageCount))
			totalMs := float64(time.Since(beginTime).Microseconds()) / 1000.0
			if totalMs >= 1000 && kit.AllowByInterval(&q.lastSlowLog, 2*time.Second) {
				logger.Warnf(ctx, "Redis-Consume slow batch=%0.2fms,msg_count=%d,publish_times=%d,consume_times=%d,current_id=%s,samples=%s", totalMs, messageCount, q.publishTimes.Load(), q.consumeTimes.Load(), currentID, kit.JoinLogSnippets(batchSamples))
			}
		}()

		for _, msg := range streams[0].Messages {
			msgType := "unknown"
			lagMs := streamMessageLagMs(msg.ID)
			rawMsg, ok := msg.Values["m"]
			if !ok {
				logger.Warnf(ctx, "Redis-Consume failed to read msg field m, msgID:%s", msg.ID)
				continue
			}
			var concreteMsgBytes []byte
			switch v := rawMsg.(type) {
			case string:
				concreteMsgBytes = []byte(v)
			case []byte:
				concreteMsgBytes = v
			default:
				logger.Warnf(ctx, "Redis-Consume unsupported msg field type:%T, msgID:%s", rawMsg, msg.ID)
				continue
			}

			concreteMsg, err := clustermessage.ParseAffair(concreteMsgBytes)
			if err != nil {
				if len(batchSamples) < 3 {
					batchSamples = append(batchSamples, kit.LogSnippet(concreteMsgBytes, 160))
				}
				logger.Warnf(ctx, "Redis-Consume failed to decode msg: %s,err:%v", string(concreteMsgBytes), err)
				continue
			}
			msgType = string(concreteMsg.Type)
			if len(batchSamples) < 3 {
				batchSamples = append(batchSamples, kit.LogSnippet(concreteMsg.Payload, 160))
			}
			_ = p.GetObserve(wsprometheus.MetricQueueLagDuration, append(q.metricLabels, msgType), lagMs)
			if lagMs >= 1000 && kit.AllowByInterval(&q.lastSlowLog, 2*time.Second) {
				logger.Warnf(ctx, "Redis-Consume lag=%0.2fms,msg_id=%s,type=%s,msg_count=%d,payload=%s", lagMs, msg.ID, msgType, messageCount, kit.LogSnippet(concreteMsg.Payload, 240))
			}
			if _, ok := q.opts.Handlers[concreteMsg.Type]; !ok {
				logger.Warnf(ctx, "Redis-Consume failed to find handler for msg: %s", string(concreteMsgBytes))
				continue
			}
			dispatchBegin := time.Now()
			q.opts.Handlers[concreteMsg.Type].Handle(ctx, concreteMsg)
			dispatchMs := float64(time.Since(dispatchBegin).Microseconds()) / 1000.0
			_ = p.GetObserve(wsprometheus.MetricQueueDispatchDuration, append(q.metricLabels, msgType), dispatchMs)
			if dispatchMs >= 50 && kit.AllowByInterval(&q.lastSlowLog, 2*time.Second) {
				logger.Warnf(ctx, "Redis-Consume dispatch slow=%0.2fms,type=%s,msg_id=%s,payload=%s", dispatchMs, msgType, msg.ID, kit.LogSnippet(concreteMsg.Payload, 240))
			}
			currentID = msg.ID
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

func (q *redisQueue) monitor(ctx context.Context) error {
	// 定期打印连接池状态
	ticker := time.NewTicker(30 * time.Second)
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

func streamMessageLagMs(id string) float64 {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 0 {
		return 0
	}
	createdAtMs, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}
	return float64(time.Now().UnixMilli() - createdAtMs)
}

// xTrimLoop 高频率近似修剪，避免大批量精确删除导致抖动
func (q *redisQueue) xTrimLoop(ctx context.Context) {
	const (
		trimInterval   = time.Second * 5
		logInterval    = time.Second * 30
		trimLimitBatch = int64(20000)
	)
	ticker := time.NewTicker(trimInterval)
	defer ticker.Stop()
	lastLog := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			beginTime := time.Now()
			// 计算10分钟前的时间戳
			minTime := time.Now().Add(-10 * time.Minute).UnixMilli()
			c, err := q.opts.RedisClient.XTrimMinIDApprox(ctx, q.opts.Topic, strconv.FormatInt(minTime, 10), trimLimitBatch).Result()
			if err != nil {
				q.opts.Logger.Warnf(ctx, "xTrimLoop failed to trim err:%v", err)
				continue
			}
			if time.Since(lastLog) >= logInterval {
				xLen := q.opts.RedisClient.XLen(ctx, q.opts.Topic).Val()
				q.opts.Logger.Infof(ctx, "xTrimLoop trim count:%d,remain %d,consume :%v ms", c, xLen, time.Since(beginTime).Milliseconds())
				lastLog = time.Now()
			}
		}
	}
}

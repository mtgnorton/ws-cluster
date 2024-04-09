package queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/queue/option"
)

type redisQueue struct {
	opts         option.Options
	redisClient  *redis.Client
	groupName    string
	consumerName string
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
		consumerName: "consumer-" + fmt.Sprint(options.Config.Values().Node),
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
	queueRedis := q.redisClient
	logger := q.opts.Logger
	topic := q.opts.Topic
	r1, err := queueRedis.XGroupCreateMkStream(ctx, topic, q.groupName, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logger.Warnf(ctx, "consume failed to create group:%s,err:%v", q.groupName, err)
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
			logger.Warnf(ctx, "consume failed to ack lastMessages msg:%s,err:%v", lastMessages[0].ID, err)
			return err
		}
	}

	go q.xTrimLoop(ctx)

	for {
		var currentID = ">"
		streams, err := queueRedis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    q.groupName,
			Consumer: q.consumerName,
			Streams:  []string{string(topic), currentID},
			Block:    time.Millisecond * 100,
			Count:    100,
		}).Result()

		logger.Debugf(ctx, "consume streams msg length:%v,err:%v", len(streams), err)
		if err == redis.Nil {
			// Logger.Debugf(Ctx, "consume no msg")
			continue
		}

		if err != nil {
			logger.Warnf(ctx, "consume failed to read group:%s,err:%v", q.groupName, err)
			continue
		}

		for _, msg := range streams[0].Messages {
			concreteMsgString := msg.Values["m"].(string)
			concreteMsg, err := clustermessage.ParseAffair([]byte(concreteMsgString))
			if err != nil {
				logger.Warnf(ctx, "consume failed to decode msg: %s,err:%v", concreteMsgString, err)
				continue
			}
			if _, ok := q.opts.Handlers[concreteMsg.Type]; !ok {
				logger.Warnf(ctx, "consume failed to find handler for msg: %s", concreteMsgString)
				continue
			}
			if isAck := q.opts.Handlers[concreteMsg.Type].Handle(ctx, concreteMsg); isAck {
				_, err := queueRedis.XAck(ctx, string(topic), q.groupName, msg.ID).Result()
				if err != nil {
					logger.Warnf(ctx, "consume failed to ack msg: %s,err:%v", msg.Values["m"].(string), err)
					continue
				}
			} else {
				logger.Warnf(ctx, "consume msg: %s,not ack", msg.Values["m"].(string))
			}
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
			}
			q.opts.Logger.Debugf(ctx, "xTrimLoop trim count:%d", c)
		}
	}
}

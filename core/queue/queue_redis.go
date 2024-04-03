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
	r1, err := queueRedis.XGroupCreateMkStream(ctx, string(topic), q.groupName, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logger.Warnf(ctx, "consume failed to create group:%s,err:%v", q.groupName, err)
		return err
	}
	logger.Debugf(ctx, "create group:%s,rs:%v,err:%v", q.groupName, r1, err)

	defer func() {
		if r := recover(); r != nil {
			logger.Warnf(ctx, "queue-redis consumer panic:%v", r)
		}
		logger.Infof(ctx, "queue-redis consumer end")
	}()

	for {

		var currentID = ">"

		streams, err := queueRedis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    q.groupName,
			Consumer: q.consumerName,
			Streams:  []string{string(topic), currentID},
			Block:    time.Millisecond * 10,
			Count:    100,
		}).Result()

		if err == redis.Nil {
			// Logger.Debugf(Ctx, "consume no msg")
			continue
		}

		if err != nil {
			logger.Warnf(ctx, "consume failed to read group:%s,err:%v", q.groupName, err)
			continue
		}

		for _, msg := range streams[0].Messages {
			//logger.Debugf(ctx, "consume topic:%s,msg id:%s,values:%s", topic, msg.ID, msg.Values)
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

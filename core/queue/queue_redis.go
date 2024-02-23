package queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mtgnorton/ws-cluster/core/queue/option"
	"github.com/mtgnorton/ws-cluster/message/queuemessage"

	"github.com/go-redis/redis/v8"
)

type RedisQueue struct {
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
	redisClient := redis.NewClient(&redis.Options{Addr: c.Values().RedisQueue.Addr, Password: c.Values().RedisQueue.Password, Username: c.Values().RedisQueue.User, DB: c.Values().RedisQueue.DB})

	return &RedisQueue{
		opts:         options,
		redisClient:  redisClient,
		groupName:    "group-" + fmt.Sprint(options.Config.Values().Node),
		consumerName: "consumer-" + fmt.Sprint(options.Config.Values().Node),
	}
}

func (q *RedisQueue) Options() option.Options {
	return q.opts
}
func (q *RedisQueue) Publish(ctx context.Context, message *queuemessage.Message) error {
	messageBytes, err := q.opts.MessageProcessor.Encode(message)
	if err != nil {
		return err
	}
	topic := q.opts.Topic
	q.opts.Logger.Debugf(ctx, "publish topic:%s,message:%s", topic, messageBytes)

	return q.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: string(topic),
		Values: map[string]interface{}{
			"m": string(messageBytes),
		},
	}).Err()
}

// Consume 开启一个协程，不断地从redis中读取消息
func (q *RedisQueue) Consume(ctx context.Context, _ interface{}) (err error) {
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
			Block:    time.Millisecond * 500,
			Count:    100,
		}).Result()

		if err == redis.Nil {
			// Logger.Debugf(Ctx, "consume no message")
			continue
		}
		//if err != nil && strings.Contains(err.Error(), "i/o timeout") {
		//	err = nil
		//}
		if err != nil {
			logger.Warnf(ctx, "consume failed to read group:%s,err:%v", q.groupName, err)
			continue
		}

		for _, message := range streams[0].Messages {
			logger.Debugf(ctx, "consume topic:%s,message id:%s,values:%s", topic, message.ID, message.Values)

			concreteMsgString := message.Values["m"].(string)
			concreteMsg, err := q.opts.MessageProcessor.Decode([]byte(concreteMsgString))
			if err != nil {
				logger.Warnf(ctx, "consume failed to decode message: %s,err:%v", concreteMsgString, err)
				continue
			}
			if isAck := q.opts.Handlers[concreteMsg.Type].Handle(ctx, *concreteMsg); isAck {
				_, err := queueRedis.XAck(ctx, string(topic), q.groupName, message.ID).Result()
				if err != nil {
					logger.Warnf(ctx, "consume failed to ack message: %s,err:%v", message.Values["m"].(string), err)
					continue
				}
			} else {
				logger.Warnf(ctx, "consume message: %s,not ack", message.Values["m"].(string))
			}
		}
	}
}

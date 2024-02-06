package queue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"

	"github.com/go-redis/redis/v8"
)

type RedisQueue struct {
	opts         Options
	groupName    string
	consumerName string
}

func NewRedisQueue(opts ...Option) (q Queue) {

	defer func() {
		go func() {
			_ = q.Consume(q.Options().ctx, TopicDefault)

		}()
	}()
	options := NewOptions(opts...)

	return &RedisQueue{
		opts:         options,
		groupName:    "group-" + fmt.Sprint(options.config.Values().Node),
		consumerName: "consumer-" + fmt.Sprint(options.config.Values().Node),
	}
}

func (q *RedisQueue) Options() Options {
	return q.opts
}
func (q *RedisQueue) Publish(ctx context.Context, topic Topic, message *queuemessage.Message) error {
	messageBytes, err := q.opts.messageProcessor.Encode(message)
	if err != nil {
		return err
	}
	q.opts.logger.Debugf(ctx, "publish topic:%s,message:%s", topic, messageBytes)

	return q.opts.queueRedis.XAdd(ctx, &redis.XAddArgs{
		Stream: string(topic),
		Values: map[string]interface{}{
			"m": string(messageBytes),
		},
	}).Err()
}

func (q *RedisQueue) Consume(ctx context.Context, topic Topic) (err error) {
	queueRedis := q.opts.queueRedis
	logger := q.opts.logger
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
			// logger.Debugf(ctx, "consume no message")
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

			// 可能的情况
			// 1. 该消息不属于该服务端，此时isAck为false，不需要ack
			// 2. 该消息属于该服务端，但是处理失败，此时isAck为false，不需要ack
			concreteMsgString := message.Values["m"].(string)
			concreteMsg, err := q.opts.messageProcessor.Decode([]byte(concreteMsgString))
			if err != nil {
				logger.Warnf(ctx, "consume failed to decode message: %s,err:%v", concreteMsgString, err)
				continue
			}
			if isAck := q.opts.handlers[concreteMsg.Type].Handle(ctx, *concreteMsg); isAck {
				_, err := queueRedis.XAck(ctx, string(topic), q.groupName, message.ID).Result()
				if err != nil {
					logger.Warnf(ctx, "consume failed to ack message: %s,err:%v", message.Values["m"].(string), err)
					continue
				}
			} else {
				logger.Warnf(ctx, "consume failed to handle message: %s", message.Values["m"].(string))
			}
		}
	}
}

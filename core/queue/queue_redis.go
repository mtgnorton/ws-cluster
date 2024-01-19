package queue

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
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
		groupName:    "group-" + fmt.Sprint(options.shared.Config.Values().Node),
		consumerName: "consumer-" + fmt.Sprint(options.shared.Config.Values().Node),
	}
}

func (q *RedisQueue) Options() Options {
	return q.opts
}
func (q *RedisQueue) Publish(ctx context.Context, topic Topic, message []byte) error {
	q.opts.shared.Logger.Debugf("publish topic:%s,message:%s", topic, message)
	return q.opts.shared.QueueRedis.XAdd(ctx, &redis.XAddArgs{
		Stream: string(topic),
		Values: map[string]interface{}{
			"m": string(message),
		},
	}).Err()
}

func (q *RedisQueue) Consume(ctx context.Context, topic Topic) (err error) {
	queueRedis := q.opts.shared.QueueRedis
	logger := q.opts.shared.Logger
	r1, err := queueRedis.XGroupCreateMkStream(ctx, string(topic), q.groupName, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		logger.Warnf("consume failed to create group:%s,err:%v", q.groupName, err)
		return err
	}
	logger.Debugf("create group:%s,rs:%v,err:%v", q.groupName, r1, err)

	var (
		lastID = "0-0"
		check  = true
	)

	for {
		var currentID string
		if check {
			currentID = lastID
		} else {
			currentID = ">"
		}

		streams, err := queueRedis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    q.groupName,
			Consumer: q.consumerName,
			Streams:  []string{string(topic), currentID},
			Block:    1,
			Count:    50,
		}).Result()

		if err != nil && strings.Contains(err.Error(), "i/o timeout") {
			err = nil
		}
		if err != nil {
			logger.Warnf("consume failed to read group:%s,err:%v", q.groupName, err)
			continue
		}
		if len(streams) == 0 || len(streams[0].Messages) == 0 {
			check = false
			continue
		}
		for _, message := range streams[0].Messages {
			logger.Debugf("consume topic:%s,message id:%s,values:%s", topic, message.ID, message.Values)
			//msg, err := q.opts.parser.Parse([]byte(message.Values["m"].(string)))
			//if err != nil {
			//	logger.Warnf("consume failed to parse message: %s,err:%v", message.Values["m"].(string), err)
			//	continue
			//}

			// 可能的情况
			// 1. 该消息不属于该服务端，此时isAck为false，不需要ack
			// 2. 该消息属于该服务端，但是处理失败，此时isAck为false，不需要ack

			if isAck := q.opts.handler.Handle([]byte(message.Values["m"].(string))); isAck {
				_, err := queueRedis.XAck(ctx, string(topic), q.groupName, message.ID).Result()
				if err != nil {
					logger.Warnf("consume failed to ack message: %s,err:%v", message.Values["m"].(string), err)
					continue
				}
			}

			lastID = message.ID
		}
	}
}

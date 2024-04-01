package queue

import (
	"context"
	"sync"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/config"

	"github.com/mtgnorton/ws-cluster/core/queue/option"
)

var QueueInstance Queue

// Queue 队列接口
// 使用发布订阅模式
// 所有的服务端都能够收到所有的消息
type Queue interface {
	Options() option.Options
	Publish(ctx context.Context, message *clustermessage.AffairMsg) error

	// ack的逻辑
	// 1. 该消息不属于该服务端，需要ack
	// 2. 该消息属于该服务端，但是处理失败，此时不进行ack,等待重试
	Consume(ctx context.Context, integration interface{}) error // integration 是为了兼容不同的queue,具体的类型由具体的queue决定
}

type Topic string

const (
	TopicDefault = "default"
)

type QueueType string

const (
	QueueTypeRedis = "redis"

	QueueTypeKafka = "kafka"
)

var once sync.Once

func GetQueueInstance(c config.Config) Queue {
	once.Do(func() {
		if c.Values().Queue.Use == QueueTypeKafka {
			QueueInstance = NewKafkaQueue(option.WithConfig(c))
		} else {
			QueueInstance = NewRedisQueue(option.WithConfig(c))
		}
	})
	return QueueInstance
}

package queue

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

var DefaultQueue = NewRedisQueue()

type Queue interface {
	Options() Options
	Publish(ctx context.Context, topic Topic, message *queuemessage.Message) error
	Consume(ctx context.Context, topic Topic) error
}

type Topic string

const (
	TopicDefault = "default"
)

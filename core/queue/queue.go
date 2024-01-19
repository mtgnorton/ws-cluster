package queue

import (
	"context"
)

var DefaultQueue = NewRedisQueue()

type Queue interface {
	Options() Options
	Publish(ctx context.Context, topic Topic, message []byte) error
	Consume(ctx context.Context, topic Topic) error
}

type Topic string

const (
	TopicDefault = "default"
)

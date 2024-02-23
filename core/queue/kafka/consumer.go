package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

type ConsumerGroupHandler interface {
	sarama.ConsumerGroupHandler
	WaitReady()
	Reset()
}

type ConsumerGroup struct {
	cg sarama.ConsumerGroup
}

type ConsumerSessionMessage struct {
	Session sarama.ConsumerGroupSession
	Message *sarama.ConsumerMessage
}

func (c *ConsumerGroup) Close() error {
	return c.cg.Close()
}

func NewConsumerGroup(ctx context.Context, broker string, topics []string, group string, handler ConsumerGroupHandler) (*ConsumerGroup, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_2_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	groupClient, err := sarama.NewConsumerGroup([]string{broker}, group, cfg)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			err := groupClient.Consume(ctx, topics, handler)
			if err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					break
				} else {
					panic(err)
				}
			}
			handler.Reset()
		}
	}()

	handler.WaitReady() // Await till the consumer has been set up

	return &ConsumerGroup{
		cg: groupClient,
	}, nil
}

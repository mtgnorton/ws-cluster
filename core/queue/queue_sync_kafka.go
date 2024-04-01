package queue

import (
	"context"
	"fmt"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/queue/kafka"

	"github.com/mtgnorton/ws-cluster/core/queue/option"

	"github.com/IBM/sarama"
)

type KafkaSyncQueue struct {
	opts     option.Options
	producer *kafka.Producer
	consumer *kafka.ConsumerGroup
}

func NewKafkaQueue(opts ...option.Option) (q Queue) {
	options := option.NewOptions(opts...)

	producer, err := kafka.NewProducer(options.Config.Values().Queue.Kafka.Broker)
	if err != nil {
		panic(err)
	}
	k := &KafkaSyncQueue{
		opts:     options,
		producer: producer,
	}
	consumer, err := kafka.NewConsumerGroup(options.Ctx, options.Config.Values().Kafka.Broker, []string{options.Topic}, "group-"+fmt.Sprint(options.Config.Values().Node), kafka.NewSyncConsumerGroupHandler(options.Ctx, k.Consume))
	if err != nil {
		panic(err)
	}
	k.consumer = consumer
	return k
}

func (k KafkaSyncQueue) Options() option.Options {
	return k.opts
}

func (k KafkaSyncQueue) Publish(ctx context.Context, m *clustermessage.AffairMsg) error {
	messageBytes, err := clustermessage.PackAffair(m)
	if err != nil {
		return err
	}
	topic := k.opts.Topic
	k.opts.Logger.Debugf(ctx, "publish topic:%s,m:%s", topic, messageBytes)
	k.producer.P.Input() <- &sarama.ProducerMessage{
		Topic: string(topic),
		Value: sarama.ByteEncoder(messageBytes),
	}
	return nil
}

// Consume 该方法的上游一直读取消息，上游会将每条消息传递给该方法
func (k KafkaSyncQueue) Consume(ctx context.Context, integration interface{}) (err error) {

	defer func() {
		if err != nil {
			k.opts.Logger.Errorf(ctx, "consume error: %s", err)
		}
	}()
	logger := k.opts.Logger
	consumerSessionMessage, ok := integration.(kafka.ConsumerSessionMessage)
	session := consumerSessionMessage.Session
	if !ok {
		logger.Infof(ctx, "consume integration type error")
		return
	}
	logger.Debugf(ctx, "consume message: %s", consumerSessionMessage.Message.Value)
	concreteMsg, err := clustermessage.ParseAffair(consumerSessionMessage.Message.Value)
	if err != nil {
		return err
	}
	if isAck := k.opts.Handlers[concreteMsg.Type].Handle(ctx, concreteMsg); isAck {
		logger.Debugf(ctx, "consume message: %s,ack", consumerSessionMessage.Message.Value)
		session.MarkMessage(consumerSessionMessage.Message, "")
	} else {
		logger.Warnf(ctx, "consume message: %s,not ack", consumerSessionMessage.Message.Value)
	}
	return nil
}

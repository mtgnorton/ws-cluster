package kafka

import (
	"github.com/IBM/sarama"
)

type Producer struct {
	P sarama.AsyncProducer
}

func NewProducer(broker string) (*Producer, error) {
	p, err := sarama.NewAsyncProducer([]string{broker}, sarama.NewConfig())
	if err != nil {
		return nil, err
	}
	return &Producer{P: p}, nil
}

func (p *Producer) Close() error {
	if p != nil {
		return p.P.Close()
	}
	return nil
}

package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

type syncConsumerGroupHandler struct {
	ctx   context.Context
	ready chan bool
	cb    func(context.Context, interface{}) error
}

func NewSyncConsumerGroupHandler(ctx context.Context, cb func(context.Context, interface{}) error) ConsumerGroupHandler {
	handler := syncConsumerGroupHandler{
		ctx:   ctx,
		ready: make(chan bool),
		cb:    cb,
	}
	return &handler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *syncConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *syncConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *syncConsumerGroupHandler) WaitReady() {
	<-h.ready
}

func (h *syncConsumerGroupHandler) Reset() {
	h.ready = make(chan bool)
}

func (h *syncConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	claimMsgChan := claim.Messages()

	for message := range claimMsgChan {
		m := ConsumerSessionMessage{
			Session: session,
			Message: message,
		}
		_ = h.cb(h.ctx, m)

	}

	return nil
}

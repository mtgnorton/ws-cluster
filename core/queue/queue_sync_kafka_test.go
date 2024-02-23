package queue

import (
	"context"
	"testing"
	"time"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

func TestKafkaQueue(t *testing.T) {
	q := NewKafkaQueue()
	ctx := context.Background()
	_ = q.Publish(ctx, &queuemessage.Message{
		Type:           queuemessage.TypeRequest,
		Identification: "111",
		PID:            "111",
		Payload:        nil,
	})
	time.Sleep(time.Second)
}

package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

var DefaultPushHandler = NewPushHandler()

type Handle interface {
	Handle(ctx context.Context, payload queuemessage.Message) (isAck bool)
}

package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/clustermessage"
)

var DefaultPushHandler = NewSendToUserHandler()

type Handle interface {
	Handle(ctx context.Context, payload *clustermessage.AffairMsg) (isAck bool)
}

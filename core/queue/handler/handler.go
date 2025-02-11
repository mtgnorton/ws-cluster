package handler

import (
	"context"

	"ws-cluster/clustermessage"
)

var DefaultPushHandler = NewSendToUserHandler()

type Handle interface {
	Handle(ctx context.Context, payload *clustermessage.AffairMsg) (isAck bool)
}

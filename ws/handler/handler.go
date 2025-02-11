package handler

import (
	"context"

	"ws-cluster/clustermessage"
	"ws-cluster/core/client"
)

var DefaultHandler = NewWsHandler()

type Handle interface {
	Handle(ctx context.Context, client client.Client, message *clustermessage.AffairMsg)
}

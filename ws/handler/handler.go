package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
)

var DefaultHandler = NewWsHandler()

type Handle interface {
	Handle(ctx context.Context, client client.Client, message *clustermessage.AffairMsg)
}

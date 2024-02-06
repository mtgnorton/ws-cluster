package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/core/client"
)

var DefaultHandler = NewWsHandler()

type Handle interface {
	Handle(ctx context.Context, client client.Client, message *wsmessage.Req)
}

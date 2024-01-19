package handler

import (
	"github.com/mtgnorton/ws-cluster/core/client"
)

var DefaultHandler = NewWsHandler()

type Handle interface {
	Handle(client client.Client, message []byte)
}

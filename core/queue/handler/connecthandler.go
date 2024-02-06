package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

type ConnectHandler struct {
	opts *Options
}

func (h *ConnectHandler) Handle(ctx context.Context, msg queuemessage.Message) (isAck bool) {
	manager, isAck := h.opts.manager, true

	servers := manager.ServersByPID(ctx, msg.PID)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, wsmessage.NewSuccessRes("user connect", "", msg))
	}
	return
}

func NewConnectHandler(opts ...Option) *ConnectHandler {
	return &ConnectHandler{
		opts: NewOptions(opts...),
	}
}

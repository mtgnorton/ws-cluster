package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
	"github.com/mtgnorton/ws-cluster/message/wsmessage"
)

type DisconnectHandler struct {
	opts *Options
}

func (h *DisconnectHandler) Handle(ctx context.Context, msg queuemessage.Message) (isAck bool) {
	manager, isAck := h.opts.manager, true

	servers := manager.ServersByPID(ctx, msg.PID)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, wsmessage.NewSuccessRes("user dis connect", "", msg))
	}
	return
}

func NewDisconnectHandler(opts ...Option) *DisconnectHandler {
	return &DisconnectHandler{
		opts: NewOptions(opts...),
	}
}

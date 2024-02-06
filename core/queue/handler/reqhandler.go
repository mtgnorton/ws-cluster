package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

type ReqHandler struct {
	opts *Options
}

func (h *ReqHandler) Handle(ctx context.Context, msg queuemessage.Message) (isAck bool) {
	manager, isAck := h.opts.manager, true

	// 根据pid找寻服务端
	servers := manager.ServersByPID(ctx, msg.PID)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, wsmessage.NewSuccessRes("user request", "", msg))
	}
	return
}

func NewReqHandler(opts ...Option) *ReqHandler {
	return &ReqHandler{
		opts: NewOptions(opts...),
	}
}

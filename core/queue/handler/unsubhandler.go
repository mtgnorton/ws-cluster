package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

type UnSubHandler struct {
	opts *Options
}

func (h *UnSubHandler) Handle(ctx context.Context, msg queuemessage.Message) (isAck bool) {
	logger, manager, isAck := h.opts.logger, h.opts.manager, true

	logger.Debugf(ctx, "queue unsub handler payload: %v", msg)
	// 根据pid找寻服务端
	servers := manager.ServersByPID(ctx, msg.PID)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, wsmessage.NewSuccessRes("", "", msg))
	}
	return
}

func NewUnSubHandler(opts ...Option) *SubHandler {
	return &SubHandler{
		opts: NewOptions(opts...),
	}
}

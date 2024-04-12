package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/clustermessage"
)

// UserHandler 从消息队列接收到用户端的消息，将其转发给服务端
type UserHandler struct {
	opts *Options
}

// SendToUserMessage 收窄发送到业务服务端消息的字段
type SendToServerMessage struct {
	AffairID string                `json:"affair_id,omitempty"` // 用户发送消息时，affair_id
	Payload  interface{}           `json:"payload,omitempty"`
	Type     clustermessage.Type   `json:"type,omitempty"`
	Source   clustermessage.Source `json:"source,omitempty"`
}

func (h *UserHandler) Handle(ctx context.Context, msg *clustermessage.AffairMsg) (isAck bool) {
	var (
		logger  = h.opts.logger
		manager = h.opts.manager
	)
	isAck = true

	servers := manager.ServersByPID(ctx, msg.Source.PID)
	logger.Debugf(ctx, "Send to server msg:%+v, servers:%+v", msg, servers)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, msg)
	}
	return
}

func NewUserHandler(opts ...Option) Handle {
	return &UserHandler{
		opts: NewOptions(opts...),
	}
}

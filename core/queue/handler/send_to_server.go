package handler

import (
	"context"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/shared"
)

// SendToServer 从消息队列接收到用户端的消息，将其转发给服务端
type SendToServer struct {
	opts *Options
}

// SendToServerMessage 收窄发送到业务服务端消息的字段
type SendToServerMessage struct {
	AffairID string                `json:"affair_id,omitempty"` // 用户发送消息时，affair_id
	Payload  interface{}           `json:"payload,omitempty"`
	Type     clustermessage.Type   `json:"type,omitempty"`
	Source   clustermessage.Source `json:"source,omitempty"`
}

func (h *SendToServer) Handle(ctx context.Context, msg *clustermessage.AffairMsg) (isAck bool) {
	var (
		logger  = h.opts.logger
		manager = h.opts.manager
	)
	isAck = true
	end := shared.TimeoutDetection.Do(time.Second*3, func() {
		logger.Errorf(ctx, "SendToServer  msg timeout,msg:%+v", msg)
	})
	defer end()

	if msg.Source == nil {
		return
	}
	servers := manager.ServersByPID(ctx, msg.Source.PID)
	logger.Debugf(ctx, "SendToServer msg:%+v, servers:%+v", msg, servers)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, msg)
	}
	return
}

func NewSendToServerHandler(opts ...Option) Handle {
	return &SendToServer{
		opts: NewOptions(opts...),
	}
}

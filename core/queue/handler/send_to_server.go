package handler

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/shared/kit"

	"github.com/mtgnorton/ws-cluster/clustermessage"
)

// SendToServer 从消息队列接收到用户端的消息，将其转发给服务端
type SendToServer struct {
	opts          *Options
	lastSlowLogAt atomic.Int64
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
	beginTime := time.Now()

	if msg.Source == nil {
		return
	}
	servers := manager.ServersByPID(ctx, msg.Source.PID)
	if len(servers) == 0 {
		return
	}
	for _, client := range servers {
		client.Send(ctx, msg)
	}
	costMs := float64(time.Since(beginTime).Microseconds()) / 1000.0
	if costMs >= 20 && kit.AllowByInterval(&h.lastSlowLogAt, 2*time.Second) {
		logger.Warnf(ctx, "QueueHandler SendToServer slow=%0.2fms,pid=%s,server_count=%d,type=%s,payload=%s", costMs, msg.Source.PID, len(servers), msg.Type, kit.LogSnippet(msg.Payload, 240))
	}
	return
}

func NewSendToServerHandler(opts ...Option) Handle {
	return &SendToServer{
		opts: NewOptions(opts...),
	}
}

package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"

	"github.com/mtgnorton/ws-cluster/clustermessage"
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
	beginTime := time.Now()

	if msg.Source == nil {
		return
	}
	servers := manager.ServersByPID(ctx, msg.Source.PID)
	if len(servers) == 0 {
		return
	}
	successCount := 0
	for _, client := range servers {
		if client.Send(ctx, msg) {
			successCount++
		}
	}
	cost := time.Since(beginTime)
	costMs := float64(cost.Microseconds()) / 1000.0
	forceTrace := costMs >= 20 || successCount != len(servers)
	nodeID := strconv.FormatInt(shared.GetNodeID(), 10)
	nodeIP := shared.GetInternalIP()
	clustermessage.AddTraceEventToMessage(msg, clustermessage.TraceEventWSDispatchToServerDone, nodeID, nodeIP, clustermessage.DurationMs(cost), forceTrace)
	if clustermessage.ShouldLogTrace(msg.Trace, forceTrace) {
		logger.Infof(ctx, clustermessage.BuildTraceLogForMessage(msg, nodeID, nodeIP, "ws_dispatch_to_server_done"))
	}
	return
}

func NewSendToServerHandler(opts ...Option) Handle {
	return &SendToServer{
		opts: NewOptions(opts...),
	}
}

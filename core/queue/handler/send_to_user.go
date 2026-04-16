package handler

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/shared/kit"
)

// SendToUser 从消息队列接收到业务服务端的消息，将其转发给用户端
type SendToUser struct {
	opts          *Options
	lastSlowLogAt atomic.Int64
}

// SendToUserMessage 收窄发送到用户端消息的字段
type SendToUserMessage struct {
	AffairID string      `json:"affair_id,omitempty"` // 用户发送消息时，affair_id
	Payload  interface{} `json:"payload,omitempty"`
}

func (h *SendToUser) Handle(ctx context.Context, msg *clustermessage.AffairMsg) (isAck bool) {
	logger, manager, isAck := h.opts.logger, h.opts.manager, true
	if msg.To == nil {
		logger.Warnf(ctx, "QueueHandler SendToUser msg.To is nil")
		return
	}
	pid, uids, cids := msg.To.PID, msg.To.UIDs, msg.To.CIDs
	beginTime := time.Now()

	if pid == "" {
		logger.Warnf(ctx, "QueueHandler SendToUser msg pid is empty,affair_id:%s", msg.AffairID)
		return
	}

	finalClients := make([]client.Client, 0, len(uids)+len(cids))
	if len(uids) == 0 && len(cids) == 0 {
		finalClients = manager.ClientsByPIDs(ctx, pid)
	} else {
		seen := make(map[string]struct{}, len(uids)+len(cids))
		appendUnique := func(clients []client.Client) {
			for _, currentClient := range clients {
				cid := currentClient.GetCID()
				if _, ok := seen[cid]; ok {
					continue
				}
				seen[cid] = struct{}{}
				finalClients = append(finalClients, currentClient)
			}
		}
		if len(uids) > 0 {
			appendUnique(manager.ClientsByUIDs(ctx, pid, uids...))
		}
		if len(cids) > 0 {
			appendUnique(manager.Clients(ctx, cids...))
		}
	}
	if len(finalClients) == 0 {
		return
	}

	sendMsg := SendToUserMessage{
		AffairID: msg.AffairID,
		Payload:  msg.Payload,
	}
	for _, client := range finalClients {
		client.Send(ctx, sendMsg)
	}

	costMs := float64(time.Since(beginTime).Microseconds()) / 1000.0
	if costMs >= 50 && kit.AllowByInterval(&h.lastSlowLogAt, 2*time.Second) {
		logger.Warnf(ctx, "QueueHandler SendToUser slow=%0.2fms,pid=%s,target=%d,uids=%d,cids=%d,payload=%s", costMs, pid, len(finalClients), len(uids), len(cids), kit.LogSnippet(msg.Payload, 240))
	}
	return
}

func NewSendToUserHandler(opts ...Option) Handle {
	return &SendToUser{
		opts: NewOptions(opts...),
	}
}

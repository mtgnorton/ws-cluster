package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/clustermessage"
)

// ServerHandler 从消息队列接收到业务服务端的消息，将其转发给用户端
type ServerHandler struct {
	opts *Options
}

// SendToUserMessage 收窄发送到用户端消息的字段
type SendToUserMessage struct {
	AffairID string      `json:"affair_id,omitempty"` // 用户发送消息时，affair_id
	Payload  interface{} `json:"payload,omitempty"`
}

func (h *ServerHandler) Handle(ctx context.Context, msg *clustermessage.AffairMsg) (isAck bool) {
	logger, manager, isAck := h.opts.logger, h.opts.manager, true
	pid, uids, cids := msg.To.PID, msg.To.UIDs, msg.To.CIDs
	if pid == "" {
		logger.Infof(ctx, "push msg pid is empty,msg:%+v", msg)
		return
	}

	finalClients := manager.ClientsByPIDs(ctx, pid)
	if len(finalClients) == 0 {
		logger.Debugf(ctx, "push msg pid %s not found,msg:%+v", pid, msg)
		return
	}
	if len(uids) > 0 {
		uClients := manager.ClientsByUIDs(ctx, pid, uids...)
		// 求交集
		finalClients = intersect(finalClients, uClients)
	}

	if len(cids) > 0 {
		clients := manager.Clients(ctx, cids...)
		finalClients = intersect(finalClients, clients)
	}

	if len(finalClients) == 0 {
		logger.Debugf(ctx, "push msg not found client,msg:%+v", msg)
		return
	}

	for _, client := range finalClients {
		client.Send(ctx, SendToUserMessage{
			AffairID: msg.AffairID,
			Payload:  msg.Payload,
		})
	}
	return
}

func NewServerHandler(opts ...Option) Handle {
	return &ServerHandler{
		opts: NewOptions(opts...),
	}
}

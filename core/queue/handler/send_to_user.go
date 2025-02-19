package handler

import (
	"context"
	"time"

	"ws-cluster/clustermessage"
	"ws-cluster/core/client"
	"ws-cluster/shared/kit"
)

// SendToUser 从消息队列接收到业务服务端的消息，将其转发给用户端
type SendToUser struct {
	opts *Options
}

// SendToUserMessage 收窄发送到用户端消息的字段
type SendToUserMessage struct {
	AffairID string      `json:"affair_id,omitempty"` // 用户发送消息时，affair_id
	Payload  interface{} `json:"payload,omitempty"`
}

func (h *SendToUser) Handle(ctx context.Context, msg *clustermessage.AffairMsg) (isAck bool) {
	var (
		source *clustermessage.Source
		to     *clustermessage.To
	)
	if msg.Source != nil {
		source = msg.Source
	}
	if msg.To != nil {
		to = msg.To
	}
	logger, manager, isAck := h.opts.logger, h.opts.manager, true
	pid, uids, cids := to.PID, to.UIDs, to.CIDs

	end := kit.DoWithTimeout(time.Second*5, func() {
		logger.Errorf(ctx, "QueueHandler SendToUser Handle msg timeout,msg:%+v", msg)
	})
	defer end()

	if pid == "" {
		logger.Infof(ctx, "QueueHandler SendToUser msg pid is empty,msg:%+v,from:%+v,to:%+v", msg.Payload, source, to)
		return
	}

	var finalClients []client.Client

	if len(uids) == 0 && len(cids) == 0 {
		finalClients = manager.ClientsByPIDs(ctx, pid)
	} else {
		if len(uids) > 0 {
			uClients := manager.ClientsByUIDs(ctx, pid, uids...)
			finalClients = kit.SliceUnion(finalClients, uClients)
		}
		if len(cids) > 0 {
			clients := manager.Clients(ctx, cids...)
			finalClients = kit.SliceUnion(finalClients, clients)
		}
	}
	if len(finalClients) == 0 {
		logger.Debugf(ctx, "QueueHandler SendToUser msg not found client,msg:%+v,from:%+v,to:%+v", msg.Payload, source, to)
		return
	}
	logger.Debugf(ctx, "QueueHandler SendToUser msg to clients:%+v,msg:%+v,from:%+v,to:%+v", finalClients, msg.Payload, source, to)
	for _, client := range finalClients {
		client.Send(ctx, SendToUserMessage{
			AffairID: msg.AffairID,
			Payload:  msg.Payload,
		})
	}
	return
}

func NewSendToUserHandler(opts ...Option) Handle {
	return &SendToUser{
		opts: NewOptions(opts...),
	}
}

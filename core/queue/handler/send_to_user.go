package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/shared"
)

// SendToUser 从消息队列接收到业务服务端的消息，将其转发给用户端
type SendToUser struct {
	opts *Options
}

// SendToUserMessage 收窄发送到用户端消息的字段
type SendToUserMessage struct {
	AffairID string                `json:"affair_id,omitempty"` // 用户发送消息时，affair_id
	Payload  interface{}           `json:"payload,omitempty"`
	Trace    *clustermessage.Trace `json:"-"`
	PID      string                `json:"-"`
}

func (m SendToUserMessage) MessageTrace() *clustermessage.Trace {
	return m.Trace
}

func (m SendToUserMessage) MessageAffairID() string {
	return m.AffairID
}

func (m SendToUserMessage) MessageType() clustermessage.Type {
	return clustermessage.TypePush
}

func (m SendToUserMessage) MessagePID() string {
	return m.PID
}

func (m SendToUserMessage) MessagePayload() interface{} {
	return m.Payload
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
		Trace:    msg.Trace,
		PID:      pid,
	}
	successCount := 0
	for _, client := range finalClients {
		if client.Send(ctx, sendMsg) {
			successCount++
		}
	}

	cost := time.Since(beginTime)
	costMs := float64(cost.Microseconds()) / 1000.0
	forceTrace := costMs >= 50 || successCount != len(finalClients)
	nodeID := strconv.FormatInt(shared.GetNodeID(), 10)
	nodeIP := shared.GetInternalIP()
	clustermessage.AddTraceEventToMessage(msg, clustermessage.TraceEventWSDispatchToUserDone, nodeID, nodeIP, clustermessage.DurationMs(cost), forceTrace)
	if clustermessage.ShouldLogTrace(msg.Trace, forceTrace) {
		logger.Infof(ctx, clustermessage.BuildTraceLogForMessage(msg, nodeID, nodeIP, "ws_dispatch_to_user_done"))
	}
	return
}

func NewSendToUserHandler(opts ...Option) Handle {
	return &SendToUser{
		opts: NewOptions(opts...),
	}
}

package handler

import (
	"context"
	"strings"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

type PushHandler struct {
	opts *Options
}

func (h *PushHandler) Handle(ctx context.Context, msg queuemessage.Message) (isAck bool) {
	logger, manager, isAck := h.opts.logger, h.opts.manager, true

	payload, err := queuemessage.ParsePush(msg.Payload)
	if err != nil {
		logger.Infof(ctx, "parse payload  error: %v", err)
		return
	}
	if payload.PID == "" {
		logger.Infof(ctx, "push msg pid is empty,msg:%+v", payload)
		return
	}

	finalClients := manager.ClientsByPIDs(ctx, payload.PID)
	if len(finalClients) == 0 {
		logger.Debugf(ctx, "push msg pid %s not found,msg:%+v", payload.PID, payload)
		return
	}
	if payload.UIDs != "" {
		uClients := manager.ClientsByUIDs(ctx, strings.Split(payload.UIDs, ",")...)
		// 求交集
		finalClients = intersect(finalClients, uClients)
	}

	if payload.CIDs != "" {
		clients := manager.Clients(ctx, strings.Split(payload.CIDs, ",")...)
		finalClients = intersect(finalClients, clients)
	}

	if payload.Tags != "" {
		clients := manager.ClientByTags(ctx, strings.Split(payload.Tags, ",")...)
		finalClients = intersect(finalClients, clients)
	}
	if len(finalClients) == 0 {
		logger.Debugf(ctx, "push msg not found client,msg:%+v", payload)
		return
	}
	for _, client := range finalClients {
		client.Send(ctx, wsmessage.NewSuccessRes("", msg.Identification, payload.Data))
	}

	return
}

func NewPushHandler(opts ...Option) Handle {
	return &PushHandler{
		opts: NewOptions(opts...),
	}
}

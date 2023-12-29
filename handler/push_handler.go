package handler

import (
	"encoding/json"
	"github.com/mtgnorton/ws-cluster/message"
	"strings"
)

type PushReq struct {
	PID       string      `json:"pid"`
	UIDs      string      `json:"uids"`
	ClientIDs string      `json:"client_ids"`
	Tags      string      `json:"tags"`
	Data      interface{} `json:"data"`
}

type PushHandler struct {
	opts *Options
}

func (p *PushHandler) Type() message.Type {

	return message.TypePush
}

func (p *PushHandler) Handle(message *message.ReqMessage) (isAck bool, err error) {
	isAck = true
	logger, manager := p.opts.logger, p.opts.manager
	req, err := p.parse(message.Payload)
	if err != nil {
		logger.Infof("parse push message error: %v", err)
		return
	}
	if req.PID == "" {
		logger.Infof("push message pid is empty")
		return
	}

	finalClients := manager.GetByPIDs(req.PID)
	if len(finalClients) == 0 {
		logger.Debugf("push message pid %s not found,msg:%+v", req.PID, message)
		return
	}
	if req.UIDs != "" {
		uClients := manager.GetByUIDs(strings.Split(req.UIDs, ",")...)
		// 求交集
		finalClients = intersect(finalClients, uClients)
	}

	if req.ClientIDs != "" {
		clients := manager.Gets(strings.Split(req.ClientIDs, ",")...)
		finalClients = intersect(finalClients, clients)
	}

	if req.Tags != "" {
		clients := manager.GetByTags(strings.Split(req.Tags, ",")...)
		finalClients = intersect(finalClients, clients)
	}
	if len(finalClients) == 0 {
		logger.Debugf("push message not found client,msg:%+v", message)
		return
	}

	for _, client := range finalClients {
		client.Send(req.Data)
	}

	return
}

func (p *PushHandler) parse(payload interface{}) (req *PushReq, err error) {
	payloadBytes, ok := payload.([]byte)
	if !ok {
		return nil, ErrInvalidPayload
	}
	req = &PushReq{}
	err = json.Unmarshal(payloadBytes, req)
	return
}

func NewPushHandler(opts ...Option) *PushHandler {
	return &PushHandler{
		opts: newOptions(opts...),
	}
}

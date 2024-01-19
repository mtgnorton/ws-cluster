package handler

import (
	"encoding/json"
	"github.com/mtgnorton/ws-cluster/http/message"
	"strings"
)

type PushHandler struct {
	opts *Options
}

type PushPayLoad struct {
	PID  string      `json:"pid"`
	UIDs string      `json:"uids"`
	CIDs string      `json:"cids"`
	Tags string      `json:"tags"`
	Data interface{} `json:"data"`
}

func (p *PushHandler) Handle(messageBytes []byte) (isAck bool) {
	logger, manager := p.opts.logger, p.opts.manager
	isAck = true

	msg, err := message.Parse(messageBytes)
	if err != nil {
		logger.Infof("parse msg error: %v", err)
		return
	}
	req, err := p.parse(msg.Payload)
	if err != nil {
		logger.Infof("parse payload  error: %v", err)
		return
	}
	if req.PID == "" {
		logger.Infof("push msg pid is empty,msg:%+v", msg)
		return
	}

	finalClients := manager.GetByPIDs(req.PID)
	if len(finalClients) == 0 {
		logger.Debugf("push msg pid %s not found,msg:%+v", req.PID, msg)
		return
	}
	if req.UIDs != "" {
		uClients := manager.GetByUIDs(strings.Split(req.UIDs, ",")...)
		// 求交集
		finalClients = intersect(finalClients, uClients)
	}

	if req.CIDs != "" {
		clients := manager.Gets(strings.Split(req.CIDs, ",")...)
		finalClients = intersect(finalClients, clients)
	}

	if req.Tags != "" {
		clients := manager.GetByTags(strings.Split(req.Tags, ",")...)
		finalClients = intersect(finalClients, clients)
	}
	if len(finalClients) == 0 {
		logger.Debugf("push msg not found client,msg:%+v", msg)
		return
	}

	for _, client := range finalClients {
		client.Send(req.Data)
	}

	return
}

func (p *PushHandler) parse(payload interface{}) (pushReq *PushPayLoad, err error) {
	pushReq = &PushPayLoad{}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	err = json.Unmarshal(payloadBytes, pushReq)
	return
}

func NewPushHandler(opts ...Option) Handle {
	return &PushHandler{
		opts: NewOptions(opts...),
	}
}

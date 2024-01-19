package handler

import (
	"encoding/json"
	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/ws/message"
	"strings"
)

type WsHandler struct {
	opts *Options
}

type SubscribePayload struct {
	Tags string `json:"tags"` // 多个tag用逗号分隔
}

func NewWsHandler(opts ...Option) *WsHandler {
	options := NewOptions(opts...)
	return &WsHandler{
		opts: options,
	}
}
func (w *WsHandler) Handle(client client.Client, messageBytes []byte) {
	w.opts.logger.Debugf("WsHandler-handle messageBytes %s", messageBytes)
	msg, err := w.opts.parser.Parse(messageBytes)
	if err != nil {
		client.Send(message.NewErrorRes("parse messageBytes error", ""))
		w.opts.logger.Infof("WsHandler-handle parse messageBytes error %v", err)
		return
	}
	switch msg.Type {
	case message.TypeSubscribe:
		w.subscribe(client, msg)
	case message.TypeUnsubscribe:
		w.unsubscribe(client, msg)
	}
}

func (w *WsHandler) subscribe(client client.Client, msg *message.Req) {
	w.opts.logger.Debugf("WsHandler-subscribe subscribe message %v", msg)
	payload, err := w.parse(msg.Payload)
	if err != nil {
		client.Send(message.NewErrorRes("parse payload error", ""))
		w.opts.logger.Infof("WsHandler-subscribe parse payload error %v", err)
		return
	}
	if payload.Tags == "" {
		client.Send(message.NewErrorRes("tags is empty", ""))
		return
	}
	w.opts.manager.BindTag(client, strings.Split(payload.Tags, ",")...)
	client.Send(message.NewSuccessRes("success", msg.RequestID))

}

func (w *WsHandler) unsubscribe(client client.Client, msg *message.Req) {
	w.opts.logger.Debugf("WsHandler-unsubscribe message %v", msg)
	payload, err := w.parse(msg.Payload)
	if err != nil {
		client.Send(message.NewErrorRes("parse payload error", ""))
		w.opts.logger.Infof("WsHandler-unsubscribe parse payload error %v", err)
		return
	}
	if payload.Tags == "" {
		client.Send(message.NewErrorRes("tags is empty", ""))
		return
	}
	w.opts.manager.UnbindTag(client, strings.Split(payload.Tags, ",")...)
	client.Send(message.NewSuccessRes("success", msg.RequestID))
}

func (w *WsHandler) parse(payload interface{}) (*SubscribePayload, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req := &SubscribePayload{}
	err = json.Unmarshal(bytes, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

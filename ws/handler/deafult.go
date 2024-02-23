package handler

import (
	"context"
	"strings"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"

	"github.com/mtgnorton/ws-cluster/core/client"
)

type WsHandler struct {
	opts *Options
}

type RequestPayload struct {
	Content string `json:"content"` // 请求标识
}

func NewWsHandler(opts ...Option) *WsHandler {
	options := NewOptions(opts...)
	return &WsHandler{
		opts: options,
	}
}
func (w *WsHandler) Handle(ctx context.Context, c client.Client, msg *wsmessage.Req) {
	w.opts.logger.Debugf(ctx, "WsHandler-handle msg %s", msg)
	// 管理端: 所有消息类型
	// 服务端: 推送
	// 用户端: 订阅、取消订阅、请求
	// 用户端的连接,断开事件需要通知服务端
	// 只处理客户端的连接,断开事件

	switch msg.Type {
	case queuemessage.TypeSubscribe:
		if c.Type() == client.CTypeServer {
			c.Send(ctx, wsmessage.NewErrorRes("permission deny", ""))
		}
		w.subscribe(ctx, c, msg)
	case queuemessage.TypeUnsubscribe:
		if c.Type() == client.CTypeServer {
			c.Send(ctx, wsmessage.NewErrorRes("permission deny", ""))
		}
		w.unsubscribe(ctx, c, msg)
	case queuemessage.TypePush:
		if c.Type() == client.CTypeUser {
			c.Send(ctx, wsmessage.NewErrorRes("permission deny", ""))
		}
		w.Push(ctx, c, msg)
	case queuemessage.TypeRequest:
		if c.Type() == client.CTypeServer {
			c.Send(ctx, wsmessage.NewErrorRes("permission deny", ""))
		}
		w.Request(ctx, c, msg)
	case queuemessage.TypeConnect:
		if c.Type() != client.CTypeUser {
			return
		}
		w.Connect(ctx, c, msg)
	case queuemessage.TypeDisconnect:
		if c.Type() != client.CTypeUser {
			return
		}
		w.Disconnect(ctx, c, msg)

	}
}

func (w *WsHandler) subscribe(ctx context.Context, c client.Client, msg *wsmessage.Req) {
	w.opts.logger.Debugf(ctx, "WsHandler-subscribe  message %v", msg)
	payload, err := queuemessage.ParseSubscribe(msg.Payload)
	if err != nil {
		c.Send(ctx, wsmessage.NewErrorRes("parse payload error", ""))
		w.opts.logger.Infof(ctx, "WsHandler-subscribe parse payload error %v", err)
		return
	}
	if payload.Tags == "" {
		c.Send(ctx, wsmessage.NewErrorRes("tags is empty", ""))
		return
	}
	w.opts.manager.BindTag(ctx, c, strings.Split(payload.Tags, ",")...)

	_, _, pid := c.GetIDs()
	queueMsg := &queuemessage.Message{
		Type:           queuemessage.TypeSubscribe,
		PID:            pid,
		Identification: msg.Identification,
		Payload:        payload,
	}
	err = w.opts.queue.Publish(ctx, queueMsg)
	if err != nil {
		c.Send(ctx, wsmessage.NewErrorRes("publish error", ""))
		w.opts.logger.Infof(ctx, "WsHandler-unsubscribe publish error %v", err)
		return
	}
	c.Send(ctx, wsmessage.NewSuccessRes("success", msg.Identification))
}

func (w *WsHandler) unsubscribe(ctx context.Context, c client.Client, msg *wsmessage.Req) {
	w.opts.logger.Debugf(ctx, "WsHandler-unsubscribe message %v", msg)
	payload, err := queuemessage.ParseUnsubscribe(msg.Payload)
	if err != nil {
		c.Send(ctx, wsmessage.NewErrorRes("parse payload error", ""))
		w.opts.logger.Infof(ctx, "WsHandler-unsubscribe parse payload error %v", err)
		return
	}
	if payload.Tags == "" {
		c.Send(ctx, wsmessage.NewErrorRes("tags is empty", ""))
		return
	}
	w.opts.manager.UnbindTag(ctx, c, strings.Split(payload.Tags, ",")...)
	_, _, pid := c.GetIDs()
	queueMsg := &queuemessage.Message{
		Type:           queuemessage.TypeUnsubscribe,
		PID:            pid,
		Identification: msg.Identification,
		Payload:        payload,
	}
	err = w.opts.queue.Publish(ctx, queueMsg)
	if err != nil {
		c.Send(ctx, wsmessage.NewErrorRes("publish error", ""))
		w.opts.logger.Infof(ctx, "WsHandler-unsubscribe publish error %v", err)
		return
	}
	c.Send(ctx, wsmessage.NewSuccessRes("success", msg.Identification))
}

// Push 推送消息,只能由服务端调用
func (w *WsHandler) Push(ctx context.Context, c client.Client, msg *wsmessage.Req) {

	queueMsg := &queuemessage.Message{
		Type:           queuemessage.TypePush,
		Identification: msg.Identification,
		Payload:        msg.Payload,
	}

	err := w.opts.queue.Publish(ctx, queueMsg)
	if err != nil {
		w.opts.logger.Infof(ctx, "WsHandler-push publish error %v", err)
		return
	}
	c.Send(ctx, wsmessage.NewSuccessRes("success", msg.Identification))
}

// Request 请求消息,只能由客户端调用
func (w *WsHandler) Request(ctx context.Context, c client.Client, msg *wsmessage.Req) {

	_, _, pid := c.GetIDs()

	queueMsg := &queuemessage.Message{
		Type:           queuemessage.TypeRequest,
		PID:            pid,
		Identification: msg.Identification,
		Payload:        msg.Payload,
	}

	err := w.opts.queue.Publish(ctx, queueMsg)
	if err != nil {
		w.opts.logger.Infof(ctx, "WsHandler-request publish error %v", err)
		return
	}
	c.Send(ctx, wsmessage.NewSuccessRes("success", msg.Identification))
}

func (w *WsHandler) Connect(ctx context.Context, c client.Client, msg *wsmessage.Req) {
	w.opts.logger.Debugf(ctx, "WsHandler-connect message %v", msg)
	id, uid, Pid := c.GetIDs()
	queueMsg := &queuemessage.Message{
		Type:    queuemessage.TypeConnect,
		PID:     Pid,
		Payload: &queuemessage.PayloadConnect{UID: uid, CID: id},
	}
	err := w.opts.queue.Publish(ctx, queueMsg)
	if err != nil {
		w.opts.logger.Infof(ctx, "WsHandler-request publish error %v", err)
		return
	}
}

func (w *WsHandler) Disconnect(ctx context.Context, c client.Client, msg *wsmessage.Req) {
	w.opts.logger.Debugf(ctx, "WsHandler-disconnect message %v", msg)
	id, uid, Pid := c.GetIDs()
	queueMsg := &queuemessage.Message{
		Type:    queuemessage.TypeDisconnect,
		PID:     Pid,
		Payload: &queuemessage.PayloadConnect{UID: uid, CID: id},
	}
	err := w.opts.queue.Publish(ctx, queueMsg)
	if err != nil {
		w.opts.logger.Infof(ctx, "WsHandler-request publish error %v", err)
		return
	}
}

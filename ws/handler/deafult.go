package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"

	"github.com/mtgnorton/ws-cluster/core/client"
)

type WsHandler struct {
	opts *Options
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
	// 用户端: 请求
	// 用户端的连接,断开事件需要通知服务端
	// 只处理客户端的连接,断开事件

	switch msg.Type {
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

// Push 推送消息,只能由服务端调用
func (w *WsHandler) Push(ctx context.Context, c client.Client, msg *wsmessage.Req) {

	_, _, pid := c.GetIDs()
	queueMsg := &queuemessage.Message{
		Type:           queuemessage.TypePush,
		Identification: msg.Identification,
		Payload:        msg.Payload,
		PID:            pid,
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
	id, uid, pid := c.GetIDs()
	queueMsg := &queuemessage.Message{
		Type:           queuemessage.TypeRequest,
		PID:            pid,
		Identification: msg.Identification,
		Payload: queuemessage.PayloadRequest{
			UID:     uid,
			CID:     id,
			Payload: msg.Payload,
		},
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

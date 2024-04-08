package handler

import (
	"context"

	"github.com/mtgnorton/ws-cluster/clustermessage"
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
func (w *WsHandler) Handle(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {
	w.opts.logger.Debugf(ctx, "WsHandler-handle msg %+v", msg)
	// 管理端: 所有消息类型
	// 服务端: 推送
	// 用户端: 请求
	// 用户端的连接,断开事件需要通知服务端, 只处理客户端的连接,断开事件
	// 判断是否是心跳消息
	if msg.Type == clustermessage.TypeHeart {
		c.Send(ctx, clustermessage.NewHeartResp(msg))
		return
	}

	if msg.Type == clustermessage.TypeConnect || msg.Type == clustermessage.TypeDisconnect {
		if c.Type() != client.CTypeUser {
			return
		}
		w.User(ctx, c, msg)
		return
	}

	// 用户端或者业务端主动发送的消息
	switch c.Type() {
	case client.CTypeUser:
		msg.Type = clustermessage.TypeRequest
		w.User(ctx, c, msg)
	case client.CTypeServer:
		msg.Type = clustermessage.TypePush
		w.Server(ctx, c, msg)
	}

}

// Server 来自Server端消息封装
func (w *WsHandler) Server(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {

	var (
		logger = w.opts.logger
		queue  = w.opts.queue
	)
	// 如果没有传递到的用户，直接返回
	if len(msg.To.CIDs) == 0 && len(msg.To.UIDs) == 0 {
		logger.Infof(ctx, "WsHandler-Server msg.To is empty")
		return
	}
	_, _, msg.To.PID = c.GetIDs()

	err := queue.Publish(ctx, msg)
	if err != nil {
		logger.Infof(ctx, "WsHandler-push server publish error %v", err)
		return
	}
	logger.Debugf(ctx, "WsHandler-push server publish success,msg:%v", msg)
	if msg.AckID != "" {
		c.Send(ctx, clustermessage.NewAck(msg.AckID))
	}

}

// User 来自用户端消息封装
func (w *WsHandler) User(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {
	cid, uid, pid := c.GetIDs()
	msg.Source = &clustermessage.Source{
		PID: pid,
		UID: uid,
		CID: cid,
	}
	err := w.opts.queue.Publish(ctx, msg)
	if err != nil {
		w.opts.logger.Infof(ctx, "WsHandler-request user publish error %v", err)
		return
	}
	w.opts.logger.Debugf(ctx, "WsHandler-request user publish success,msg:%v", msg)
	if msg.AckID != "" {
		c.Send(ctx, clustermessage.NewAck(msg.AckID))
	}
}

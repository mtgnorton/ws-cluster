package handler

import (
	"context"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
)

type WsHandler struct {
	opts *Options
}

func NewWsHandler(opts ...Option) *WsHandler {
	options := NewOptions(opts...)
	w := &WsHandler{
		opts: options,
	}
	go w.sendClientsLoop()
	return w
}

// sendClientsLoop 定时推送用户端的连接信息
func (w *WsHandler) sendClientsLoop() {
	var (
		ctx    = w.opts.ctx
		logger = w.opts.logger
	)
	for range time.Tick(2 * time.Second) {
		// 获取所有的用户端的连接信息
		// 遍历所有的服务端
		// 发送给服务端
		for _, projectServerClients := range w.opts.manager.Projects(ctx) {
			cids := make([]string, 0)
			for _, c := range projectServerClients.Clients {
				cid, _, _ := c.GetIDs()
				cids = append(cids, cid)
			}
			if len(cids) == 0 {
				continue
			}
			msg := clustermessage.AffairMsg{
				AffairID: "",
				AckID:    "",
				Payload:  cids,
				Type:     clustermessage.TypeOnlineClients,
				Source: &clustermessage.Source{
					PID: projectServerClients.PID,
					UID: "",
					CID: "",
				},
				To: nil,
			}
			err := w.opts.queue.Publish(ctx, &msg)
			if err != nil {
				logger.Infof(ctx, "WsHandler-sendClientsLoop publish error %v", err)
			}

		}
	}
}
func (w *WsHandler) Handle(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {
	//w.opts.logger.Debugf(ctx, "Receive msg %+v", msg)
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
		w.handleMsgFromUser(ctx, c, msg)
		return
	}

	// 用户端或者业务端主动发送的消息
	switch c.Type() {
	case client.CTypeUser:
		msg.Type = clustermessage.TypeRequest
		w.handleMsgFromUser(ctx, c, msg)
	case client.CTypeServer:
		msg.Type = clustermessage.TypePush
		w.handleMsgFromServer(ctx, c, msg)
	}

}

// handleMsgFromServer 来自Server端消息封装
func (w *WsHandler) handleMsgFromServer(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {

	var (
		logger = w.opts.logger
		queue  = w.opts.queue
	)
	// 如果没有传递到的用户，直接返回
	if len(msg.To.CIDs) == 0 && len(msg.To.UIDs) == 0 {
		logger.Infof(ctx, "WsHandler-handleMsgFromServer msg.To is empty")
		return
	}
	_, _, msg.To.PID = c.GetIDs()

	err := queue.Publish(ctx, msg)
	if err != nil {
		logger.Infof(ctx, "WsHandler-handleMsgFromServerserver publish error %v", err)
		return
	}
	logger.Debugf(ctx, "WsHandler-handleMsgFromServer  msg  success,msg:%+v,To:%+v", msg, msg.To)
	if msg.AckID != "" {
		c.Send(ctx, clustermessage.NewAck(msg.AckID))
	}

}

// handleMsgFromUser 来自用户端消息封装
func (w *WsHandler) handleMsgFromUser(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {
	cid, uid, pid := c.GetIDs()
	msg.Source = &clustermessage.Source{
		PID: pid,
		UID: uid,
		CID: cid,
	}
	err := w.opts.queue.Publish(ctx, msg)
	if err != nil {
		w.opts.logger.Infof(ctx, "WsHandler-handleMsgFromUser user publish error %v", err)
		return
	}
	w.opts.logger.Debugf(ctx, "WsHandler-handleMsgFromUser  msg  success,msg:%v", msg)
	if msg.AckID != "" {
		c.Send(ctx, clustermessage.NewAck(msg.AckID))
	}
}

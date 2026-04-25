package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/shared/kit"
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

type OnlineClient struct {
	CID string `json:"cid"`
	UID string `json:"uid"`
}

// sendClientsLoop 定时推送用户端的连接信息
func (w *WsHandler) sendClientsLoop() {
	var (
		ctx    = w.opts.ctx
		logger = w.opts.logger
	)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// 获取所有的用户端的连接信息
		// 遍历所有的服务端
		// 发送给服务端
		for _, projectServerClients := range w.opts.manager.Projects(ctx) {
			onlineClients := make([]*OnlineClient, 0)
			for _, c := range projectServerClients.Clients {
				cid, uid, _ := c.GetIDs()
				onlineClients = append(onlineClients, &OnlineClient{
					CID: cid,
					UID: uid,
				})
			}
			if len(onlineClients) == 0 {
				continue
			}

			kit.ChunkSlice(onlineClients, 1000, func(chunkOnlineClients []*OnlineClient) {
				msg := clustermessage.AffairMsg{
					AffairID: "",
					AckID:    "",
					Payload:  chunkOnlineClients,
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
					logger.Warnf(ctx, "WsHandler-sendClientsLoop publish error %v", err)
				}
			})

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
		c.Send(ctx, clustermessage.NewHeartResp(msg, c.GetCID()))
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
		nodeID = strconv.FormatInt(shared.GetNodeID(), 10)
		nodeIP = shared.GetInternalIP()
	)
	if msg.To == nil {
		logger.Warnf(ctx, "WsHandler-FromServer msg.To is nil")
		return
	}
	_, _, msg.To.PID = c.GetIDs()
	missingTrace := msg.Trace == nil || msg.Trace.ID == ""
	clustermessage.EnsureTrace(msg, "")
	if missingTrace {
		clustermessage.AddTraceEventToMessage(msg, clustermessage.TraceEventTraceMissingFromBiz, nodeID, nodeIP, 0, false)
	}
	clustermessage.AddTraceEventToMessage(msg, clustermessage.TraceEventWSRecvFromServer, nodeID, nodeIP, 0, false)

	err := queue.Publish(ctx, msg)
	if err != nil {
		logger.Warnf(ctx, "WsHandler-FromServer publish error %v", err)
		if clustermessage.ShouldLogTrace(msg.Trace, true) {
			logger.Warnf(ctx, clustermessage.BuildTraceLogForMessage(msg, nodeID, nodeIP, "ws_publish_failed"))
		}
		return
	}
	if msg.AckID != "" {
		c.Send(ctx, clustermessage.NewAck(msg.AckID))
	}

}

// handleMsgFromUser 来自用户端消息封装
func (w *WsHandler) handleMsgFromUser(ctx context.Context, c client.Client, msg *clustermessage.AffairMsg) {
	nodeID := strconv.FormatInt(shared.GetNodeID(), 10)
	nodeIP := shared.GetInternalIP()
	cid, uid, pid := c.GetIDs()
	msg.Source = &clustermessage.Source{
		PID: pid,
		UID: uid,
		CID: cid,
	}
	clustermessage.EnsureTrace(msg, "")
	clustermessage.AddTraceEventToMessage(msg, clustermessage.TraceEventWSRecvFromUser, nodeID, nodeIP, 0, false)
	err := w.opts.queue.Publish(ctx, msg)
	if err != nil {
		w.opts.logger.Warnf(ctx, "WsHandler-FromUser user publish error %v", err)
		if clustermessage.ShouldLogTrace(msg.Trace, true) {
			w.opts.logger.Warnf(ctx, clustermessage.BuildTraceLogForMessage(msg, nodeID, nodeIP, "ws_publish_failed"))
		}
		return
	}
	if msg.AckID != "" {
		c.Send(ctx, clustermessage.NewAck(msg.AckID))
	}
}

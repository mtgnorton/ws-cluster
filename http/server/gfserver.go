package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gutil"

	"ws-cluster/shared"
	"ws-cluster/shared/auth"

	"ws-cluster/clustermessage"
	"ws-cluster/core/checking"
	"ws-cluster/core/client"

	"ws-cluster/tools/wsprometheus"
	"ws-cluster/tools/wssentry"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type gfServer struct {
	opts   Options
	server *ghttp.Server
	sentry *wssentry.Handler
}

func New(opts ...Option) Server {
	s := &gfServer{
		opts:   NewOptions(opts...),
		server: g.Server("http"),
		sentry: wssentry.GfSentry,
	}

	return s
}
func (g gfServer) Name() string {
	return "gf"
}

func (g gfServer) Init(option ...Option) {
	for _, o := range option {
		o(&g.opts)
	}
}

func (g gfServer) Options() Options {
	return g.opts
}

func (g gfServer) Run() {
	g.server.Group("/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(g.sentry.MiddleWare)
		group.POST("/push", func(r *ghttp.Request) {
			ctx := r.Context()
			beginTime := time.Now()

			g.sentry.RecoverHttp(r, g.handler)

			// prometheus add metrics
			p := g.opts.prometheus

			err := p.GetAdd(wsprometheus.MetricRequestTotal, nil, 1)
			if err != nil {
				g.opts.logger.Infof(ctx, "add metric error:%s", err.Error())
			}
			err = p.GetAdd(wsprometheus.MetricRequestURLTotal, []string{"http", strconv.Itoa(r.Response.Status)}, 1)
			if err != nil {
				g.opts.logger.Infof(ctx, "add metric error:%s", err.Error())
			}
			err = p.GetObserve(wsprometheus.MetricRequestDuration, []string{"http"}, time.Since(beginTime).Seconds())
			if err != nil {
				g.opts.logger.Infof(ctx, "add metric error:%s", err.Error())
			}
		})

		group.GET("/reset_metrics", func(r *ghttp.Request) {
			nodeID := shared.GetNodeID()
			serverIP := shared.GetInternalIP()
			p := g.opts.prometheus.Get(wsprometheus.MetricWsConnection)
			p.Reset([]string{fmt.Sprintf("%d", nodeID), serverIP})

			r.Response.WriteJson(clustermessage.NewSuccessResp(fmt.Sprintf("%d,%s reset success", nodeID, serverIP)))
		})
	})

	g.opts.logger.Infof(context.Background(), "http server run on port:%d", g.opts.port)
	g.server.SetPort(g.opts.port)
	g.server.Run()
}

func (g gfServer) Stop() error {
	return g.server.Shutdown()
}

// 推送消息
//
//	@Summary		业务系统通过该接口推送消息
//	@Description	业务系统通过该接口推送消息,当同时传递了uids和cids时，会求并集
//	@ID				push-message
//	@Accept			json
//	@Produce		json
//	@Param			uids	query		string		false	"用户id，多个用户id以逗号隔开"
//	@Param			cids	query		string		false	"客户端id,多个客户端id以逗号隔开"
//	@Param			token	query		string		true	"签名"
//	@Param			data	query		string		true	"推送的消息内容,建议为json"
//	@Success		200		{string}	string		"{"code":1,"msg":"success","payload":{}}"
//	@Failure		200		{object}	message.Res	"code=0,msg=error"
//	@Router			/push [post]
func (g gfServer) handler(r *ghttp.Request) {
	var (
		token  = r.Get("token").String()
		uidStr = r.Get("uids").String()
		cidStr = r.Get("cids").String()
		data   = r.Get("data").String()
		uids   []string
		cids   []string
	)
	userData, err := auth.Decode(token)
	if err != nil {
		r.Response.WriteJson(clustermessage.NewErrorResp("token error"))
		return
	}
	if !checking.DefaultChecking.IsExist(userData.PID) {
		r.Response.WriteJson(clustermessage.NewErrorResp("PID denied"))
		return
	}
	if userData.ClientType == int(client.CTypeUser) {
		r.Response.WriteJson(clustermessage.NewErrorResp("permission denied"))
		return
	}
	for _, uid := range strings.Split(uidStr, ",") {
		uid = strings.TrimSpace(uid)
		if len(uid) > 0 {
			uids = append(uids, uid)
		}
	}
	for _, cid := range strings.Split(cidStr, ",") {
		cid = strings.TrimSpace(cid)
		if len(cid) > 0 {
			cids = append(cids, cid)
		}
	}
	if len(uids) == 0 && len(cids) == 0 {
		r.Response.WriteJson(clustermessage.NewErrorResp("uids or cids is required"))
		return
	}

	msg := &clustermessage.AffairMsg{
		Payload: data,
		Type:    clustermessage.TypePush,
		To:      &clustermessage.To{PID: userData.PID, UIDs: uids, CIDs: cids},
	}
	gutil.Dump(msg)

	// 随机休眠0-10s
	// time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	msg.Type = clustermessage.TypePush

	err = g.opts.queue.Publish(r.Context(), msg)
	if err != nil {
		g.opts.logger.Warnf(r.Context(), "publish message error:%s", err.Error())
		r.Response.WriteJson(clustermessage.NewErrorResp("publish message error"))
		return
	}
	r.Response.WriteJson(clustermessage.NewSuccessResp())
}

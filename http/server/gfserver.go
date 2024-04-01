package server

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"

	"github.com/mtgnorton/ws-cluster/shared"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"
	"github.com/mtgnorton/ws-cluster/tools/wssentry"
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
//	@Description	业务系统通过该接口推送消息
//	@ID				push-message
//	@Accept			json
//	@Produce		json
//	@Param			pid		query		string		true	"项目id"
//	@Param			uids	query		string		false	"用户id，多个用户id以逗号隔开"
//	@Param			cids	query		string		false	"客户端id,多个客户端id以逗号隔开"
//	@Param			tags	query		string		false	"标签,多个标签以逗号隔开"
//	@Param			sign	query		string		true	"签名"
//	@Param			data	query		string		true	"推送的消息内容"
//	@Success		200		{string}	string		"{"code":1,"msg":"success","payload":{}}"
//	@Failure		200		{object}	message.Res	"code=0,msg=error"
//	@Router			/push [post]
func (g gfServer) handler(r *ghttp.Request) {

	claims, err := shared.DefaultJwtWs.Parse(r.Get("token").String())
	if err != nil {
		r.Response.WriteJson(clustermessage.NewErrorResp("token error"))
		return
	}
	if claims.ClientType == int(client.CTypeUser) {
		r.Response.WriteJson(clustermessage.NewErrorResp("permission denied"))
		return
	}
	msg := &clustermessage.AffairMsg{}
	err = json.Unmarshal(r.GetBody(), &msg)
	if err != nil {
		r.Response.WriteJson(clustermessage.NewErrorResp("parse message error"))
		return
	}
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

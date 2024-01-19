package server

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/mtgnorton/ws-cluster/core/queue"
	"github.com/mtgnorton/ws-cluster/http/message"
)

type gfServer struct {
	opts   Options
	server *ghttp.Server
}

func New(opts ...Option) Server {
	return &gfServer{
		opts:   NewOptions(opts...),
		server: g.Server("http"),
	}
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

	ctx := g.opts.ctx
	g.server.BindHandler("/", func(r *ghttp.Request) {
		_, err := message.Parse(r.GetBody())
		if err != nil {
			r.Response.WriteJson(message.NewErrorRes("parse message error"))
			return
		}
		err = queue.DefaultQueue.Publish(ctx, queue.TopicDefault, r.GetBody())
		if err != nil {
			r.Response.WriteJson(message.NewErrorRes("publish message error"))
			return
		}
		r.Response.WriteJson(message.NewSuccessRes())
	})
	g.opts.shared.Logger.Debugf("http server run on port:%d", g.opts.port)
	g.server.SetPort(g.opts.port)
	g.server.Run()
}

func (g gfServer) Stop() error {
	return g.server.Shutdown()
}

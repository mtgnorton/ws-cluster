package server

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/mtgnorton/ws-cluster/core/queue"
	"github.com/mtgnorton/ws-cluster/http/message"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"
	"github.com/mtgnorton/ws-cluster/tools/wssentry"
)

type gfServer struct {
	opts   Options
	server *ghttp.Server
	sentry *wssentry.Handler
}

func New(opts ...Option) Server {
	return &gfServer{
		opts:   NewOptions(opts...),
		server: g.Server("http"),
		sentry: wssentry.GfSentry,
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
	fmt.Println(111)

	g.server.Use(g.sentry.MiddleWare)
	g.server.BindHandler("/", func(r *ghttp.Request) {
		beginTime := time.Now()

		g.sentry.RecoverHttp(r, handler)

		// prometheus add metrics
		metricManager := g.opts.prometheus.Options().MetricManager
		_ = metricManager.Get(wsprometheus.MetricRequestTotal).Inc(nil)
		_ = metricManager.Get(wsprometheus.MetricRequestURLTotal).Inc([]string{
			"http",
			strconv.Itoa(r.Response.Status),
		})
		_ = metricManager.Get(wsprometheus.MetricRequestDuration).Observe([]string{"http"}, time.Since(beginTime).Seconds())

	})
	g.opts.shared.Logger.Infof("http server run on port:%d", g.opts.port)
	g.server.EnablePProf()
	g.server.SetPort(g.opts.port)
	g.server.Run()
}

func (g gfServer) Stop() error {
	return g.server.Shutdown()
}

func handler(r *ghttp.Request) {

	_, err := message.Parse(r.GetBody())
	if err != nil {
		r.Response.WriteJson(message.NewErrorRes("parse message error"))
		return
	}
	// 随机休眠0-10s
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	err = queue.DefaultQueue.Publish(r.Context(), queue.TopicDefault, r.GetBody())
	if err != nil {
		r.Response.WriteJson(message.NewErrorRes("publish message error"))
		return
	}
	r.Response.WriteJson(message.NewSuccessRes())
}

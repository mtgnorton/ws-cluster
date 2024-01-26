package server

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/tools/wssentry"
	"github.com/mtgnorton/ws-cluster/ws/message"
)

type gfServer struct {
	opts   Options
	server *ghttp.Server
	sentry *wssentry.Handler
}

func New(opts ...Option) Server {
	return &gfServer{
		opts:   NewOptions(opts...),
		server: g.Server("ws"),
		sentry: wssentry.GfSentry,
	}
}

func (s *gfServer) Name() string {
	return "gf"
}

func (s *gfServer) Init(opts ...Option) {
	for _, o := range opts {
		o(&s.opts)
	}
}

func (s *gfServer) Options() Options {
	return s.opts
}

func (s *gfServer) Run() {
	s.server.Use(s.sentry.MiddleWare)
	s.server.BindHandler("/connect", func(r *ghttp.Request) {
		s.sentry.RecoverHttp(r, s.connect)
	})
	s.server.SetServerRoot(gfile.MainPkgPath())
	s.opts.shared.Logger.Debugf("ws server run on port:%d", s.opts.port)
	s.server.SetPort(s.opts.port)
	s.server.Run()
}

func (s *gfServer) Stop() error {
	return s.server.Shutdown()
}

func (s *gfServer) connect(r *ghttp.Request) {

	logger := s.opts.shared.Logger

	if !s.auth(r) {
		return
	}
	socket, err := r.WebSocket()
	if err != nil {
		logger.Debugf("Websocket err:%v", err)
		r.Exit()
	}

	uid := r.Get("uid").String()
	if uid == "" {
		logger.Debugf("Websocket uid is empty")
		r.Exit()
	}
	pid := r.Get("pid").String()
	if pid == "" {
		logger.Debugf("Websocket pid is empty")
		r.Exit()
	}

	c := client.NewClient(uid, pid, socket.Conn)
	s.opts.manager.Join(c)
	c.Send(message.NewSuccessRes("connect success", ""))
	for {
		_, rawMsg, err := socket.ReadMessage()

		//if hub := wssentry.GetHubFromContext(r); hub != nil {
		//	hub.WithScope(func(scope *sentry.Scope) {
		//		scope.SetExtra("gf_sentry_key˚", "11111")
		//	})
		//}

		if err != nil {
			logger.Infof("Websocket ReadMessage err: %v", err)
			s.opts.manager.Remove(c)
			return
		}
		s.opts.handler.Handle(c, rawMsg)
	}
}

func (s *gfServer) auth(r *ghttp.Request) bool {
	return true
}

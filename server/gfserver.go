package server

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/mtgnorton/ws-cluster/client"
	"github.com/mtgnorton/ws-cluster/queue"
)

type gfServer struct {
	opts   Options
	server *ghttp.Server
}

func New(opts ...Option) Server {
	return &gfServer{
		opts:   newOptions(opts...),
		server: g.Server(),
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
	s.server.BindHandler("/connect", func(r *ghttp.Request) {
		s.connect(r)
	})
	s.server.SetServerRoot(gfile.MainPkgPath())
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

	c := client.NewClient("1", "2", socket.Conn)
	s.opts.manager.Join(c)
	for {
		_, rawMsg, err := socket.ReadMessage()
		if err != nil {
			logger.Infof("Websocket ReadMessage err: %v", err)
			s.opts.manager.Remove(c)
			return
		}
		logger.Debugf("Websocket ReadMessage rawMsg: %s", rawMsg)

		err = s.opts.queue.Publish(s.opts.ctx, queue.TopicDefault, rawMsg)
		if err != nil {
			logger.Warnf("Websocket Publish err: %v", err)
			continue
		}

	}
}

func (s *gfServer) auth(r *ghttp.Request) bool {
	return true
}

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mtgnorton/ws-cluster/shared/auth"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
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
	s.opts.logger.Debugf(context.Background(), "ws server run on port:%d", s.opts.port)
	s.server.SetPort(s.opts.port)

	if s.opts.config.Values().Router.Enable {
		go s.registerToRegistryLoop()
	}
	go s.sendClientsLoop()

	s.server.Run()

}

func (s *gfServer) Stop() error {
	return s.server.Shutdown()
}

func (s *gfServer) connect(r *ghttp.Request) {
	ctx := r.Context()

	logger := s.opts.logger

	socket, err := r.WebSocket()
	if err != nil {
		logger.Debugf(ctx, "Websocket err:%v", err)
		r.Exit()
	}

	token := r.Get("token").String()

	// claims, err := shared.DefaultJwtWs.Parse(token)

	userData, err := auth.Decode(token)
	if err != nil {
		_ = socket.WriteMessage(1, []byte("token error"))
		_ = socket.Close()
		logger.Debugf(ctx, "Websocket token is error:%v", err)
		r.Exit()
	}

	c := client.NewClient(ctx, userData.UID, userData.PID, client.CType(userData.ClientType), socket.Conn)

	s.opts.manager.Join(ctx, c)

	s.opts.handler.Handle(ctx, c, &clustermessage.AffairMsg{
		Type: clustermessage.TypeConnect,
	})

	s.addMetrics()
	c.Send(ctx, clustermessage.NewSuccessResp("connect success"))
	for {

		//if hub := wssentry.GetHubFromContext(r); hub != nil {
		//	hub.WithScope(func(scope *sentry.Scope) {
		//		scope.SetExtra("gf_sentry_key˚", "11111")
		//	})
		//}
		_, msgBytes, err := socket.ReadMessage()

		if err != nil {
			logger.Infof(ctx, "Websocket Read err: %v", err)
			s.opts.manager.Remove(ctx, c)
			s.opts.handler.Handle(ctx, c, &clustermessage.AffairMsg{
				Type: clustermessage.TypeDisconnect,
			})
			err = s.opts.prometheus.GetAdd(wsprometheus.MetricWsConnection, nil, -1)
			if err != nil {
				logger.Infof(ctx, "Websocket GetAdd err: %v", err)
			}
			return
		}

		msg, err := clustermessage.ParseAffair(msgBytes)
		if err != nil {
			logger.Infof(ctx, "parse err:%v", err)
			continue
		}
		s.opts.handler.Handle(ctx, c, msg)
	}
}

func (s *gfServer) registerToRegistryLoop() {
	routerAddr := s.opts.config.Values().Router.Addr
	if routerAddr == "" {
		return
	}
	outHost := s.opts.config.Values().Router.OutHost
	addr := fmt.Sprintf("ws://%s:%d/connect", outHost, s.opts.port)
	// 注册到路由
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
	}()
	ctx := context.Background()

	for range ticker.C {

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second) // nolint

		_, err := g.Client().Post(ctx, routerAddr, g.Map{
			"addr": addr,
		})

		if err != nil {
			cancel()
			s.opts.logger.Infof(ctx, "register to router err:%v", err)
			continue
		}
		// content := r.ReadAllString()
		// s.opts.logger.Infof(ctx, "register to router response:%s", content)
		cancel()
	}
}

// sendClientsLoop 定时推送用户端的连接信息
func (s *gfServer) sendClientsLoop() {
	var (
		ctx = s.opts.ctx
	)
	for range time.Tick(2 * time.Second) {
		// 获取所有的用户端的连接信息
		// 遍历所有的服务端
		// 发送给服务端
		for _, projectServerClients := range s.opts.manager.Projects(ctx) {
			cids := make([]string, 0)
			for _, c := range projectServerClients.Clients {
				cid, _, _ := c.GetIDs()
				cids = append(cids, cid)
			}
			msg := clustermessage.AffairMsg{
				AffairID: "",
				AckID:    "",
				Payload:  cids,
				Type:     clustermessage.TypeOnlineClients,
				Source:   nil,
				To: &clustermessage.To{
					PID:  projectServerClients.PID,
					UIDs: nil,
					CIDs: nil,
				},
			}
			for _, s := range projectServerClients.Servers {
				s.Send(ctx, &msg)
			}
		}
	}
}

func (s *gfServer) addMetrics() {
	// prometheus add metrics
	p := s.opts.prometheus
	//_ = p.GetAdd(wsprometheus.MetricRequestTotal, nil, 1)
	//_ = p.GetAdd(wsprometheus.MetricRequestURLTotal, []string{"ws", "200"}, 1)
	_ = p.GetAdd(wsprometheus.MetricWsConnection, nil, 1)

}

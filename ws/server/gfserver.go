package server

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"ws-cluster/shared"
	"ws-cluster/shared/auth"

	"ws-cluster/clustermessage"
	"ws-cluster/core/client"
	"ws-cluster/tools/wsprometheus"
	"ws-cluster/tools/wssentry"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gorilla/websocket"
)

type gfServer struct {
	opts         Options
	server       *ghttp.Server
	sentry       *wssentry.Handler
	onlineNumber *atomic.Int64
}

func New(opts ...Option) Server {
	return &gfServer{
		opts:         NewOptions(opts...),
		server:       g.Server("ws"),
		sentry:       wssentry.GfSentry,
		onlineNumber: &atomic.Int64{},
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
	s.server.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.Write("ok")
	})
	s.server.SetServerRoot(gfile.MainPkgPath())
	s.opts.logger.Debugf(context.Background(), "ws server run on port:%d", s.opts.port)
	s.server.SetPort(s.opts.port)

	if s.opts.config.Values().Router.Enable {
		go s.registerToRegistryLoop()
	}
	go s.printOnlineNumber()
	s.server.Run()

}

func (s *gfServer) Stop() error {
	return s.server.Shutdown()
}

func (s *gfServer) connect(r *ghttp.Request) {
	ctx := r.Context()

	logger := s.opts.logger
	token := r.Get("token").String()
	userData, err := auth.Decode(token)

	var socket *ghttp.WebSocket

	if err != nil {
		socket, err = r.WebSocket()
		if err != nil {
			logger.Debugf(ctx, "Websocket err:%v", err)
			r.Exit()
		}
		_ = socket.WriteMessage(1, []byte("token error"))
		_ = socket.Close()
		logger.Debugf(ctx, "Websocket token is error:%v", err)
		r.Exit()
	} else {
		if userData.ClientType == int(client.CTypeServer) {
			socket, err = customSizeWebsocket(r, 16384)
			if err != nil {
				logger.Debugf(ctx, "Websocket err:%v", err)
				r.Exit()
			}
		} else {
			socket, err = customSizeWebsocket(r, 2048)
			if err != nil {
				logger.Debugf(ctx, "Websocket err:%v", err)
				r.Exit()
			}
		}
	}

	// claims, err := shared.DefaultJwtWs.Parse(token)

	c := client.NewClient(ctx, userData.UID, userData.PID, client.CType(userData.ClientType), socket.Conn)

	s.opts.manager.Join(ctx, c)

	s.opts.handler.Handle(ctx, c, &clustermessage.AffairMsg{
		Type: clustermessage.TypeConnect,
	})

	cID, _, _ := c.GetIDs()
	logger.Infof(ctx, "new client connect:%s", cID)

	s.addMetrics()

	nodeID := shared.GetNodeID()
	connectMsg := fmt.Sprintf("connect to node:%d success,clientID:%s", nodeID, cID)
	c.Send(ctx, clustermessage.NewSuccessResp(connectMsg))

	s.onlineNumber.Add(1)

	for {
		//if hub := wssentry.GetHubFromContext(r); hub != nil {
		//	hub.WithScope(func(scope *sentry.Scope) {
		//		scope.SetExtra("gf_sentry_key˚", "11111")
		//	})
		//}
		_, msgBytes, err := socket.ReadMessage()
		if err != nil {
			logger.Debugf(ctx, "Websocket Read err: %v", err)
			fmt.Printf("client disconnect:%s\n", c.String())
			s.opts.manager.Remove(ctx, c)
			s.opts.handler.Handle(ctx, c, &clustermessage.AffairMsg{
				Type: clustermessage.TypeDisconnect,
			})
			nodeID := shared.GetNodeID()
			serverIP := shared.ServerIP
			err = s.opts.prometheus.GetAdd(wsprometheus.MetricWsConnection, []string{fmt.Sprintf("%d", nodeID), serverIP}, -1)
			if err != nil {
				logger.Infof(ctx, "Websocket GetAdd err: %v", err)
			}
			s.onlineNumber.Add(-1)
			return
		}
		c.UpdateInteractTime()
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

func (s *gfServer) addMetrics() {
	// prometheus add metrics
	p := s.opts.prometheus
	//_ = p.GetAdd(wsprometheus.MetricRequestTotal, nil, 1)
	//_ = p.GetAdd(wsprometheus.MetricRequestURLTotal, []string{"ws", "200"}, 1)
	nodeID := shared.GetNodeID()
	serverIP := shared.ServerIP
	_ = p.GetAdd(wsprometheus.MetricWsConnection, []string{fmt.Sprintf("%d", nodeID), serverIP}, 1)

}

func (s *gfServer) printOnlineNumber() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.opts.logger.Infof(s.opts.ctx, "online number:%d", s.onlineNumber.Load())
	}
}

func customSizeWebsocket(r *ghttp.Request, size int) (*ghttp.WebSocket, error) {
	var upgrader = websocket.Upgrader{
		WriteBufferSize: size,
		ReadBufferSize:  size,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}

	if conn, err := upgrader.Upgrade(r.Response.Writer, r.Request, nil); err == nil {
		return &ghttp.WebSocket{
			Conn: conn,
		}, nil
	} else {
		return nil, err
	}
}

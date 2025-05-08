package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/shared/auth"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"
	"github.com/mtgnorton/ws-cluster/tools/wssentry"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gorilla/websocket"
)

type gfServer struct {
	opts         Options
	server       *ghttp.Server
	sentry       *wssentry.Handler
	onlineNumber sync.Map
}

func New(opts ...Option) Server {
	return &gfServer{
		opts:         NewOptions(opts...),
		server:       g.Server("ws"),
		sentry:       wssentry.GfSentry,
		onlineNumber: sync.Map{},
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
	} else if !s.opts.checking.IsExist(userData.PID) {
		socket, err = r.WebSocket()
		if err != nil {
			logger.Debugf(ctx, "Websocket err:%v", err)
			r.Exit()
		}
		_ = socket.WriteMessage(1, []byte("pid error"))
		_ = socket.Close()
		logger.Debugf(ctx, "Websocket pid is error:%v", err)
		r.Exit()
	}

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
	// todo 后续独立版去掉
	if userData.PID == "66" {
		userData.PID = "77"
	}

	// claims, err := shared.DefaultJwtWs.Parse(token)

	c := client.NewClient(ctx, userData.UID, userData.PID, client.CType(userData.ClientType), socket.Conn)

	s.opts.manager.Join(ctx, c)

	s.opts.handler.Handle(ctx, c, &clustermessage.AffairMsg{
		Type: clustermessage.TypeConnect,
	})

	cID, _, _ := c.GetIDs()
	logger.Debugf(ctx, "new client connect:%s", cID)

	s.addMetrics()

	nodeID := shared.GetNodeID()
	connectMsg := fmt.Sprintf("connect to node:%d success,clientID:%s", nodeID, cID)
	c.Send(ctx, clustermessage.NewSuccessResp(connectMsg))

	number, _ := s.onlineNumber.LoadOrStore(userData.PID, 1)
	s.onlineNumber.Store(userData.PID, number.(int)+1)

	for {
		//if hub := wssentry.GetHubFromContext(r); hub != nil {
		//	hub.WithScope(func(scope *sentry.Scope) {
		//		scope.SetExtra("gf_sentry_key˚", "11111")
		//	})
		//}
		_, msgBytes, err := socket.ReadMessage()
		if err != nil {
			logger.Debugf(ctx, "Websocket Read err: %v", err)
			s.opts.manager.Remove(ctx, c)
			s.opts.handler.Handle(ctx, c, &clustermessage.AffairMsg{
				Type: clustermessage.TypeDisconnect,
			})
			nodeID := shared.GetNodeID()
			serverIP := shared.GetInternalIP()
			err = s.opts.prometheus.GetAdd(wsprometheus.MetricWsConnection, []string{fmt.Sprintf("%d", nodeID), serverIP}, -1)
			if err != nil {
				logger.Infof(ctx, "Websocket GetAdd err: %v", err)
			}
			number, _ := s.onlineNumber.Load(userData.PID)
			s.onlineNumber.Store(userData.PID, number.(int)-1)
			return
		}
		msg, err := clustermessage.ParseAffair(msgBytes)
		if err != nil {
			logger.Infof(ctx, "parse err:%v", err)
			continue
		}
		c.UpdateInteractTime()

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
	serverIP := shared.GetInternalIP()
	_ = p.GetAdd(wsprometheus.MetricWsConnection, []string{fmt.Sprintf("%d", nodeID), serverIP}, 1)

}

func (s *gfServer) printOnlineNumber() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		prompt := "current online number,"
		total := 0
		s.onlineNumber.Range(func(key, value any) bool {
			prompt += fmt.Sprintf(" %s:%d,", key, value)
			total += value.(int)
			return true
		})
		prompt += fmt.Sprintf(" total:%d", total)
		s.opts.logger.Infof(s.opts.ctx, prompt)
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

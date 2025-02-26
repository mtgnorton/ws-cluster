package client

import (
	"context"
	"ws-cluster/logger"

	"github.com/gorilla/websocket"
	"github.com/sasha-s/go-deadlock"
)

type ClientMsg struct {
	ClientID string
	Message  interface{}
}

type SendManager struct {
	ctx     context.Context
	sockets map[string]*websocket.Conn // key: clientID
	ch      chan *ClientMsg
	deadlock.RWMutex
	logger logger.Logger
}

func NewSendManager(ctx context.Context, logger logger.Logger) *SendManager {
	return &SendManager{
		ctx:     ctx,
		sockets: make(map[string]*websocket.Conn),
		ch:      make(chan *ClientMsg, 10000),
		logger:  logger,
	}
}

func (s *SendManager) AddSocket(clientID string, socket *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	s.sockets[clientID] = socket
}

func (s *SendManager) RemoveSocket(clientID string) {
	s.Lock()
	defer s.Unlock()
	delete(s.sockets, clientID)
}

func (s *SendManager) Send(ctx context.Context, clientID string, message interface{}) {
	s.ch <- &ClientMsg{
		ClientID: clientID,
		Message:  message,
	}
}

func (s *SendManager) Run() {
	for {
		select {
		case message := <-s.ch:
			s.Lock()
			defer s.Unlock()
			socket, ok := s.sockets[message.ClientID]
			if !ok {
				s.logger.Debugf(s.ctx, "client:%s not found", message.ClientID)
				continue
			}
			if err := socket.WriteJSON(message.Message); err != nil {
				s.logger.Debugf(s.ctx, "client:%s send message error:%v", message.ClientID, err)
				continue
			}
		case <-s.ctx.Done():
			return
		}
	}
}

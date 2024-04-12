package client

import (
	"context"
	"fmt"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/sasha-s/go-deadlock"

	"github.com/gorilla/websocket"
)

type defaultClient struct {
	opts             *Options
	ID               string
	UID              string
	PID              string
	cancel           context.CancelFunc
	cType            CType           // 用户端还是服务端
	socket           *websocket.Conn // 连接
	lastInteractTime int64
	messageChan      chan interface{}
	deadlock.RWMutex
}

func (c *defaultClient) Init(opts ...Option) {
	for _, o := range opts {
		o(c.opts)
	}
}

func (c *defaultClient) Options() Options {
	return *c.opts
}

//func (c *defaultClient) Read(ctx context.Context) (msg *wsmessage.Req, isTerminate bool, err error) {
//	_, msgBytes, err := c.socket.ReadMessage()
//	if err != nil {
//		return nil, true, err
//	}
//	msg, err = c.opts.messageProcessor.ReqDecode(msgBytes)
//	return
//}

func (c *defaultClient) Send(ctx context.Context, message interface{}) {
	defer func() {
		if err := recover(); err != nil {
			c.opts.logger.Debugf(ctx, "send message recover err:%v", err)
		}
	}()
	c.RLock()
	defer c.RUnlock()

	if c.messageChan == nil {
		return
	}
	select {
	case <-ctx.Done():
		return
	case c.messageChan <- message:
	default:

		c.opts.logger.Debugf(ctx, "client:%s,send message:%v failed", c, message)
	}
}

func (c *defaultClient) Close() {
	c.Lock()
	defer c.Unlock()
	close(c.messageChan)
	c.cancel()
	c.messageChan = nil
	c.opts.logger.Debugf(context.Background(), "client close:%s", c.ID)
}

func (c *defaultClient) Status() Status {
	c.RLock()
	defer c.RUnlock()
	if c.messageChan == nil {
		return StatusClosed
	}
	return StatusNormal
}

func (c *defaultClient) UpdateInteractTime() {
	c.Lock()
	defer c.Unlock()
	c.lastInteractTime = time.Now().Unix()
}

func (c *defaultClient) GetInteractTime() int64 {
	c.RLock()
	defer c.RUnlock()
	return c.lastInteractTime
}

func (c *defaultClient) GetIDs() (id string, uid string, pid string) {
	c.RLock()
	defer c.RUnlock()
	return c.ID, c.UID, c.PID
}

func (c *defaultClient) Type() CType {
	c.RLock()
	defer c.RUnlock()
	return c.cType
}

func (c *defaultClient) String() string {
	return fmt.Sprintf("User[ID:%s,UID:%s,PID:%s,Type:%s]", c.ID, c.UID, c.PID, c.cType)
}

func (c *defaultClient) sendLoop(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			c.opts.logger.Debugf(ctx, "send loop done")
			return
		case message := <-c.messageChan:
			if err := c.socket.WriteJSON(message); err != nil {
				c.opts.logger.Debugf(ctx, "send message error:%v", err)
				return
			}
		}
	}
}

// NewClient 创建一个新的客户端,uid,pid为用户id和项目id,socket为websocket连接
func NewClient(ctx context.Context, uid string, pid string, cType CType, socket *websocket.Conn, options ...Option) Client {
	ctx, cancel := context.WithCancel(ctx)
	options = append(options, WithContext(ctx))

	opts := NewOptions(options...)
	c := &defaultClient{
		opts:        opts,
		ID:          shared.SnowflakeNode.Generate().String(),
		UID:         uid,
		PID:         pid,
		cancel:      cancel,
		cType:       cType,
		socket:      socket,
		messageChan: make(chan interface{}),
	}

	go c.sendLoop(ctx)
	return c
}

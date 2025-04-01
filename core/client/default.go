package client

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"ws-cluster/logger"
	"ws-cluster/shared"
	"ws-cluster/shared/kit"

	"github.com/sasha-s/go-deadlock"

	"github.com/gorilla/websocket"
)

var allClientSendStatistics atomic.Int64
var allClientSendStatisticsCh = make(chan *atomic.Int64)

func init() {
	go kit.Sampling(allClientSendStatisticsCh, time.Second*10, 0, func(v *atomic.Int64) {
		logger.DefaultLogger.Infof(context.Background(), "client send statistics:%v", v.Load())
	})
}

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
	status           atomic.Int32
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
		if r := recover(); r != nil {
			c.opts.logger.Warnf(ctx, "PANIC client:%s,send message panic,message is %v,panic is:%v", c, message, r)
		}
	}()
	if c.status.Load() == int32(StatusClosed) {
		c.opts.logger.Debugf(ctx, "client:%s,send message:%v ,client is closed", c, message)
		return
	}

	c.RLock()
	if c.status.Load() == int32(StatusClosed) || c.messageChan == nil {
		c.RUnlock()
		return
	}

	select {
	case c.messageChan <- message:
	default:
		c.opts.logger.Warnf(ctx, "client:%s,send message:%v ,channel is full,dropped", c, message)
	}
	c.RUnlock()

}

func (c *defaultClient) Close() {
	if !c.status.CompareAndSwap(int32(StatusNormal), int32(StatusClosed)) {
		return
	}

	c.Lock()
	defer c.Unlock()

	c.cancel()

	if c.messageChan != nil {
		close(c.messageChan)
		c.messageChan = nil
	}
	c.socket.Close()

	c.opts.logger.Debugf(context.Background(), "client close:%s", c.ID)
}

func (c *defaultClient) Status() Status {
	return Status(c.status.Load())
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

func (c *defaultClient) GetCID() string {
	c.RLock()
	defer c.RUnlock()
	return c.ID
}

func (c *defaultClient) GetUID() string {
	c.RLock()
	defer c.RUnlock()
	return c.UID
}

func (c *defaultClient) GetPID() string {
	c.RLock()
	defer c.RUnlock()
	return c.PID
}

func (c *defaultClient) Type() CType {
	c.RLock()
	defer c.RUnlock()
	return c.cType
}

func (c *defaultClient) String() string {
	return fmt.Sprintf("Client[ID:%s,UID:%s,PID:%s,Type:%s]", c.ID, c.UID, c.PID, c.cType)
}

func (c *defaultClient) sendLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.opts.logger.Warnf(ctx, "client:%s sendLoop panic: %v", c.ID, r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.opts.logger.Debugf(ctx, "client:%s send loop done by context", c.ID)
			return
		case message, ok := <-c.messageChan:
			if !ok {
				c.opts.logger.Debugf(ctx, "client:%s send loop done by closed channel", c.ID)
				return
			}

			if c.status.Load() == int32(StatusClosed) {
				c.opts.logger.Debugf(ctx, "client:%s is closed, stop sending", c.ID)
				return
			}

			if err := c.socket.WriteJSON(message); err != nil {
				c.opts.logger.Debugf(ctx, "client:%s send message error:%v", c.ID, err)
				c.Close()
				return
			}

			allClientSendStatistics.Add(1)
			allClientSendStatisticsCh <- &allClientSendStatistics
		}
	}
}

// NewClient 创建一个新的客户端,uid,pid为用户id和项目id,socket为websocket连接
func NewClient(ctx context.Context, uid string, pid string, cType CType, socket *websocket.Conn, options ...Option) Client {
	ctx, cancel := context.WithCancel(ctx)
	options = append(options, WithContext(ctx))

	opts := NewOptions(options...)
	messageChan := make(chan interface{}, 500)
	if cType == CTypeServer {
		messageChan = make(chan interface{}, 20000)
	}
	c := &defaultClient{
		opts:        opts,
		ID:          shared.GetSnowflakeNode().Generate().String(),
		UID:         uid,
		PID:         pid,
		cancel:      cancel,
		cType:       cType,
		socket:      socket,
		messageChan: messageChan,
	}
	c.status.Store(int32(StatusNormal))
	go c.sendLoop(ctx)
	return c
}

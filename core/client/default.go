package client

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/shared/kit"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"

	"github.com/gorilla/websocket"
)

type outboundMessage struct {
	payload    interface{}
	enqueuedAt time.Time
}

type defaultClient struct {
	opts             *Options
	ID               string
	UID              string
	PID              string
	cancel           context.CancelFunc
	cType            CType           // 用户端还是服务端
	socket           *websocket.Conn // 连接
	lastInteractTime atomic.Int64
	messageChan      chan *outboundMessage
	metricLabels     []string
	lastSlowLogAt    atomic.Int64
	lastDropLogAt    atomic.Int64
	status           atomic.Int32
	sync.RWMutex
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
	case c.messageChan <- &outboundMessage{
		payload:    message,
		enqueuedAt: time.Now(),
	}:
	default:
		_ = wsprometheus.DefaultPrometheus.GetAdd(wsprometheus.MetricClientSendDrop, c.metricLabels, 1)
		if kit.AllowByInterval(&c.lastDropLogAt, 2*time.Second) {
			c.opts.logger.Warnf(ctx, "client:%s send queue full,dropped,len=%d,cap=%d", c.ID, len(c.messageChan), cap(c.messageChan))
		}
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
	c.lastInteractTime.Store(time.Now().Unix())
}

func (c *defaultClient) GetInteractTime() int64 {
	return c.lastInteractTime.Load()
}

func (c *defaultClient) GetIDs() (id string, uid string, pid string) {
	return c.ID, c.UID, c.PID
}

func (c *defaultClient) GetCID() string {
	return c.ID
}

func (c *defaultClient) GetUID() string {
	return c.UID
}

func (c *defaultClient) GetPID() string {
	return c.PID
}

func (c *defaultClient) Type() CType {
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

			queueWaitMs := float64(time.Since(message.enqueuedAt).Microseconds()) / 1000.0
			_ = wsprometheus.DefaultPrometheus.GetObserve(wsprometheus.MetricClientSendQueueWaitDuration, c.metricLabels, queueWaitMs)
			if queueWaitMs >= 1000 && kit.AllowByInterval(&c.lastSlowLogAt, 2*time.Second) {
				c.opts.logger.Warnf(ctx, "client:%s send queue wait=%0.2fms,len=%d,cap=%d,type=%s,pid=%s,message=%s", c.ID, queueWaitMs, len(c.messageChan), cap(c.messageChan), c.cType, c.PID, kit.LogSnippet(message.payload, 240))
			}

			writeBegin := time.Now()
			if err := c.socket.WriteJSON(message.payload); err != nil {
				c.opts.logger.Debugf(ctx, "client:%s send message error:%v", c.ID, err)
				c.Close()
				return
			}
			writeMs := float64(time.Since(writeBegin).Microseconds()) / 1000.0
			_ = wsprometheus.DefaultPrometheus.GetObserve(wsprometheus.MetricClientWriteDuration, c.metricLabels, writeMs)
			if writeMs >= 200 && kit.AllowByInterval(&c.lastSlowLogAt, 2*time.Second) {
				c.opts.logger.Warnf(ctx, "client:%s websocket write slow=%0.2fms,type=%s,pid=%s,message=%s", c.ID, writeMs, c.cType, c.PID, kit.LogSnippet(message.payload, 240))
			}
		}
	}
}

// NewClient 创建一个新的客户端,uid,pid为用户id和项目id,socket为websocket连接
func NewClient(ctx context.Context, uid string, pid string, cType CType, socket *websocket.Conn, options ...Option) Client {
	ctx, cancel := context.WithCancel(ctx)
	options = append(options, WithContext(ctx))

	opts := NewOptions(options...)
	messageChan := make(chan *outboundMessage, 500)
	if cType == CTypeServer {
		messageChan = make(chan *outboundMessage, 20000)
	}
	nodeID := shared.GetNodeID()
	nodeIP := shared.GetInternalIP()
	c := &defaultClient{
		opts:         opts,
		ID:           shared.GetSnowflakeNode().Generate().String(),
		UID:          uid,
		PID:          pid,
		cancel:       cancel,
		cType:        cType,
		socket:       socket,
		messageChan:  messageChan,
		metricLabels: []string{strconv.FormatInt(nodeID, 10), nodeIP, cType.String()},
	}
	c.status.Store(int32(StatusNormal))
	c.lastInteractTime.Store(time.Now().Unix())
	go c.sendLoop(ctx)
	return c
}

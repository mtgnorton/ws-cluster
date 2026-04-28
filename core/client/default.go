package client

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/core/tracing"
	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"

	"github.com/gorilla/websocket"
)

type outboundMessage struct {
	payload    interface{}
	enqueuedAt time.Time
	trace      *clustermessage.Trace
	meta       tracing.Meta
}

type Outbound struct {
	Payload interface{}
	Trace   *clustermessage.Trace
	Meta    tracing.Meta
}

func (m *outboundMessage) setTrace(trace *clustermessage.Trace) {
	if m == nil {
		return
	}
	m.trace = trace
	switch payload := m.payload.(type) {
	case *clustermessage.AffairMsg:
		payload.Trace = trace
	case clustermessage.AffairMsg:
		payload.Trace = trace
		m.payload = payload
	}
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
	node             tracing.Node
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

func (c *defaultClient) Send(ctx context.Context, message interface{}) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			c.opts.logger.Warnf(ctx, "PANIC client:%s,send message panic,message is %v,panic is:%v", c, message, r)
			ok = false
		}
	}()
	if c.status.Load() == int32(StatusClosed) {
		c.opts.logger.Debugf(ctx, "client:%s,send message:%v ,client is closed", c, message)
		return false
	}

	c.RLock()
	if c.status.Load() == int32(StatusClosed) || c.messageChan == nil {
		c.RUnlock()
		return false
	}

	outbound := prepareOutboundMessage(message)
	select {
	case c.messageChan <- outbound:
		ok = true
	default:
		_ = wsprometheus.DefaultPrometheus.GetAdd(wsprometheus.MetricClientSendDrop, c.metricLabels, 1)
		tracing.RecordTrace(ctx, c.opts.logger, outbound.trace, outbound.meta, c.node, tracing.Event{
			Name:  sendQueueFullEvent(c.cType),
			Force: true,
			Warn:  true,
		})
	}
	c.RUnlock()
	return ok
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

			queueWait := time.Since(message.enqueuedAt)
			queueWaitMs := float64(queueWait.Microseconds()) / 1000.0
			_ = wsprometheus.DefaultPrometheus.GetObserve(wsprometheus.MetricClientSendQueueWaitDuration, c.metricLabels, queueWaitMs)
			queueSlow := queueWaitMs >= 1000
			reason := "send_queue_dequeue"
			if queueSlow {
				reason = "send_queue_wait_slow"
			}
			message.setTrace(tracing.RecordTrace(ctx, c.opts.logger, message.trace, message.meta, c.node, tracing.Event{
				Name:       queueDequeueEvent(c.cType),
				Reason:     reason,
				DurationMs: tracing.DurationMs(queueWait),
				Force:      queueSlow,
			}))

			writeBegin := time.Now()
			if err := c.socket.WriteJSON(message.payload); err != nil {
				c.opts.logger.Debugf(ctx, "client:%s send message error:%v", c.ID, err)
				message.setTrace(tracing.RecordTrace(ctx, c.opts.logger, message.trace, message.meta, c.node, tracing.Event{
					Name:       wsWriteFailedEvent(c.cType),
					Reason:     "ws_write_failed",
					DurationMs: tracing.DurationMs(time.Since(writeBegin)),
					Force:      true,
					Warn:       true,
				}))
				c.Close()
				return
			}
			writeDuration := time.Since(writeBegin)
			writeMs := float64(writeDuration.Microseconds()) / 1000.0
			_ = wsprometheus.DefaultPrometheus.GetObserve(wsprometheus.MetricClientWriteDuration, c.metricLabels, writeMs)
			writeSlow := writeMs >= 200
			reason = "ws_write_done"
			if writeSlow {
				reason = "ws_write_slow"
			}
			message.setTrace(tracing.RecordTrace(ctx, c.opts.logger, message.trace, message.meta, c.node, tracing.Event{
				Name:       wsWriteDoneEvent(c.cType),
				Reason:     reason,
				DurationMs: tracing.DurationMs(writeDuration),
				Force:      writeSlow,
			}))
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
		node:         tracing.NodeInfo(strconv.FormatInt(nodeID, 10), nodeIP),
	}
	c.status.Store(int32(StatusNormal))
	c.lastInteractTime.Store(time.Now().Unix())
	go c.sendLoop(ctx)
	return c
}

func prepareOutboundMessage(message interface{}) *outboundMessage {
	outbound := &outboundMessage{
		payload:    message,
		enqueuedAt: time.Now(),
	}
	switch msg := message.(type) {
	case Outbound:
		outbound.payload = msg.Payload
		outbound.trace = msg.Trace.Clone()
		outbound.meta = msg.Meta
	case *clustermessage.AffairMsg:
		if msg == nil {
			return outbound
		}
		cp := *msg
		cp.Trace = msg.Trace.Clone()
		outbound.payload = &cp
		outbound.trace = cp.Trace
		outbound.meta = tracing.MetaFromMessage(&cp)
	case clustermessage.AffairMsg:
		cp := msg
		cp.Trace = msg.Trace.Clone()
		outbound.payload = cp
		outbound.trace = cp.Trace
		outbound.meta = tracing.MetaFromMessage(&cp)
	}
	return outbound
}

func queueDequeueEvent(cType CType) string {
	if cType == CTypeServer {
		return tracing.EventServerQueueDequeue
	}
	return tracing.EventClientQueueDequeue
}

func wsWriteDoneEvent(cType CType) string {
	if cType == CTypeServer {
		return tracing.EventServerWSWriteDone
	}
	return tracing.EventClientWSWriteDone
}

func wsWriteFailedEvent(cType CType) string {
	if cType == CTypeServer {
		return tracing.EventServerWSWriteFailed
	}
	return tracing.EventClientWSWriteFailed
}

func sendQueueFullEvent(cType CType) string {
	if cType == CTypeServer {
		return tracing.EventServerSendQueueFull
	}
	return tracing.EventClientSendQueueFull
}

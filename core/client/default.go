package client

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"

	"github.com/gorilla/websocket"
)

type outboundMessage struct {
	payload    interface{}
	enqueuedAt time.Time
	trace      *clustermessage.Trace
	affairID   string
	msgType    clustermessage.Type
	pid        string
	logPayload interface{}
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
	nodeID           string
	nodeIP           string
	status           atomic.Int32
	sync.RWMutex
}

type traceCarrier interface {
	MessageTrace() *clustermessage.Trace
	MessageAffairID() string
	MessageType() clustermessage.Type
	MessagePID() string
}

type payloadCarrier interface {
	MessagePayload() interface{}
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

	payload, trace, affairID, msgType, pid, logPayload := prepareOutboundMessage(message)
	select {
	case c.messageChan <- &outboundMessage{
		payload:    payload,
		enqueuedAt: time.Now(),
		trace:      trace,
		affairID:   affairID,
		msgType:    msgType,
		pid:        pid,
		logPayload: logPayload,
	}:
		ok = true
	default:
		_ = wsprometheus.DefaultPrometheus.GetAdd(wsprometheus.MetricClientSendDrop, c.metricLabels, 1)
		clustermessage.AddTraceEvent(trace, sendQueueFullEvent(c.cType), c.nodeID, c.nodeIP, 0, true)
		if clustermessage.ShouldLogTrace(trace, true) {
			c.opts.logger.Warnf(ctx, clustermessage.BuildTraceLogForPayload(trace, msgType, pid, affairID, c.nodeID, c.nodeIP, "send_queue_full", logPayload))
		}
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
			clustermessage.AddTraceEvent(message.trace, queueDequeueEvent(c.cType), c.nodeID, c.nodeIP, clustermessage.DurationMs(queueWait), queueSlow)
			if clustermessage.ShouldLogTrace(message.trace, queueSlow) {
				reason := "send_queue_dequeue"
				if queueSlow {
					reason = "send_queue_wait_slow"
				}
				c.opts.logger.Infof(ctx, clustermessage.BuildTraceLogForPayload(message.trace, message.msgType, message.pid, message.affairID, c.nodeID, c.nodeIP, reason, message.logPayload))
			}

			writeBegin := time.Now()
			if err := c.socket.WriteJSON(message.payload); err != nil {
				c.opts.logger.Debugf(ctx, "client:%s send message error:%v", c.ID, err)
				clustermessage.AddTraceEvent(message.trace, wsWriteFailedEvent(c.cType), c.nodeID, c.nodeIP, clustermessage.DurationMs(time.Since(writeBegin)), true)
				if clustermessage.ShouldLogTrace(message.trace, true) {
					c.opts.logger.Warnf(ctx, clustermessage.BuildTraceLogForPayload(message.trace, message.msgType, message.pid, message.affairID, c.nodeID, c.nodeIP, "ws_write_failed", message.logPayload))
				}
				c.Close()
				return
			}
			writeDuration := time.Since(writeBegin)
			writeMs := float64(writeDuration.Microseconds()) / 1000.0
			_ = wsprometheus.DefaultPrometheus.GetObserve(wsprometheus.MetricClientWriteDuration, c.metricLabels, writeMs)
			writeSlow := writeMs >= 200
			clustermessage.AddTraceEvent(message.trace, wsWriteDoneEvent(c.cType), c.nodeID, c.nodeIP, clustermessage.DurationMs(writeDuration), writeSlow)
			if clustermessage.ShouldLogTrace(message.trace, writeSlow) {
				reason := "ws_write_done"
				if writeSlow {
					reason = "ws_write_slow"
				}
				c.opts.logger.Infof(ctx, clustermessage.BuildTraceLogForPayload(message.trace, message.msgType, message.pid, message.affairID, c.nodeID, c.nodeIP, reason, message.logPayload))
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
		nodeID:       strconv.FormatInt(nodeID, 10),
		nodeIP:       nodeIP,
	}
	c.status.Store(int32(StatusNormal))
	c.lastInteractTime.Store(time.Now().Unix())
	go c.sendLoop(ctx)
	return c
}

func prepareOutboundMessage(message interface{}) (payload interface{}, trace *clustermessage.Trace, affairID string, msgType clustermessage.Type, pid string, logPayload interface{}) {
	payload = message
	switch msg := message.(type) {
	case *clustermessage.AffairMsg:
		if msg == nil {
			return
		}
		cp := *msg
		cp.Trace = msg.Trace.Clone()
		payload = &cp
		trace = cp.Trace
		affairID = cp.AffairID
		msgType = cp.Type
		pid = clustermessage.MessagePID(&cp)
		logPayload = cp.Payload
	case clustermessage.AffairMsg:
		cp := msg
		cp.Trace = msg.Trace.Clone()
		payload = cp
		trace = cp.Trace
		affairID = cp.AffairID
		msgType = cp.Type
		pid = clustermessage.MessagePID(&cp)
		logPayload = cp.Payload
	case traceCarrier:
		trace = msg.MessageTrace().Clone()
		affairID = msg.MessageAffairID()
		msgType = msg.MessageType()
		pid = msg.MessagePID()
		if carrier, ok := message.(payloadCarrier); ok {
			logPayload = carrier.MessagePayload()
		}
	}
	return
}

func queueDequeueEvent(cType CType) string {
	if cType == CTypeServer {
		return clustermessage.TraceEventServerQueueDequeue
	}
	return clustermessage.TraceEventClientQueueDequeue
}

func wsWriteDoneEvent(cType CType) string {
	if cType == CTypeServer {
		return clustermessage.TraceEventServerWSWriteDone
	}
	return clustermessage.TraceEventClientWSWriteDone
}

func wsWriteFailedEvent(cType CType) string {
	if cType == CTypeServer {
		return clustermessage.TraceEventServerWSWriteFailed
	}
	return clustermessage.TraceEventClientWSWriteFailed
}

func sendQueueFullEvent(cType CType) string {
	if cType == CTypeServer {
		return clustermessage.TraceEventServerSendQueueFull
	}
	return clustermessage.TraceEventClientSendQueueFull
}

package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mtgnorton/ws-cluster/shared"
	"sync"
	"time"
)

type defaultClient struct {
	opts            Options
	ID              string
	UID             string
	PID             string
	socket          *websocket.Conn // 连接
	lastReceiveTime int64
	messageChan     chan interface{}
	sync.RWMutex
}

func (d *defaultClient) Init(opts ...Option) {
	for _, o := range opts {
		o(&d.opts)
	}

}

func (d *defaultClient) Options() Options {
	return d.opts
}

func (d *defaultClient) Send(message interface{}) {
	if d.messageChan == nil {
		return
	}
	d.messageChan <- message
}

func (d *defaultClient) Close() {
	d.Lock()
	defer d.Unlock()
	if d.messageChan == nil {
		return
	}
	close(d.messageChan)
	d.messageChan = nil
}

func (d *defaultClient) Status() Status {
	d.RLock()
	defer d.RUnlock()
	if d.messageChan == nil {
		return StatusClosed
	}
	return StatusNormal
}

func (d *defaultClient) UpdateReplyTime() {
	d.Lock()
	defer d.Unlock()
	d.lastReceiveTime = time.Now().Unix()
}

func (d *defaultClient) GetReplyTime() int64 {
	d.RLock()
	defer d.RUnlock()
	return d.lastReceiveTime
}

func (d *defaultClient) GetIDs() (id string, uid string, pid string) {
	d.RLock()
	defer d.RUnlock()
	return d.ID, d.UID, d.PID
}

func (d *defaultClient) String() string {
	return fmt.Sprintf("Client[ID:%s,UID:%s,PID:%s]", d.ID, d.UID, d.PID)
}

func NewClient(uid string, pid string, socket *websocket.Conn, options ...Option) Client {
	opts := Options{
		SnowflakeNode: shared.DefaultShared.SnowflakeNode,
	}
	for _, o := range options {
		o(&opts)
	}
	return &defaultClient{
		opts:        opts,
		ID:          opts.SnowflakeNode.Generate().String(),
		UID:         uid,
		PID:         pid,
		socket:      socket,
		messageChan: make(chan interface{}),
	}
}

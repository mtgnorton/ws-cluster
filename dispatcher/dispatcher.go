package dispatcher

import (
	"github.com/mtgnorton/ws-cluster/handler"
)

var DefaultDispatcher = NewDefaultDispatcher()

type Dispatcher interface {
	Options() Options
	Dispatch(msg []byte) (isAck bool)
}

type defaultDispatcher struct {
	opts Options
}

func NewDefaultDispatcher(opts ...Option) Dispatcher {
	options := newOptions(opts...)
	return &defaultDispatcher{opts: options}
}

func (d *defaultDispatcher) Options() Options {
	return d.opts
}

func (d *defaultDispatcher) Dispatch(msg []byte) (isAck bool) {

	// Parse the message
	req, err := d.opts.parser.Parse(msg)
	if err != nil {
		d.opts.logger.Warnf("Failed to parse message: %s", err.Error())
		return false
	}
	var (
		h  handler.Handler
		ok bool
	)

	if h, ok = d.opts.handlers[req.Type]; !ok {
		d.opts.logger.Warnf("No handler for message type: %s", req.Type)
		return false
	}
	r, err := h.Handle(req)
	if err != nil {
		d.opts.logger.Warnf("Failed to handle message: %s", err.Error())
		return false
	}
	return r
}

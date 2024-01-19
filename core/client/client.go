package client

type Status int

const (
	StatusNormal Status = iota
	StatusClosed
)

type Client interface {
	Init(opts ...Option)
	Options() Options
	Send(message interface{})
	Close()
	Status() Status
	UpdateReplyTime()
	GetReplyTime() int64
	GetIDs() (id string, uid string, pid string)
	String() string
}

package client

import (
	"context"

	"github.com/mtgnorton/ws-cluster/message/wsmessage"
)

type Status int

const (
	StatusNormal Status = iota
	StatusClosed
)

type CType int

// 一个连接上来的client，可能是用户端，可能是服务端,也可能是管理端
const (
	CTypeUser   CType = 0 // 用户端
	CTypeServer CType = 1 // 服务端
	CTypeAdmin  CType = 2 // 管理端
)

func (c CType) String() string {
	switch c {
	case CTypeUser:
		return "user"
	case CTypeServer:
		return "server"
	case CTypeAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

type Client interface {
	Init(opts ...Option)
	Options() Options
	Read(ctx context.Context) (message *wsmessage.Req, isTerminate bool, err error)
	Send(ctx context.Context, message *wsmessage.Res)
	Close()
	Status() Status
	UpdateReplyTime()
	GetReplyTime() int64
	GetIDs() (id string, uid string, pid string)
	Type() CType
	String() string
}

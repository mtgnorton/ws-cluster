package client

import (
	"context"
)

// Status 客户端状态,正常或者关闭
type Status int

const (
	StatusNormal Status = iota
	StatusClosed
)

// CType 客户端类型,一个连接上来的client，可能是用户端，可能是服务端,也可能是管理端
type CType int

const (
	CTypeUser   CType = 0       // 用户端
	CTypeServer CType = 32832   // 服务端
	CTypeAdmin  CType = 5345345 // 管理端
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
	//Read(ctx context.Context) (message *wsmessage.Req, isTerminate bool, err error)
	// message 直接为golang类型
	Send(ctx context.Context, message interface{})
	Close()
	Status() Status
	UpdateInteractTime()
	GetInteractTime() int64
	GetIDs() (cid string, uid string, pid string)
	Type() CType
	String() string
}

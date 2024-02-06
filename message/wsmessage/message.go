package wsmessage

import (
	"errors"

	"github.com/mtgnorton/ws-cluster/message/queuemessage"
)

var ErrReadMessage = errors.New("read message error")

const (
	TypeSubscribe   = queuemessage.TypeSubscribe
	TypeUnsubscribe = queuemessage.TypeUnsubscribe
	TypeRequest     = queuemessage.TypeRequest
	TypePush        = queuemessage.TypePush

	TypeConnect    = queuemessage.TypeConnect
	TypeDisconnect = queuemessage.TypeDisconnect
)

// Req 通过ws传递的消息结构
//
// Type 必填
// Identification 可选，用于标识请求的唯一id,前端传递过来后，后端返回时会原样返回
// Payload 必填，根据不同的type，body的内容不同，由具体的handler解析
type Req struct {
	Type           queuemessage.Type `json:"type"`
	Identification string            `json:"identification"`
	Payload        interface{}       `json:"payload"` // 必填，根据不同的type，body的内容不同，由具体的handler解析
}

type Res struct {
	Code           int         `json:"code"`           // 1 成功，0失败
	Identification string      `json:"identification"` // 可选，用于标识请求的唯一id,前端传递过来后，后端返回时会原样返回
	Msg            string      `json:"msg"`            // 提示信息
	Payload        interface{} `json:"payload"`        // 根据不同的type，body的内容不同
}

func NewErrorRes(msg string, Identification string) *Res {
	return &Res{
		Code:           0,
		Identification: Identification,
		Msg:            msg,
		Payload:        "",
	}
}

func NewSuccessRes(msg string, Identification string, payload ...interface{}) *Res {
	if len(payload) > 0 {
		return &Res{
			Code:           1,
			Identification: Identification,
			Msg:            msg,
			Payload:        payload[0],
		}
	}
	return &Res{
		Code:           1,
		Identification: Identification,
		Msg:            msg,
		Payload:        "",
	}
}

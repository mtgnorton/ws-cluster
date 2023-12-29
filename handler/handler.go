package handler

import (
	"errors"
	"github.com/mtgnorton/ws-cluster/message"
)

var ErrInvalidPayload = errors.New("invalid payload")

type Handler interface {
	Type() message.Type // 获取handler的类型，subscribe, unsubscribe, push
	Handle(message *message.ReqMessage) (isAck bool, err error)
}

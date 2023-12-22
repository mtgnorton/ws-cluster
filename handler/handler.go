package handler

import "github.com/mtgnorton/ws-cluster/message"

type Handler interface {
	Type() message.Type // 获取handler的类型，subscribe, unsubscribe, push
	Handle(message *message.ReqMessage) (isAck bool)
}

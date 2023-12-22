package handler

import (
	"fmt"
	"github.com/mtgnorton/ws-cluster/message"
)

type PushHandler struct {
}

func (p PushHandler) Type() message.Type {

	return message.TypePush
}

func (p PushHandler) Handle(message *message.ReqMessage) (resMessage *message.ResMessage, err error) {
	fmt.Println(message)
	return
}

func NewPushHandler() *PushHandler {
	return &PushHandler{}
}

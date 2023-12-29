package handler

import (
	"fmt"
	"github.com/mtgnorton/ws-cluster/message"
)

type UnSubscribeHandler struct {
}

func (u UnSubscribeHandler) Type() message.Type {

	return message.TypeUnsubscribe
}

func (u UnSubscribeHandler) Handle(message *message.ReqMessage) (resMessage *message.ResMessage, err error) {
	fmt.Println(message)
	return
}

func NewUnSubscribeHandler() *UnSubscribeHandler {
	return &UnSubscribeHandler{}
}

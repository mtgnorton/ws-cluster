package handler

import (
	"fmt"
	"github.com/mtgnorton/ws-cluster/message"
)

type SubscribeHandler struct {
}

func (s SubscribeHandler) Type() message.Type {
	return message.TypeSubscribe
}

func (s SubscribeHandler) Handle(message *message.ReqMessage) (resMessage *message.ResMessage, err error) {
	fmt.Println(message)
	return
}

func NewSubscribeHandler() *SubscribeHandler {
	return &SubscribeHandler{}
}

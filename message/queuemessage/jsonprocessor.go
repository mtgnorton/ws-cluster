package queuemessage

import (
	"encoding/json"
)

type JsonProcessor struct{}

func (j JsonProcessor) Decode(messageBytes []byte) (message *Message, err error) {
	message = &Message{}
	err = json.Unmarshal(messageBytes, message)
	return
}

func (j JsonProcessor) Encode(message *Message) (messageBytes []byte, err error) {
	messageBytes, err = json.Marshal(message)
	return
}

func NewJsonProcessor() *JsonProcessor {
	return &JsonProcessor{}
}

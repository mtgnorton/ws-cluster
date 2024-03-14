package queuemessage

import (
	"encoding/json"
)

type QueueMSgJsonProcessor struct{}

func (j QueueMSgJsonProcessor) Decode(messageBytes []byte) (message *Message, err error) {
	message = &Message{}
	err = json.Unmarshal(messageBytes, message)
	return
}

func (j QueueMSgJsonProcessor) Encode(message *Message) (messageBytes []byte, err error) {
	messageBytes, err = json.Marshal(message)
	return
}

func NewJsonProcessor() *QueueMSgJsonProcessor {
	return &QueueMSgJsonProcessor{}
}

package wsmessage

import (
	"encoding/json"
)

type WsMsgJsonProcessor struct{}

func (j WsMsgJsonProcessor) ReqEncode(message *Req) (messageBytes []byte, err error) {
	messageBytes, err = json.Marshal(message)
	return
}
func (j WsMsgJsonProcessor) ReqDecode(msg []byte) (message *Req, err error) {
	message = &Req{}
	err = json.Unmarshal(msg, message)
	return
}

func (j WsMsgJsonProcessor) ResDecode(messageBytes []byte) (message *Res, err error) {
	message = &Res{}
	err = json.Unmarshal(messageBytes, message)
	return
}

func (j WsMsgJsonProcessor) ResEncode(message *Res) (messageBytes []byte, err error) {
	messageBytes, err = json.Marshal(message)
	return
}

func NewJsonProcessor() *WsMsgJsonProcessor {
	return &WsMsgJsonProcessor{}
}

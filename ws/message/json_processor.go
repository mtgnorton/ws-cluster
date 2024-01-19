package message

import "encoding/json"

type JsonProcessor struct{}

func (j JsonProcessor) Parse(msg []byte) (message *Req, err error) {
	message = &Req{}
	err = json.Unmarshal(msg, message)
	return
}

func (j JsonProcessor) Synthesize(message *Res) (msg []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func NewJsonProcessor() *JsonProcessor {
	return &JsonProcessor{}
}

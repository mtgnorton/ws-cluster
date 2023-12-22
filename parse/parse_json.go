package parse

import (
	"encoding/json"
	"github.com/mtgnorton/ws-cluster/message"
)

type JsonParser struct {
}

func NewJsonParser() *JsonParser {
	return &JsonParser{}
}

func (p *JsonParser) Parse(bytes []byte) (m *message.ReqMessage, err error) {
	m = &message.ReqMessage{}
	err = json.Unmarshal(bytes, m)
	return
}

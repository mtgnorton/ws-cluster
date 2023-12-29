package message

import (
	"encoding/json"
)

type JsonParser struct{}

func NewJsonParser() *JsonParser {
	return &JsonParser{}
}

func (p *JsonParser) Parse(bytes []byte) (m *WsReqMessage, err error) {
	m = &WsReqMessage{}
	err = json.Unmarshal(bytes, m)
	return
}

type JsonSynthesizer struct{}

func NewJsonSynthesizer() *JsonSynthesizer{
	return &JsonSynthesizer{}
}

func (s *JsonSynthesizer)

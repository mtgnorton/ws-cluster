package queuemessage

import "encoding/json"

type Type string

const (
	TypeSubscribe   Type = "subscribe"
	TypeUnsubscribe Type = "unsubscribe"
	TypePush        Type = "push"
	TypeRequest     Type = "request"
	TypeConnect     Type = "connect"
	TypeDisconnect  Type = "disconnect"
)

type Message struct {
	Type Type `json:"type"`
	//TraceID        string      `json:"trace_id"` // 追踪消息的流转
	Identification string      `json:"identification"`
	PID            string      `json:"pid"`     // 需要发送到业务服务端的需要携带pid,发送到用户端的不需要
	Payload        interface{} `json:"payload"` // 必填，根据不同的type，body的内容不同，由具体的handler解析
}

type PayloadSubscribe struct {
	Tags string `json:"tags"` // 多个tag用逗号分隔
}

func ParseSubscribe(payloadInterface interface{}) (payload *PayloadSubscribe, err error) {
	payload = &PayloadSubscribe{}
	payloadBytes, err := json.Marshal(payloadInterface)
	if err != nil {
		return
	}
	err = json.Unmarshal(payloadBytes, payload)
	return
}

type PayloadUnsubscribe = PayloadSubscribe

func ParseUnsubscribe(payloadInterface interface{}) (payload *PayloadUnsubscribe, err error) {
	payload = &PayloadUnsubscribe{}
	payloadBytes, err := json.Marshal(payloadInterface)
	if err != nil {
		return
	}
	err = json.Unmarshal(payloadBytes, payload)
	return
}

type PayLoadPush struct {
	PID  string      `json:"pid"`
	UIDs string      `json:"uids"`
	CIDs string      `json:"cids"`
	Tags string      `json:"tags"`
	Data interface{} `json:"data"`
}

func ParsePush(payloadInterface interface{}) (payload *PayLoadPush, err error) {
	payload = &PayLoadPush{}
	payloadBytes, err := json.Marshal(payloadInterface)
	if err != nil {
		return
	}
	err = json.Unmarshal(payloadBytes, payload)
	return
}

type PayloadConnect struct {
	UID string `json:"uid"`
	CID string `json:"cid"`
}

func ParseConnect(payloadInterface interface{}) (payload *PayloadConnect, err error) {
	payload = &PayloadConnect{}
	payloadBytes, err := json.Marshal(payloadInterface)
	if err != nil {
		return
	}
	err = json.Unmarshal(payloadBytes, payload)
	return
}

type PayloadDisconnect = PayloadConnect

func ParseDisConnect(payloadInterface interface{}) (payload *PayloadDisconnect, err error) {
	payload = &PayloadDisconnect{}
	payloadBytes, err := json.Marshal(payloadInterface)
	if err != nil {
		return
	}
	err = json.Unmarshal(payloadBytes, payload)
	return
}

package queuemessage

import "encoding/json"

type Type string

const (
	TypePush       Type = "push"
	TypeRequest    Type = "request"
	TypeConnect    Type = "connect"
	TypeDisconnect Type = "disconnect"
)

type Message struct {
	Type Type `json:"type"`
	//TraceID        string      `json:"trace_id"` // 追踪消息的流转
	Identification string      `json:"identification"`
	PID            string      `json:"pid"`     // 需要发送到业务服务端的需要携带pid,发送到用户端的不需要
	Payload        interface{} `json:"payload"` // 必填，根据不同的type，body的内容不同，由具体的handler解析
}

type PayLoadPush struct {
	PID  string      `json:"pid"`  // 项目id
	UIDs string      `json:"uids"` // 接收人uid
	CIDs string      `json:"cids"` // client id
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

type PayloadRequest struct {
	UID     string      `json:"uid"` // 发送人uid
	CID     string      `json:"cid"` // 发送人 client id
	Payload interface{} `json:"payload"`
}

func ParseRequest(payloadInterface interface{}) (payload *PayloadRequest, err error) {
	payload = &PayloadRequest{}
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

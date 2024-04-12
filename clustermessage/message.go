package clustermessage

import (
	"encoding/json"
	"time"
)

type Type string

const (
	TypePush       Type = "push"
	TypeRequest    Type = "request"
	TypeConnect    Type = "connect"
	TypeDisconnect Type = "disconnect"

	TypeReport Type = "report" // 用户端上报设备信息,该信息会保存到ws集群中
	TypeHeart  Type = "heart"
)

// AffairMsg 为用户端，ws集群，业务服务端之间传递的消息结构
// 整体的消息流转过程如下：
// 1. 用户端通过ws连接发送消息到ws集群
//
//	{
//	   "affair_id":"11111",
//	   "ack_id":"2222",
//	   "payload":{
//	       "operation":"request",
//	        "type":"market"
//	   }
//	}
//
// 2. ws集群收到消息后，返回{"ack_id":2222},告知用户端响应成功,然后附加部分字段后将消息发送到消息队列
//
//	{
//	   "affair_id": "11111",
//	   "source":{   // 集群附加
//	       "uid":"1",
//	       "cid":"1",
//	   },
//	   "type":"request" // 集群附加
//	   "payload": {
//	       "operation": "request",
//	       "type": "market"
//	   }
//	}
//
// 3. ws集群从消息队列中获取到消息，根据pid，确定该pid对应的业务服务端在自己的服务器上后，将消息发送到业务服务端
// 4. 业务服务端处理完消息后，将响应消息发送到ws集群，ws集群返回{"ack_id":3333},告知业务端响应成功,然后附加部分字段后将消息发送到消息队列
//
//	{
//	   "affair_id": "11111",
//	   "ack_id":"3333",
//	   "type":"push", 集群附加
//	   "to": {
//	       "pid":"1",  // 集群附加
//	       "uids": [], // 业务端根据接收人填写
//	       "cids": [], // 业务端根据接收人填写
//	   },
//	   "payload": {业务端响应内容}
//	 }
//
// 5. ws集群从消息队列中获取到消息，根据pid找到uid，然后收窄字段后发送到所有用户端
//
//	{
//	   "affair_id": "11111",
//	   "payload": {业务端响应内容}
//	 }

// 心跳消息
// 用户端发送
//
//	{
//	  "type":"heart",
//	  "ack_id":"1111",
//	  "payload": {
//		 "ping": "2021-01-01 12:00:00"
//	}
//
// }
// ws集群返回  调用方法NewHeartResp
//
//	{
//	  "type":"heart",
//	  "ack_id":"1111",
//	  "payload": {
//		 "pong": "2021-01-01 12:00:00"
//	}

type AffairMsg struct {
	AffairID string      `json:"affair_id,omitempty"` // 业务唯一id，用户发送消息时附加affair_id
	AckID    string      `json:"ack_id,omitempty" `   // ws应答唯一id，当ws集群收到消息时，会将该ackId返回给用户端，告知用户端接受成功,如果ackID为空，ws集群不会回复成功消息
	Payload  interface{} `json:"payload,omitempty"`
	Type     Type        `json:"type,omitempty"`
	Source   *Source     `json:"source,omitempty"` // 用户消息，在发送到消息队列时，需要由ws集群附加source
	To       *To         `json:"to,omitempty"`     // 需要由业务服务端附加
}

type Source struct {
	PID string `json:"pid,omitempty"`
	UID string `json:"uid,omitempty"`
	CID string `json:"cid,omitempty"`
}
type To struct {
	PID  string   `json:"pid,omitempty"`  // ws集群附加
	UIDs []string `json:"uids,omitempty"` // 业务端附加
	CIDs []string `json:"cids,omitempty"` // 业务端附加
}

func ParseAffair(bytes []byte) (message *AffairMsg, err error) {
	message = &AffairMsg{}
	err = json.Unmarshal(bytes, message)
	return
}

func PackAffair(message *AffairMsg) ([]byte, error) {
	return json.Marshal(message)
}

// AckMSg
// ws集群返回给客户端的消息
// 有两种情况
// 1. ws流中客户端请求时具有ack_id,则使用 NewAck 应答,否则不应答
// 2. 连接成功,失败或没有ack_id的消息，使用 NewSuccessResp 或 NewErrorResp 应答
type AckMsg struct {
	AckID string `json:"ack_id"`
	Msg   string `json:"msg"` // 提示信息
	Code  int    `json:"code"`
}

func ParseAck(bytes []byte) (ack *AckMsg, err error) {
	ack = &AckMsg{}
	err = json.Unmarshal(bytes, ack)
	return
}

func NewAck(ackID string) AckMsg {
	return AckMsg{
		AckID: ackID,
	}
}

func NewErrorResp(msg string) AckMsg {
	return newResp("", 0, msg)
}

func NewSuccessResp(msg ...string) AckMsg {
	if len(msg) > 0 {
		return newResp("", 1, msg[0])
	}
	return newResp("", 1, "success")
}

func newResp(ackID string, code int, msg string) AckMsg {
	return AckMsg{
		AckID: ackID,
		Msg:   msg,
		Code:  code,
	}
}

func NewHeartResp(msg *AffairMsg) AffairMsg {
	return AffairMsg{
		Type:  TypeHeart,
		AckID: msg.AckID,
		Payload: map[string]string{
			"pong": time.Now().Format("2006-01-02 15:04:05"),
		},
	}
}

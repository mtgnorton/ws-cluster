package message

type Type string

const (
	TypeSubscribe   Type = "subscribe"
	TypeUnsubscribe Type = "unsubscribe"
	TypePush        Type = "push"
)

type WsReqMessage struct {
	Type      Type        `json:"type"`       // 必填，subscribe, unsubscribe, push
	RequestID string      `json:"request_id"` // 可选，用于标识请求的唯一id,前端传递过来后，后端返回时会原样返回
	Payload   interface{} `json:"payload"`    // 必填，根据不同的type，body的内容不同，由具体的handler解析
	Metadata  interface{} `json:"metadata"`   // 可选，用于传递一些额外的信息，由具体的handler解析
}

type WsResMessage struct {
	Code      int         `json:"code"` // 200 成功，其他失败
	RequestID string      `json:"request_id"`
	Msg       string      `json:"msg"`
	Payload   interface{} `json:"payload"` // 根据不同的type，body的内容不同
}

type HttpReqMessage struct {
}

type HttpResMessage struct {
}

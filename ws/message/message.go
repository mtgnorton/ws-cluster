package message

type Type string

const (
	TypeSubscribe   Type = "subscribe"
	TypeUnsubscribe Type = "unsubscribe"
)

type Req struct {
	Type      Type        `json:"type"`       // 必填，subscribe, unsubscribe
	RequestID string      `json:"request_id"` // 可选，用于标识请求的唯一id,前端传递过来后，后端返回时会原样返回
	Payload   interface{} `json:"payload"`    // 必填，根据不同的type，body的内容不同，由具体的handler解析
	Metadata  interface{} `json:"metadata"`   // 可选，用于传递一些额外的信息，由具体的handler解析
}

type Res struct {
	Code      int         `json:"code"`       // 1 成功，0失败
	RequestID string      `json:"request_id"` // 可选，用于标识请求的唯一id,前端传递过来后，后端返回时会原样返回
	Msg       string      `json:"msg"`        // 提示信息
	Payload   interface{} `json:"payload"`    // 根据不同的type，body的内容不同
}

func NewErrorRes(msg string, requestID string) *Res {
	return &Res{
		Code:      0,
		RequestID: requestID,
		Msg:       msg,
		Payload:   struct{}{},
	}
}

func NewSuccessRes(msg string, requestID string) *Res {
	return &Res{
		Code:      1,
		RequestID: requestID,
		Msg:       msg,
		Payload:   struct{}{},
	}
}

package httpmessage

type Type string

type Req struct {
	Type    Type        `json:"type"`    // 必填，push
	Payload interface{} `json:"payload"` // 必填，根据不同的type，body的内容不同，由具体的handler解析
}

type Res struct {
	Code    int         `json:"code"` // 1 成功，0 失败
	Msg     string      `json:"msg"`
	Payload interface{} `json:"payload"`
}

func NewErrorRes(msg string) *Res {
	return &Res{
		Code:    0,
		Msg:     msg,
		Payload: struct{}{},
	}
}
func NewSuccessRes() *Res {
	return &Res{
		Code:    1,
		Msg:     "success",
		Payload: struct{}{},
	}
}

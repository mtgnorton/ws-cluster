package message

var DefaultParser WsParser = NewJsonParser()

// 即可以解析消息，又可以压缩消息

type WsParser interface {
	Parse(msg []byte) (message *WsReqMessage, err error)
}

type WsSynthesizer interface {
	Synthesize(message WsResMessage) (msg []byte, err error)
}

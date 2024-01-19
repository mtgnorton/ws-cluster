package message

var DefaultParser Parse = NewJsonProcessor()
var DefaultSynthesizer Synthesize = NewJsonProcessor()

// 即可以解析消息，又可以压缩消息

type Parse interface {
	Parse(msg []byte) (message *Req, err error)
}

type Synthesize interface {
	Synthesize(message *Res) (msg []byte, err error)
}

package wsmessage

var DefaultWsProcessor Processor = NewJsonProcessor()

type Processor interface {
	ReqEncode(message *Req) (messageBytes []byte, err error)
	ReqDecode(messageBytes []byte) (message *Req, err error)
	ResEncode(message *Res) (messageBytes []byte, err error)
	ResDecode(messageBytes []byte) (message *Res, err error)
}

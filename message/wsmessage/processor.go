package wsmessage

var DefaultProcessor Processor = NewJsonProcessor()

type Processor interface {
	Decode(messageBytes []byte) (message *Req, err error)
	Encode(message *Res) (messageBytes []byte, err error)
}

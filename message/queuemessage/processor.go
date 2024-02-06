package queuemessage

var DefaultProcessor Processor = NewJsonProcessor()

type Processor interface {
	Decode(messageBytes []byte) (message *Message, err error)
	Encode(message *Message) (messageBytes []byte, err error)
}

package parse

import "github.com/mtgnorton/ws-cluster/message"

var DefaultParser Parser = NewJsonParser()

type Parser interface {
	Parse(msg []byte) (message *message.ReqMessage, err error)
}

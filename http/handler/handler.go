package handler

var DefaultPushHandler = NewPushHandler()

type Handle interface {
	Handle(message []byte) (isAck bool)
}

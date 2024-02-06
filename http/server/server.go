package server

var DefaultHttpServer = New()

type Server interface {
	Name() string
	Init(...Option)
	Options() Options
	Run()
	Stop() error
}

package server

var DefaultWsServer = New()

type Server interface {
	Name() string
	Init(...Option)
	Options() Options
	Run()
	Stop() error
	RegisterToRegistryLoop()
}

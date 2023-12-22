package server

type Server interface {
	Name() string
	Init(...Option)
	Options() Options
	Run()
	Stop() error
}

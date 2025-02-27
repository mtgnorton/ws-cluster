package server

// var _ Server = New()

type Server interface {
	Name() string
	Init(...Option)
	Options() Options
	Run()
	Stop() error
}

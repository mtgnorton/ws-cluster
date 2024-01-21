package main

import (
	"github.com/getsentry/sentry-go"
	httpServer "github.com/mtgnorton/ws-cluster/http/server"
	"github.com/mtgnorton/ws-cluster/tools/sentry_instance"
	wsServer "github.com/mtgnorton/ws-cluster/ws/server"
)

func main() {
	err := sentry_instance.DefaultSentryInstance.Init()
	if err != nil {
		panic(err)
	}
	go httpServer.New().Run()
	sentry.CaptureMessage("It works!")

	wsServer.New().Run()
}

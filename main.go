package main

import (
	"time"

	"github.com/getsentry/sentry-go"
	httpServer "github.com/mtgnorton/ws-cluster/http/server"
)

func main() {

	defer sentry.Flush(time.Second * 3)

	httpServer.New().Run()
	//wsServer.New().Run()
}

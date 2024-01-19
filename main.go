package main

import (
	httpServer "github.com/mtgnorton/ws-cluster/http/server"
	wsServer "github.com/mtgnorton/ws-cluster/ws/server"
)

func main() {
	go httpServer.New().Run()
	wsServer.New().Run()
}

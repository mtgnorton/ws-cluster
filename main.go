package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ws-cluster/shared"

	httpServer "ws-cluster/http/server"

	"ws-cluster/ws/server"

	"ws-cluster/tools/swagger"
	"ws-cluster/tools/wsprometheus"

	"github.com/gogf/gf/v2/frame/g"
	swaggerFiles "github.com/swaggo/files"

	"github.com/gogf/gf/v2/net/ghttp"

	"ws-cluster/config"
	"ws-cluster/docs"

	"github.com/getsentry/sentry-go"
	"github.com/sasha-s/go-deadlock"
)

//	@title			Ws-cluster API
//	@version		1.0
//	@description	包含ws连接、消息发送、消息接收、消息推送等接口
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	mtgnorton

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

func main() {

	c := config.DefaultConfig
	if c.Values().Env == config.Prod {
		deadlock.Opts.Disable = true
	}
	shared.InitRedis(c)
	shared.InitIP()

	shared.GetNodeID(c)
	toolServer(c)
	defer sentry.Flush(time.Second * 3)

	go httpServer.DefaultHttpServer.Run()

	go server.DefaultWsServer.Run()
	// 程序退出信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	fmt.Println("正在关闭服务...")

	if shared.NodeIDWorker != nil {
		shared.NodeIDWorker.Release()
	}
	// 关闭HTTP服务器
	if err := httpServer.DefaultHttpServer.Stop(); err != nil {
		fmt.Printf("HTTP服务器关闭失败: %v\n", err)
	}

	// 关闭WebSocket服务器
	if err := server.DefaultWsServer.Stop(); err != nil {
		fmt.Printf("WebSocket服务器关闭失败: %v\n", err)
	}

	fmt.Println("服务已安全关闭")
}

func toolServer(c config.Config) {
	if c.Values().Pprof.Enable {
		go ghttp.StartPProfServer(c.Values().Pprof.Port)
	}
	if c.Values().Swagger.Enable {
		docs.SwaggerInfo.Host = "localhost:" + fmt.Sprintf("%d", c.Values().HttpServer.Port)
		docs.SwaggerInfo.BasePath = "/v1"
		s := g.Server("swagger")
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.GET(fmt.Sprintf("%s/*any", c.Values().Swagger.Path), swagger.WrapHandler(swaggerFiles.Handler))
		})
		s.SetPort(c.Values().Swagger.Port)
		go s.Run()
	}

	if c.Values().Prometheus.Enable {
		wsprometheus.DefaultPrometheus.Init()
	}

}

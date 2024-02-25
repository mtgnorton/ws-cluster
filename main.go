package main

import (
	"fmt"
	"time"

	"github.com/mtgnorton/ws-cluster/shared"

	server2 "github.com/mtgnorton/ws-cluster/http/server"

	"github.com/mtgnorton/ws-cluster/ws/server"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/mtgnorton/ws-cluster/tools/swagger"
	"github.com/mtgnorton/ws-cluster/tools/wsprometheus"
	swaggerFiles "github.com/swaggo/files"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/getsentry/sentry-go"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/docs"
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
	shared.InitSnowflakeRedisJwt(c)
	toolServer(c)
	defer sentry.Flush(time.Second * 3)

	go server2.DefaultHttpServer.Run()

	server.DefaultWsServer.Run()
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

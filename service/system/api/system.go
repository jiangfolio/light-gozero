package main

import (
	"flag"
	"fmt"

	"github.com/jiangbyte/light-gozero/common/response"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/config"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/handler"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/middleware"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/system.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 1. 注册统一响应包络
	response.Setup()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 2. 全路由写入客户端 IP/UA；公开鉴权路由再套周期限流
	server.Use(middleware.ClientMeta(c.TrustProxy))
	if ctx.RateLimit != nil {
		server.Use(ctx.RateLimit)
	}

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

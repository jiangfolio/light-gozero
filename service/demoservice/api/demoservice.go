// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"github.com/jiangbyte/light-gozero/common/response"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/config"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/handler"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/demoservice.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 1. 注册统一响应 Ok/Error Handler（含路由 404/405），必须在处理请求前完成
	response.Setup()

	server := rest.MustNewServer(c.RestConf, response.RestOptions()...)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	// 2. 注册分片 PUT / 公开直链等原始二进制路由
	handler.RegisterRawFileHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

package handler

import (
	"net/http"

	file "github.com/jiangbyte/light-gozero/service/demoservice/api/internal/handler/file"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterRawFileHandlers 注册二进制分片上传与公开直链下载（非 goctl JSON 路由）。
func RegisterRawFileHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPut,
				Path:    "/file/upload/part",
				Handler: file.FilePutPartHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/file/public",
				Handler: file.FilePublicHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/file/download",
				Handler: file.FileDownloadHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}

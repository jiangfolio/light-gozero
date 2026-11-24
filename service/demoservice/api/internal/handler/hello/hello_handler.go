// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/logic/hello"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HelloHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HelloReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := hello.NewHelloLogic(r.Context(), svcCtx)
		resp, err := l.Hello(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

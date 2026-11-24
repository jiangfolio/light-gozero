// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/config"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ConfigPublicHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := config.NewConfigPublicLogic(r.Context(), svcCtx)
		resp, err := l.ConfigPublic()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

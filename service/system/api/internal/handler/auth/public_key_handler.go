// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/auth"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PublicKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth.NewPublicKeyLogic(r.Context(), svcCtx)
		resp, err := l.PublicKey()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package shortcut

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/shortcut"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ShortcutCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShortcutCreateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := shortcut.NewShortcutCreateLogic(r.Context(), svcCtx)
		resp, err := l.ShortcutCreate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

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

func ShortcutDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IdsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := shortcut.NewShortcutDeleteLogic(r.Context(), svcCtx)
		err := l.ShortcutDelete(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

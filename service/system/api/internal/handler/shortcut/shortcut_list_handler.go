// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package shortcut

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/shortcut"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 拉当前用户快捷入口；登录即可，结果集有上限
func ShortcutListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := shortcut.NewShortcutListLogic(r.Context(), svcCtx)
		resp, err := l.ShortcutList()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

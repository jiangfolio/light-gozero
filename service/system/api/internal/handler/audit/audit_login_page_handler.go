// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/audit"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuditLoginPageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuditLoginPageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := audit.NewAuditLoginPageLogic(r.Context(), svcCtx)
		resp, err := l.AuditLoginPage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

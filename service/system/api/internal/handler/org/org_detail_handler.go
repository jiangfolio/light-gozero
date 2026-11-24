// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package org

import (
	"net/http"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/org"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func OrgDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := org.NewOrgDetailLogic(r.Context(), svcCtx)
		resp, err := l.OrgDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

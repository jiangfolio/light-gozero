package org

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type OrgPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgPageLogic {
	return &OrgPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *OrgPageLogic) OrgPage(req *types.OrgPageReq) (resp *types.OrgPageResp, err error) {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemOrgPage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	var status int64
	hasStatus := false
	if req.Status != "" {
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}

	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	scopeOrgIds, selfOrgId, empty := datascope.ForOrgPage(scope, sess.OrgId)
	if empty {
		return &types.OrgPageResp{
			Items: []types.OrgItem{}, Page: convert.Int64ToString(int64(page)),
			PageSize: convert.Int64ToString(int64(pageSize)), Total: "0",
		}, nil
	}

	list, total, err := l.svcCtx.Repo.Org.Page(l.ctx, req.Name, req.Code, status, hasStatus, scopeOrgIds, selfOrgId, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.OrgItem, 0, len(list))
	for i := range list {
		items = append(items, *toOrgItem(&list[i]))
	}
	return &types.OrgPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}

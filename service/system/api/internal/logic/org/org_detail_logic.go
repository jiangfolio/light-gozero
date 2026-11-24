package org

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	orgmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/org"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type OrgDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgDetailLogic {
	return &OrgDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *OrgDetailLogic) OrgDetail(req *types.OrgDetailReq) (resp *types.OrgItem, err error) {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemOrgDetail); err != nil {
		return nil, err
	}
	org, err := l.svcCtx.Repo.Org.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "组织不存在")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	if err := datascope.AllowOrg(scope, sess.OrgId, org.Id); err != nil {
		return nil, err
	}
	return toOrgItem(org), nil
}

func toOrgItem(org *orgmodel.SysOrg) *types.OrgItem {
	return &types.OrgItem{
		Id: org.Id, ParentId: org.ParentId, Name: org.Name, Code: org.Code,
		Sort: convert.Int64ToString(org.Sort), Status: convert.Int64ToString(org.Status),
		CreatedBy: org.CreatedBy, UpdatedBy: org.UpdatedBy,
		CreatedAt: convert.TimeToString(org.CreatedAt), UpdatedAt: convert.TimeToString(org.UpdatedAt),
	}
}

package org

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type OrgDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgDeleteLogic {
	return &OrgDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *OrgDeleteLogic) OrgDelete(req *types.IdsReq) error {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemOrgDelete); err != nil {
		return err
	}
	ids, err := batchutil.CapRequestIDs(req.Ids)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errorx.New(400, "缺少 ids")
	}
	orgs, err := l.svcCtx.Repo.Org.ListByIds(l.ctx, ids)
	if err != nil {
		return err
	}
	if len(orgs) != len(ids) {
		return errorx.New(404, "组织不存在")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	for i := range orgs {
		if err := datascope.AllowOrg(scope, sess.OrgId, orgs[i].Id); err != nil {
			return err
		}
	}
	n, err := l.svcCtx.Repo.Org.CountChildren(l.ctx, ids)
	if err != nil {
		return err
	}
	if n > 0 {
		return errorx.New(400, "组织存在子节点，无法删除")
	}
	u, err := l.svcCtx.Repo.User.CountByOrgIds(l.ctx, ids)
	if err != nil {
		return err
	}
	if u > 0 {
		return errorx.New(400, "组织下仍有用户，无法删除")
	}
	if err := l.svcCtx.Repo.Org.Delete(l.ctx, ids); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "删除组织 "+convert.Int64ToString(int64(len(ids)))+" 个")
	return nil
}

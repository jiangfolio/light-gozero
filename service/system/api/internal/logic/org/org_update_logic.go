package org

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type OrgUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgUpdateLogic {
	return &OrgUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *OrgUpdateLogic) OrgUpdate(req *types.OrgUpdateReq) error {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemOrgUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	org, err := l.svcCtx.Repo.Org.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "组织不存在")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	if err := datascope.AllowOrg(scope, sess.OrgId, org.Id); err != nil {
		return err
	}
	if req.ParentId != "" {
		if req.ParentId == req.Id {
			return errorx.New(400, "上级不能是自身")
		}
		if _, err := l.svcCtx.Repo.Org.FindById(l.ctx, req.ParentId); err != nil {
			return errorx.New(400, "上级组织不存在")
		}
		if err := datascope.AllowOrg(scope, sess.OrgId, req.ParentId); err != nil {
			return err
		}
		org.ParentId = req.ParentId
	}
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Code != "" {
		org.Code = req.Code
	}
	if req.Sort != "" {
		v, err := convert.ParseSort(req.Sort, org.Sort)
		if err != nil {
			return err
		}
		org.Sort = v
	}
	if req.Status != "" {
		v, err := convert.ParseStatus(req.Status, org.Status)
		if err != nil {
			return err
		}
		org.Status = v
	}
	org.UpdatedBy = uid
	if err := l.svcCtx.Repo.Org.Save(l.ctx, org); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "更新组织「"+org.Name+"」")
	return nil
}

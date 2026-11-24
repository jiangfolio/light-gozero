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
	orgmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/org"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type OrgCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgCreateLogic {
	return &OrgCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *OrgCreateLogic) OrgCreate(req *types.OrgCreateReq) (resp *types.IdResp, err error) {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemOrgCreate); err != nil {
		return nil, err
	}
	if req.Name == "" || req.Code == "" {
		return nil, errorx.New(400, "名称和编码不能为空")
	}
	parentId := req.ParentId
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	if parentId != "" {
		if _, err := l.svcCtx.Repo.Org.FindById(l.ctx, parentId); err != nil {
			return nil, errorx.New(400, "上级组织不存在")
		}
		if err := datascope.AllowOrg(scope, sess.OrgId, parentId); err != nil {
			return nil, err
		}
	} else if !scope.All {
		// 非全局不可创建顶层组织
		return nil, errorx.New(403, "无数据权限")
	}
	sort, err := convert.ParseSort(req.Sort, 0)
	if err != nil {
		return nil, err
	}
	status, err := convert.ParseStatus(req.Status, 1)
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	err = l.svcCtx.Repo.Org.Create(l.ctx, &orgmodel.SysOrg{
		Id: id, ParentId: parentId, Name: req.Name, Code: req.Code, Sort: sort, Status: status,
		CreatedBy: uid, UpdatedBy: uid,
	})
	if err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建组织「"+req.Name+"」")
	return &types.IdResp{Id: id}, nil
}

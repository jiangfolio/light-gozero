package role

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	rolemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/role"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type RoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleUpdateLogic {
	return &RoleUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RoleUpdateLogic) RoleUpdate(req *types.RoleUpdateReq) error {
	// 1. 校验权限与角色
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRoleUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	role, err := l.svcCtx.Repo.Role.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "角色不存在")
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	if req.Remark != "" {
		role.Remark = req.Remark
	}
	if req.Status != "" {
		v, err := convert.ParseStatus(req.Status, role.Status)
		if err != nil {
			return err
		}
		role.Status = v
	}
	scopeChanged := false
	if req.DataScope != "" {
		ds, err := parseDataScope(req.DataScope, role.DataScope)
		if err != nil {
			return err
		}
		if ds != role.DataScope {
			scopeChanged = true
		}
		role.DataScope = ds
	}
	replaceOrgs := req.OrgIds != nil || scopeChanged
	var cappedOrgIds []string
	if req.OrgIds != nil {
		cappedOrgIds, err = batchutil.CapRequestIDs(req.OrgIds)
		if err != nil {
			return err
		}
	}
	if role.DataScope == rolemodel.DataScopeCustom && replaceOrgs && len(cappedOrgIds) == 0 && req.OrgIds != nil {
		return errorx.New(400, "指定部门范围须选择组织")
	}
	op := authctx.UserId(l.ctx)
	role.UpdatedBy = op

	var orgRows []rolemodel.RelRoleOrg
	if replaceOrgs {
		ids := cappedOrgIds
		if role.DataScope != rolemodel.DataScopeCustom {
			ids = nil
		}
		orgRows, err = buildRoleOrgRows(role.Id, op, ids, l.svcCtx.IdGen.Next)
		if err != nil {
			return err
		}
	}

	// 2. 短事务落库；范围变更后刷对应用户缓存
	users, err := l.svcCtx.Repo.User.ListUserIdsByRoles(l.ctx, []string{role.Id})
	if err != nil {
		return err
	}
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		if err := tx.Role.Save(l.ctx, role); err != nil {
			return err
		}
		if replaceOrgs {
			return tx.Role.ReplaceOrgs(l.ctx, role.Id, orgRows)
		}
		return nil
	}); err != nil {
		return err
	}
	if scopeChanged || replaceOrgs {
		l.svcCtx.RbacGuard.InvalidateUsers(l.ctx, users)
	}
	operlog.SetSummary(l.ctx, "更新角色「"+role.Name+"」")
	return nil
}

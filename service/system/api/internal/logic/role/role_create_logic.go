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

type RoleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleCreateLogic {
	return &RoleCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RoleCreateLogic) RoleCreate(req *types.RoleCreateReq) (resp *types.IdResp, err error) {
	// 1. 校验权限与入参
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRoleCreate); err != nil {
		return nil, err
	}
	if req.Name == "" || req.Code == "" {
		return nil, errorx.New(400, "名称和编码不能为空")
	}
	status, err := convert.ParseStatus(req.Status, 1)
	if err != nil {
		return nil, err
	}
	dataScope, err := parseDataScope(req.DataScope, rolemodel.DataScopeAll)
	if err != nil {
		return nil, err
	}
	orgIds, err := batchutil.CapRequestIDs(req.OrgIds)
	if err != nil {
		return nil, err
	}
	if dataScope == rolemodel.DataScopeCustom && len(orgIds) == 0 {
		return nil, errorx.New(400, "指定部门范围须选择组织")
	}
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	op := authctx.UserId(l.ctx)
	orgRows, err := buildRoleOrgRows(id, op, orgIds, l.svcCtx.IdGen.Next)
	if err != nil {
		return nil, err
	}
	if dataScope != rolemodel.DataScopeCustom {
		orgRows = nil
	}

	// 2. 短事务：角色 + 指定部门
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		if err := tx.Role.Create(l.ctx, &rolemodel.SysRole{
			Id: id, Name: req.Name, Code: req.Code, DataScope: dataScope, Remark: req.Remark, Status: status,
			CreatedBy: op, UpdatedBy: op,
		}); err != nil {
			return err
		}
		if len(orgRows) > 0 {
			return tx.Role.ReplaceOrgs(l.ctx, id, orgRows)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建角色「"+req.Name+"」")
	return &types.IdResp{Id: id}, nil
}

package role

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type RoleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleDetailLogic {
	return &RoleDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RoleDetailLogic) RoleDetail(req *types.RoleDetailReq) (resp *types.RoleItem, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRoleDetail); err != nil {
		return nil, err
	}
	role, err := l.svcCtx.Repo.Role.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "角色不存在")
	}
	orgIds, err := l.svcCtx.Repo.Role.ListOrgIdsByRoleId(l.ctx, role.Id)
	if err != nil {
		return nil, err
	}
	resourceIds, err := l.svcCtx.Repo.Role.ListResourceIdsByRoleId(l.ctx, role.Id)
	if err != nil {
		return nil, err
	}
	return toRoleItem(role, orgIds, resourceIds), nil
}

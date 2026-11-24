package role

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type RoleUnbindResourcesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleUnbindResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleUnbindResourcesLogic {
	return &RoleUnbindResourcesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RoleUnbindResourcesLogic) RoleUnbindResources(req *types.RoleBindResourcesReq) error {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRoleUnbindResources); err != nil {
		return err
	}
	resourceIds, err := batchutil.CapRequestIDs(req.ResourceIds)
	if err != nil {
		return err
	}
	if req.RoleId == "" || len(resourceIds) == 0 {
		return errorx.New(400, "角色和资源不能为空")
	}
	// 2. 批量解绑资源
	if err := l.svcCtx.Repo.Role.UnbindResources(l.ctx, req.RoleId, resourceIds); err != nil {
		return err
	}
	// 3. 刷受影响用户缓存
	users, err := l.svcCtx.Repo.User.ListUserIdsByRoles(l.ctx, []string{req.RoleId})
	if err != nil {
		return err
	}
	l.svcCtx.RbacGuard.InvalidateUsers(l.ctx, users)
	return nil
}

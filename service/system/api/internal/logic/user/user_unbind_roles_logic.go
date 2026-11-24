package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserUnbindRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUnbindRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUnbindRolesLogic {
	return &UserUnbindRolesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserUnbindRolesLogic) UserUnbindRoles(req *types.UserBindRolesReq) error {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserUnbindRoles); err != nil {
		return err
	}
	roleIds, err := batchutil.CapRequestIDs(req.RoleIds)
	if err != nil {
		return err
	}
	if req.UserId == "" || len(roleIds) == 0 {
		return errorx.New(400, "用户和角色不能为空")
	}
	user, err := l.svcCtx.Repo.User.FindById(l.ctx, req.UserId)
	if err != nil {
		return errorx.New(404, "用户不存在")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return err
	}
	if err := datascope.AllowUser(scope, uid, user.Id, user.OrgId); err != nil {
		return err
	}
	if err := l.svcCtx.Repo.User.UnbindRoles(l.ctx, req.UserId, roleIds); err != nil {
		return err
	}
	l.svcCtx.RbacGuard.InvalidateUser(l.ctx, req.UserId)
	operlog.SetSummary(l.ctx, "用户解绑角色")
	return nil
}

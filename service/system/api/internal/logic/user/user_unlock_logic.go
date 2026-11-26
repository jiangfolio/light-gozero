package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserUnlockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUnlockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUnlockLogic {
	return &UserUnlockLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// UserUnlock 清除登录失败锁定（失败次数与锁定截止时间）。
func (l *UserUnlockLogic) UserUnlock(req *types.UserUnlockReq) error {
	// 1. 权限与入参
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	// 2. 加载用户并校验数据范围
	user, err := l.svcCtx.Repo.User.FindById(l.ctx, req.Id)
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
	// 3. 清锁定状态
	user.FailCount = 0
	user.LockedUntil = nil
	user.UpdatedBy = uid
	if err := l.svcCtx.Repo.User.Save(l.ctx, user); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "解锁用户「"+user.Username+"」")
	return nil
}

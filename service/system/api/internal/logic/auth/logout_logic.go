package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Logout 注销当前 token。
func (l *LogoutLogic) Logout() error {
	s, ok := authctx.SessionFrom(l.ctx)
	if !ok {
		return errorx.New(401, "未登录或登录已失效")
	}
	// 1. 强制改密场景允许注销；否则需 logout 权限
	if !authctx.MustChangePassword(l.ctx) {
		if err := l.svcCtx.RbacGuard.Require(l.ctx, s.UserId, perm.SystemAuthLogout); err != nil {
			return err
		}
	}
	if err := l.svcCtx.SessionStore.Delete(l.ctx, s.Token, s.UserId); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "注销登录")
	return nil
}

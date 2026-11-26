package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/reqmeta"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/audit"
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

// Logout 注销当前 token；登录即可，不走角色授权。同时写入登录日志（登出）。
func (l *LogoutLogic) Logout() error {
	// 1. 校验会话并删除 token
	s, ok := authctx.SessionFrom(l.ctx)
	if !ok {
		return errorx.New(401, "未登录或登录已失效")
	}
	if err := l.svcCtx.SessionStore.Delete(l.ctx, s.Token, s.UserId); err != nil {
		return err
	}
	// 2. 写入登录/登出审计，供网络安全「登录日志」查询
	l.writeLogoutLog(s.UserId, s.Username)
	operlog.SetSummary(l.ctx, "注销登录")
	return nil
}

func (l *LogoutLogic) writeLogoutLog(userId, username string) {
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return
	}
	row := &audit.LoginLog{
		Id: id, UserId: userId, Username: username,
		Ip: reqmeta.ClientIP(l.ctx), UserAgent: reqmeta.UserAgent(l.ctx),
		Result: "logout", Message: "退出登录",
	}
	if err := l.svcCtx.Repo.Audit.CreateLoginLog(l.ctx, row); err != nil {
		l.Errorf("write logout log: %v", err)
	}
}

package auth

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/passwd"
	"github.com/jiangbyte/light-gozero/common/reqmeta"
	"github.com/jiangbyte/light-gozero/common/session"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/audit"
	usermodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/user"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Login 验证码 + 锁定策略 + 密码校验后签发会话。
func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	ip := reqmeta.ClientIP(l.ctx)
	ua := reqmeta.UserAgent(l.ctx)

	// 1. 校验图形验证码（一次性）
	if !l.svcCtx.CaptchaStore.Verify(l.ctx, req.CaptchaId, req.CaptchaCode) {
		l.writeLoginLog("", req.Username, "fail", "验证码错误", ip, ua)
		return nil, errorx.New(400, "验证码错误")
	}

	// 2. 解密密码密文
	plainPwd, err := l.svcCtx.PwdCipher.Decrypt(req.Password)
	if err != nil {
		l.writeLoginLog("", req.Username, "fail", "密码密文无效", ip, ua)
		return nil, errorx.New(400, "密码密文无效")
	}

	// 3. 查用户；不存在也统一文案，避免枚举
	user, err := l.svcCtx.Repo.User.FindByUsername(l.ctx, req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.writeLoginLog("", req.Username, "fail", "用户名或密码错误", ip, ua)
			return nil, errorx.New(401, "用户名或密码错误")
		}
		return nil, err
	}
	if user.Status != 1 {
		l.writeLoginLog(user.Id, user.Username, "fail", "账号已停用", ip, ua)
		return nil, errorx.New(401, "用户名或密码错误")
	}

	// 4. 锁定检查（对外统一文案，防账号枚举）
	now := time.Now()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		l.writeLoginLog(user.Id, user.Username, "locked", "账号已锁定", ip, ua)
		return nil, errorx.New(401, "用户名或密码错误")
	}

	// 5. 校验密码；失败累计并可能锁定
	if !passwd.Match(user.PasswordHash, plainPwd) {
		locked, err := l.onPasswordFail(user, now)
		if err != nil {
			return nil, err
		}
		if locked {
			l.writeLoginLog(user.Id, user.Username, "locked", "连续登录失败，账号已锁定", ip, ua)
		} else {
			l.writeLoginLog(user.Id, user.Username, "fail", "用户名或密码错误", ip, ua)
		}
		return nil, errorx.New(401, "用户名或密码错误")
	}

	// 6. 成功：清失败计数，判断强制改密/过期
	user.FailCount = 0
	user.LockedUntil = nil
	mustChange := user.MustChangePassword == 1 || l.passwordExpired(user, now)
	if mustChange {
		user.MustChangePassword = 1
	}
	_ = l.svcCtx.Repo.User.Save(l.ctx, user)

	token, err := l.svcCtx.SessionStore.Create(l.ctx, session.Info{
		UserId: user.Id, Username: user.Username, OrgId: user.OrgId,
		MustChangePassword: mustChange,
	})
	if err != nil {
		return nil, err
	}
	l.writeLoginLog(user.Id, user.Username, "success", "登录成功", ip, ua)

	flag := "0"
	if mustChange {
		flag = "1"
	}
	return &types.LoginResp{
		Token:              token,
		MustChangePassword: flag,
		User: types.LoginUser{
			Id: user.Id, Username: user.Username, Nickname: user.Nickname, OrgId: user.OrgId,
		},
	}, nil
}

// onPasswordFail 累计失败次数；达到阈值则锁定并返回 locked=true。
func (l *LoginLogic) onPasswordFail(user *usermodel.SysUser, now time.Time) (locked bool, err error) {
	maxFail := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyMaxFailAttempts, appcfg.DefaultMaxFailAttempts)
	if maxFail <= 0 {
		maxFail = appcfg.DefaultMaxFailAttempts
	}
	lockMin := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyLockMinutes, appcfg.DefaultLockMinutes)
	if lockMin <= 0 {
		lockMin = appcfg.DefaultLockMinutes
	}
	user.FailCount++
	if int(user.FailCount) >= maxFail {
		until := now.Add(time.Duration(lockMin) * time.Minute)
		user.LockedUntil = &until
		locked = true
	}
	return locked, l.svcCtx.Repo.User.Save(l.ctx, user)
}

func (l *LoginLogic) passwordExpired(user *usermodel.SysUser, now time.Time) bool {
	days := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordExpireDays, appcfg.DefaultPasswordExpireDays)
	if days <= 0 {
		return false
	}
	if user.PasswordChangedAt.IsZero() {
		return true
	}
	return now.After(user.PasswordChangedAt.AddDate(0, 0, days))
}

func (l *LoginLogic) writeLoginLog(userId, username, result, message, ip, ua string) {
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return
	}
	row := &audit.LoginLog{
		Id: id, UserId: userId, Username: username, Ip: ip, UserAgent: ua,
		Result: result, Message: message,
	}
	if err := l.svcCtx.Repo.Audit.CreateLoginLog(l.ctx, row); err != nil {
		l.Errorf("write login log: %v", err)
	}
}

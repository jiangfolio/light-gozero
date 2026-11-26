package auth

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/passwd"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	usermodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/user"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ChangePassword 自助改密：登录即可，校验旧密、策略与历史，成功后吊销全部会话。
func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) error {
	// 1. 仅要求登录态（账号共有能力，不走角色授权）
	uid := authctx.UserId(l.ctx)
	if uid == "" {
		return errorx.New(401, "未登录或登录已失效")
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return errorx.New(400, "旧密码和新密码不能为空")
	}

	// 2. 解密密文
	oldPlain, err := l.svcCtx.PwdCipher.Decrypt(req.OldPassword)
	if err != nil {
		return errorx.New(400, "密码密文无效")
	}
	newPlain, err := l.svcCtx.PwdCipher.Decrypt(req.NewPassword)
	if err != nil {
		return errorx.New(400, "密码密文无效")
	}
	if oldPlain == newPlain {
		return errorx.New(400, "新密码不能与旧密码相同")
	}

	// 3. 最小长度 + 可选复杂度
	minLen := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordMinLen, appcfg.DefaultPasswordMinLen)
	complexityOn := l.svcCtx.AppConfig.GetBool(l.ctx, appcfg.KeyPasswordComplexityEnabled, appcfg.DefaultPasswordComplexityEnabled)
	if err := passwd.ValidateComplexity(newPlain, minLen, complexityOn); err != nil {
		return err
	}

	user, err := l.svcCtx.Repo.User.FindById(l.ctx, uid)
	if err != nil {
		return errorx.New(404, "用户不存在")
	}
	if !passwd.Match(user.PasswordHash, oldPlain) {
		return errorx.New(400, "旧密码不正确")
	}

	// 4. 历史密码（含当前）
	histN := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordHistoryCount, appcfg.DefaultPasswordHistoryCount)
	if histN <= 0 {
		histN = appcfg.DefaultPasswordHistoryCount
	}
	hashes, err := l.svcCtx.Repo.User.ListRecentPasswordHashes(l.ctx, uid, histN)
	if err != nil {
		return err
	}
	check := append([]string{user.PasswordHash}, hashes...)
	if passwd.InHistory(newPlain, check) {
		return errorx.New(400, "不能使用最近用过的密码")
	}

	newHash, err := passwd.Hash(newPlain)
	if err != nil {
		return err
	}
	histId, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return err
	}
	oldHash := user.PasswordHash
	now := time.Now()
	user.PasswordHash = newHash
	user.PasswordChangedAt = now
	user.MustChangePassword = 0
	user.FailCount = 0
	user.LockedUntil = nil
	user.UpdatedBy = uid

	// 5. 短事务：历史 + 更新用户
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		if err := tx.User.AddPasswordHistory(l.ctx, &usermodel.PasswordHistory{
			Id: histId, UserId: uid, PasswordHash: oldHash,
		}); err != nil {
			return err
		}
		if err := tx.User.Save(l.ctx, user); err != nil {
			return err
		}
		return tx.User.TrimPasswordHistory(l.ctx, uid, histN)
	}); err != nil {
		return err
	}

	// 6. 事务外吊销全部会话
	_ = l.svcCtx.SessionStore.DeleteAllByUser(l.ctx, uid)
	operlog.SetSummary(l.ctx, "修改登录密码")
	return nil
}

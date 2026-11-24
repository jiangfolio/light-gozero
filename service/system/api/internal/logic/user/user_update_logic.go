package user

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/passwd"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	usermodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/user"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateLogic {
	return &UserUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserUpdateLogic) UserUpdate(req *types.UserUpdateReq) error {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
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

	kickSessions := false
	passwordChanged := false
	orgChanged := false
	var oldHash, newHash, histId string

	if req.OrgId != "" {
		if _, err := l.svcCtx.Repo.Org.FindById(l.ctx, req.OrgId); err != nil {
			return errorx.New(400, "组织不存在")
		}
		sess, _ := authctx.SessionFrom(l.ctx)
		if err := datascope.AllowOrg(scope, sess.OrgId, req.OrgId); err != nil {
			return err
		}
		if req.OrgId != user.OrgId {
			orgChanged = true
		}
		user.OrgId = req.OrgId
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Password != "" {
		plainPwd, err := l.svcCtx.PwdCipher.Decrypt(req.Password)
		if err != nil {
			return errorx.New(400, "密码密文无效")
		}
		minLen := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordMinLen, appcfg.DefaultPasswordMinLen)
		complexityOn := l.svcCtx.AppConfig.GetBool(l.ctx, appcfg.KeyPasswordComplexityEnabled, appcfg.DefaultPasswordComplexityEnabled)
		if err := passwd.ValidateComplexity(plainPwd, minLen, complexityOn); err != nil {
			return err
		}
		histN := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordHistoryCount, appcfg.DefaultPasswordHistoryCount)
		if histN <= 0 {
			histN = appcfg.DefaultPasswordHistoryCount
		}
		hashes, err := l.svcCtx.Repo.User.ListRecentPasswordHashes(l.ctx, user.Id, histN)
		if err != nil {
			return err
		}
		if passwd.InHistory(plainPwd, append([]string{user.PasswordHash}, hashes...)) {
			return errorx.New(400, "不能使用最近用过的密码")
		}
		newHash, err = passwd.Hash(plainPwd)
		if err != nil {
			return err
		}
		histId, err = l.svcCtx.IdGen.Next()
		if err != nil {
			return err
		}
		oldHash = user.PasswordHash
		user.PasswordHash = newHash
		user.PasswordChangedAt = time.Now()
		user.MustChangePassword = 1
		user.FailCount = 0
		user.LockedUntil = nil
		passwordChanged = true
		kickSessions = true
	}
	if req.Status != "" {
		v, err := convert.ParseStatus(req.Status, user.Status)
		if err != nil {
			return err
		}
		if v == 1 {
			user.FailCount = 0
			user.LockedUntil = nil
		}
		if v == 0 {
			kickSessions = true
		}
		user.Status = v
	}
	user.UpdatedBy = authctx.UserId(l.ctx)

	// 2. 落库（改密时带历史）
	if passwordChanged {
		histN := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordHistoryCount, appcfg.DefaultPasswordHistoryCount)
		if histN <= 0 {
			histN = appcfg.DefaultPasswordHistoryCount
		}
		if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
			if err := tx.User.AddPasswordHistory(l.ctx, &usermodel.PasswordHistory{
				Id: histId, UserId: user.Id, PasswordHash: oldHash,
			}); err != nil {
				return err
			}
			if err := tx.User.Save(l.ctx, user); err != nil {
				return err
			}
			return tx.User.TrimPasswordHistory(l.ctx, user.Id, histN)
		}); err != nil {
			return err
		}
	} else if err := l.svcCtx.Repo.User.Save(l.ctx, user); err != nil {
		return err
	}

	// 3. 事务外踢会话；组织变更刷新数据范围缓存
	if kickSessions {
		_ = l.svcCtx.SessionStore.DeleteAllByUser(l.ctx, user.Id)
	}
	if orgChanged {
		l.svcCtx.RbacGuard.InvalidateUser(l.ctx, user.Id)
	}
	operlog.SetSummary(l.ctx, "更新用户「"+user.Username+"」")
	return nil
}

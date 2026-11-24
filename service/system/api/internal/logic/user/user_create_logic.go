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
	usermodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/user"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCreateLogic {
	return &UserCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserCreateLogic) UserCreate(req *types.UserCreateReq) (resp *types.IdResp, err error) {
	// 1. 校验操作权限与数据范围
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserCreate); err != nil {
		return nil, err
	}
	if req.Username == "" || req.Password == "" || req.OrgId == "" {
		return nil, errorx.New(400, "组织、用户名和密码不能为空")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	if err := datascope.AllowOrg(scope, sess.OrgId, req.OrgId); err != nil {
		return nil, err
	}
	plainPwd, err := l.svcCtx.PwdCipher.Decrypt(req.Password)
	if err != nil {
		return nil, errorx.New(400, "密码密文无效")
	}
	minLen := l.svcCtx.AppConfig.GetInt(l.ctx, appcfg.KeyPasswordMinLen, appcfg.DefaultPasswordMinLen)
	complexityOn := l.svcCtx.AppConfig.GetBool(l.ctx, appcfg.KeyPasswordComplexityEnabled, appcfg.DefaultPasswordComplexityEnabled)
	if err := passwd.ValidateComplexity(plainPwd, minLen, complexityOn); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.Repo.Org.FindById(l.ctx, req.OrgId); err != nil {
		return nil, errorx.New(400, "组织不存在")
	}
	status, err := convert.ParseStatus(req.Status, 1)
	if err != nil {
		return nil, err
	}
	hash, err := passwd.Hash(plainPwd)
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	op := authctx.UserId(l.ctx)
	now := time.Now()
	err = l.svcCtx.Repo.User.Create(l.ctx, &usermodel.SysUser{
		Id: id, OrgId: req.OrgId, Username: req.Username, PasswordHash: hash,
		Nickname: req.Nickname, Status: status, PasswordChangedAt: now, MustChangePassword: 1,
		CreatedBy: op, UpdatedBy: op,
	})
	if err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建用户「"+req.Username+"」")
	return &types.IdResp{Id: id}, nil
}

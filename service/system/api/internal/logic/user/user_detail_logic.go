package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	usermodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/user"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDetailLogic {
	return &UserDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserDetailLogic) UserDetail(req *types.UserDetailReq) (resp *types.UserItem, err error) {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserDetail); err != nil {
		return nil, err
	}
	user, err := l.svcCtx.Repo.User.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "用户不存在")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	if err := datascope.AllowUser(scope, uid, user.Id, user.OrgId); err != nil {
		return nil, err
	}
	return toUserItem(user), nil
}

func toUserItem(u *usermodel.SysUser) *types.UserItem {
	return &types.UserItem{
		Id: u.Id, OrgId: u.OrgId, Username: u.Username, Nickname: u.Nickname,
		Status: convert.Int64ToString(u.Status), CreatedBy: u.CreatedBy, UpdatedBy: u.UpdatedBy,
		CreatedAt: convert.TimeToString(u.CreatedAt), UpdatedAt: convert.TimeToString(u.UpdatedAt),
	}
}

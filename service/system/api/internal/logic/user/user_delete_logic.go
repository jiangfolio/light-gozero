package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDeleteLogic {
	return &UserDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserDeleteLogic) UserDelete(req *types.IdsReq) error {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserDelete); err != nil {
		return err
	}
	ids, err := batchutil.CapRequestIDs(req.Ids)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errorx.New(400, "缺少 ids")
	}
	// 1. 批量加载并校验数据范围
	users, err := l.svcCtx.Repo.User.ListByIds(l.ctx, ids)
	if err != nil {
		return err
	}
	if len(users) != len(ids) {
		return errorx.New(404, "用户不存在")
	}
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return err
	}
	for i := range users {
		if err := datascope.AllowUser(scope, uid, users[i].Id, users[i].OrgId); err != nil {
			return err
		}
	}
	// 2. 短事务：清关联、密码历史后删用户
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		if err := tx.User.DeleteRelationsByUserIds(l.ctx, ids); err != nil {
			return err
		}
		if err := tx.User.DeletePasswordHistoryByUserIds(l.ctx, ids); err != nil {
			return err
		}
		return tx.User.Delete(l.ctx, ids)
	}); err != nil {
		return err
	}
	l.svcCtx.RbacGuard.InvalidateUsers(l.ctx, ids)
	for _, id := range ids {
		_ = l.svcCtx.SessionStore.DeleteAllByUser(l.ctx, id)
	}
	operlog.SetSummary(l.ctx, "删除用户 "+convert.Int64ToString(int64(len(ids)))+" 个")
	return nil
}

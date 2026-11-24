package role

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type RoleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleDeleteLogic {
	return &RoleDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RoleDeleteLogic) RoleDelete(req *types.IdsReq) error {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRoleDelete); err != nil {
		return err
	}
	ids, err := batchutil.CapRequestIDs(req.Ids)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errorx.New(400, "缺少 ids")
	}
	// 2. 事务前查出受影响用户，供事后刷缓存
	users, err := l.svcCtx.Repo.User.ListUserIdsByRoles(l.ctx, ids)
	if err != nil {
		return err
	}
	// 3. 短事务：清用户-角色、角色-资源、指定部门后删角色
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		if err := tx.User.DeleteRelationsByRoleIds(l.ctx, ids); err != nil {
			return err
		}
		if err := tx.Role.DeleteRelationsByRoleIds(l.ctx, ids); err != nil {
			return err
		}
		if err := tx.Role.DeleteOrgsByRoleIds(l.ctx, ids); err != nil {
			return err
		}
		return tx.Role.Delete(l.ctx, ids)
	}); err != nil {
		return err
	}
	// 4. 事务外刷缓存
	l.svcCtx.RbacGuard.InvalidateUsers(l.ctx, users)
	operlog.SetSummary(l.ctx, "删除角色 "+convert.Int64ToString(int64(len(ids)))+" 个")
	return nil
}

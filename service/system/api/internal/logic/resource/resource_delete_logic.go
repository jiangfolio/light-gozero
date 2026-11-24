package resource

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ResourceDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceDeleteLogic {
	return &ResourceDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResourceDeleteLogic) ResourceDelete(req *types.IdsReq) error {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourceDelete); err != nil {
		return err
	}
	ids, err := batchutil.CapRequestIDs(req.Ids)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errorx.New(400, "缺少 ids")
	}
	// 2. 有子节点则拒绝删除
	n, err := l.svcCtx.Repo.Resource.CountChildren(l.ctx, ids)
	if err != nil {
		return err
	}
	if n > 0 {
		return errorx.New(400, "资源存在子节点，无法删除")
	}
	// 3. 事务前查出受影响用户
	users, err := l.svcCtx.Repo.User.ListUserIdsByResources(l.ctx, ids)
	if err != nil {
		return err
	}
	// 4. 短事务：清角色-资源关联后删资源
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		if err := tx.Role.DeleteRelationsByResourceIds(l.ctx, ids); err != nil {
			return err
		}
		return tx.Resource.Delete(l.ctx, ids)
	}); err != nil {
		return err
	}
	// 5. 事务外刷缓存
	l.svcCtx.RbacGuard.InvalidateUsers(l.ctx, users)
	return nil
}

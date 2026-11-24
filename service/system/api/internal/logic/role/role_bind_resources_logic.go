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
	rolemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/role"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type RoleBindResourcesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleBindResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleBindResourcesLogic {
	return &RoleBindResourcesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RoleBindResourcesLogic) RoleBindResources(req *types.RoleBindResourcesReq) error {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRoleBindResources); err != nil {
		return err
	}
	// 2. 校验角色与资源（资源一次 IN 查出，禁止循环单查）
	resourceIds, err := batchutil.CapRequestIDs(req.ResourceIds)
	if err != nil {
		return err
	}
	if req.RoleId == "" || len(resourceIds) == 0 {
		return errorx.New(400, "角色和资源不能为空")
	}
	if _, err := l.svcCtx.Repo.Role.FindById(l.ctx, req.RoleId); err != nil {
		return errorx.New(404, "角色不存在")
	}
	resList, err := l.svcCtx.Repo.Resource.ListByIds(l.ctx, resourceIds)
	if err != nil {
		return err
	}
	if len(resList) != len(resourceIds) {
		return errorx.New(400, "资源不存在")
	}
	// 3. 批量写入关联（冲突忽略，幂等）；resource_type 从资源表带出，禁止调用方手填
	typeById := make(map[string]string, len(resList))
	for i := range resList {
		typeById[resList[i].Id] = resList[i].Type
	}
	op := authctx.UserId(l.ctx)
	rows := make([]rolemodel.RelRoleResource, 0, len(resourceIds))
	for _, rid := range resourceIds {
		id, err := l.svcCtx.IdGen.Next()
		if err != nil {
			return err
		}
		rows = append(rows, rolemodel.RelRoleResource{
			Id: id, RoleId: req.RoleId, ResourceId: rid, ResourceType: typeById[rid], CreatedBy: op,
		})
	}
	if err := l.svcCtx.Repo.Role.BindResources(l.ctx, rows); err != nil {
		return err
	}
	// 4. 事务外按角色用户刷缓存
	users, err := l.svcCtx.Repo.User.ListUserIdsByRoles(l.ctx, []string{req.RoleId})
	if err != nil {
		return err
	}
	l.svcCtx.RbacGuard.InvalidateUsers(l.ctx, users)
	operlog.SetSummary(l.ctx, "角色绑定资源 "+convert.Int64ToString(int64(len(resourceIds)))+" 个")
	return nil
}

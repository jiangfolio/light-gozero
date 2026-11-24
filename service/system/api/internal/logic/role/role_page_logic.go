package role

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type RolePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRolePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePageLogic {
	return &RolePageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RolePageLogic) RolePage(req *types.RolePageReq) (resp *types.RolePageResp, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemRolePage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	var status int64
	hasStatus := false
	if req.Status != "" {
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	list, total, err := l.svcCtx.Repo.Role.Page(l.ctx, req.Name, req.Code, status, hasStatus, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	// 批量加载指定部门，禁止按行查询
	roleIds := make([]string, 0, len(list))
	for i := range list {
		roleIds = append(roleIds, list[i].Id)
	}
	orgMap, err := l.svcCtx.Repo.Role.ListOrgIdsByRoleIds(l.ctx, roleIds)
	if err != nil {
		return nil, err
	}
	items := make([]types.RoleItem, 0, len(list))
	for i := range list {
		items = append(items, *toRoleItem(&list[i], orgMap[list[i].Id]))
	}
	return &types.RolePageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}

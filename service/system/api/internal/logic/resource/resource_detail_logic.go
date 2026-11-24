package resource

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	resmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ResourceDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceDetailLogic {
	return &ResourceDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResourceDetailLogic) ResourceDetail(req *types.ResourceDetailReq) (resp *types.ResourceItem, err error) {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourceDetail); err != nil {
		return nil, err
	}
	res, err := l.svcCtx.Repo.Resource.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "资源不存在")
	}
	return toResourceItem(res), nil
}

func toResourceItem(res *resmodel.SysResource) *types.ResourceItem {
	return &types.ResourceItem{
		Id: res.Id, ParentId: res.ParentId, Name: res.Name, Type: res.Type,
		PermCode: convert.DerefString(res.PermCode), Path: res.Path, ComponentPath: res.ComponentPath,
		Redirect: res.Redirect, RouteName: res.RouteName, Icon: res.Icon,
		Hidden: convert.Int64ToString(res.Hidden), KeepAlive: convert.Int64ToString(res.KeepAlive), Affix: convert.Int64ToString(res.Affix),
		Sort: convert.Int64ToString(res.Sort), Status: convert.Int64ToString(res.Status),
		CreatedBy: res.CreatedBy, UpdatedBy: res.UpdatedBy,
		CreatedAt: convert.TimeToString(res.CreatedAt), UpdatedAt: convert.TimeToString(res.UpdatedAt),
	}
}

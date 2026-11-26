package resource

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ResourceListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceListLogic {
	return &ResourceListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ResourceList 按 parent_id 一次性返回子资源列表（按钮权限等场景，禁止当全表列表用）。
func (l *ResourceListLogic) ResourceList(req *types.ResourceListReq) (*types.ResourceListResp, error) {
	// 1. 权限与必填父节点
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourceList); err != nil {
		return nil, err
	}
	if req.ParentId == "" {
		return nil, errorx.New(400, "缺少 parent_id")
	}
	// 2. 解析状态过滤
	var status int64
	hasStatus := false
	if req.Status != "" {
		var err error
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	// 3. 按父节点拉齐子资源
	list, err := l.svcCtx.Repo.Resource.ListByParent(l.ctx, req.ParentId, req.Name, req.Type, status, hasStatus)
	if err != nil {
		return nil, err
	}
	items := make([]types.ResourceItem, 0, len(list))
	for i := range list {
		items = append(items, *toResourceItem(&list[i]))
	}
	return &types.ResourceListResp{Items: items}, nil
}

package resource

import (
	"context"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/perm"
	resmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ResourceTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceTreeLogic {
	return &ResourceTreeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResourceTreeLogic) ResourceTree(req *types.ResourceTreeReq) (resp *types.ResourceTreeResp, err error) {
	// 1. 校验树查询权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourceTree); err != nil {
		return nil, err
	}
	// 2. 解析状态过滤
	var status int64
	hasStatus := false
	if req.Status != "" {
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	// 3. 仅拉目录/菜单组树（不含按钮；按钮按菜单 id 走 page）
	list, err := l.svcCtx.Repo.Resource.ListForTree(l.ctx, req.Name, req.Type, status, hasStatus)
	if err != nil {
		return nil, err
	}
	return &types.ResourceTreeResp{Items: buildAdminResourceTree(list)}, nil
}

type adminResNode struct {
	item     types.ResourceTreeNode
	children []*adminResNode
}

// buildAdminResourceTree 管理端资源树：字段与 ResourceItem 对齐，并带 children。
func buildAdminResourceTree(list []resmodel.SysResource) []types.ResourceTreeNode {
	nodes := make(map[string]*adminResNode, len(list))
	order := make([]string, 0, len(list))
	for i := range list {
		r := &list[i]
		item := toResourceItem(r)
		nodes[r.Id] = &adminResNode{item: types.ResourceTreeNode{
			Id: item.Id, ParentId: item.ParentId, Name: item.Name, Type: item.Type,
			PermCode: item.PermCode, Path: item.Path, ComponentPath: item.ComponentPath,
			Redirect: item.Redirect, RouteName: item.RouteName, Icon: item.Icon,
			Hidden: item.Hidden, KeepAlive: item.KeepAlive, Affix: item.Affix,
			Sort: item.Sort, Status: item.Status, CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy,
			CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		}}
		order = append(order, r.Id)
	}
	roots := make([]*adminResNode, 0)
	for _, id := range order {
		n := nodes[id]
		if n.item.ParentId != "" {
			if p, ok := nodes[n.item.ParentId]; ok {
				p.children = append(p.children, n)
				continue
			}
		}
		roots = append(roots, n)
	}
	out := make([]types.ResourceTreeNode, 0, len(roots))
	for _, r := range roots {
		out = append(out, flattenAdminResNode(r))
	}
	return out
}

func flattenAdminResNode(n *adminResNode) types.ResourceTreeNode {
	sort.Slice(n.children, func(i, j int) bool {
		return n.children[i].item.Sort < n.children[j].item.Sort
	})
	item := n.item
	item.Children = make([]types.ResourceTreeNode, 0, len(n.children))
	for _, c := range n.children {
		item.Children = append(item.Children, flattenAdminResNode(c))
	}
	return item
}

package user

import (
	"context"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	resmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserResourcesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserResourcesLogic {
	return &UserResourcesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserResourcesLogic) UserResources(req *types.UserResourcesReq) (resp *types.UserResourcesResp, err error) {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemUserResources); err != nil {
		return nil, err
	}
	// 2. 仅允许查自己的资源树
	userId := req.UserId
	cur := authctx.UserId(l.ctx)
	if userId == "" {
		userId = cur
	}
	if userId != cur {
		return nil, errorx.New(403, "只能查询本人资源")
	}

	var cached types.UserResourcesResp
	ok, _ := l.svcCtx.RbacGuard.GetCachedResources(l.ctx, userId, &cached)
	if ok {
		return &cached, nil
	}

	list, err := l.svcCtx.Repo.Resource.ListByUser(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	perms, err := l.svcCtx.RbacGuard.ListPerms(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	ptrs := make([]*resmodel.SysResource, 0, len(list))
	for i := range list {
		ptrs = append(ptrs, &list[i])
	}
	resp = &types.UserResourcesResp{Resources: buildResourceTree(ptrs), Perms: perms}
	l.svcCtx.RbacGuard.CacheResources(l.ctx, userId, resp)
	return resp, nil
}

type resNode struct {
	item     types.ResourceNode
	children []*resNode
}

func buildResourceTree(list []*resmodel.SysResource) []types.ResourceNode {
	nodes := make(map[string]*resNode, len(list))
	order := make([]string, 0, len(list))
	for _, r := range list {
		nodes[r.Id] = &resNode{item: types.ResourceNode{
			Id: r.Id, ParentId: r.ParentId, Name: r.Name, Type: r.Type,
			PermCode: convert.DerefString(r.PermCode), Path: r.Path, ComponentPath: r.ComponentPath,
			Redirect: r.Redirect, RouteName: r.RouteName, Icon: r.Icon,
			Hidden: convert.Int64ToString(r.Hidden), KeepAlive: convert.Int64ToString(r.KeepAlive), Affix: convert.Int64ToString(r.Affix),
			Sort: convert.Int64ToString(r.Sort), Status: convert.Int64ToString(r.Status),
		}}
		order = append(order, r.Id)
	}
	roots := make([]*resNode, 0)
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
	out := make([]types.ResourceNode, 0, len(roots))
	for _, r := range roots {
		out = append(out, flattenNode(r))
	}
	return out
}

func flattenNode(n *resNode) types.ResourceNode {
	sort.Slice(n.children, func(i, j int) bool {
		return n.children[i].item.Sort < n.children[j].item.Sort
	})
	item := n.item
	item.Children = make([]types.ResourceNode, 0, len(n.children))
	for _, c := range n.children {
		item.Children = append(item.Children, flattenNode(c))
	}
	return item
}

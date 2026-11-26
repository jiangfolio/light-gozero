package org

import (
	"context"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	orgmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/org"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type OrgTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTreeLogic {
	return &OrgTreeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *OrgTreeLogic) OrgTree(req *types.OrgTreeReq) (resp *types.OrgTreeResp, err error) {
	// 1. 校验树查询权限
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemOrgTree); err != nil {
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
	// 3. 套数据范围后全量查出再组树
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	sess, _ := authctx.SessionFrom(l.ctx)
	scopeOrgIds, selfOrgId, empty := datascope.ForOrgPage(scope, sess.OrgId)
	if empty {
		return &types.OrgTreeResp{Items: []types.OrgNode{}}, nil
	}
	list, err := l.svcCtx.Repo.Org.ListForTree(l.ctx, req.Name, req.Code, status, hasStatus, scopeOrgIds, selfOrgId)
	if err != nil {
		return nil, err
	}
	return &types.OrgTreeResp{Items: buildOrgTree(list)}, nil
}

type orgTreeNode struct {
	item     types.OrgNode
	children []*orgTreeNode
}

// buildOrgTree 按 parent_id 组装组织树；父不在结果集中的节点作为根。
func buildOrgTree(list []orgmodel.SysOrg) []types.OrgNode {
	nodes := make(map[string]*orgTreeNode, len(list))
	order := make([]string, 0, len(list))
	for i := range list {
		o := &list[i]
		item := toOrgItem(o)
		nodes[o.Id] = &orgTreeNode{item: types.OrgNode{
			Id: item.Id, ParentId: item.ParentId, Name: item.Name, Code: item.Code,
			Sort: item.Sort, Status: item.Status, CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy,
			CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		}}
		order = append(order, o.Id)
	}
	roots := make([]*orgTreeNode, 0)
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
	out := make([]types.OrgNode, 0, len(roots))
	for _, r := range roots {
		out = append(out, flattenOrgNode(r))
	}
	return out
}

func flattenOrgNode(n *orgTreeNode) types.OrgNode {
	sort.Slice(n.children, func(i, j int) bool {
		return n.children[i].item.Sort < n.children[j].item.Sort
	})
	item := n.item
	item.Children = make([]types.OrgNode, 0, len(n.children))
	for _, c := range n.children {
		item.Children = append(item.Children, flattenOrgNode(c))
	}
	return item
}

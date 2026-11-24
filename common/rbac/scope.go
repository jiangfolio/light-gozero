package rbac

import (
	"context"
	"encoding/json"
)

const userScopeKey = "rbac:user:scope:"

// DataScopeResult 用户可见数据范围（多角色并集后）。
type DataScopeResult struct {
	All      bool     `json:"all"`
	SelfOnly bool     `json:"self_only"`
	OrgIds   []string `json:"org_ids"`
}

// ScopeLoader 从持久层解析用户数据范围。
type ScopeLoader func(ctx context.Context, userId string) (*DataScopeResult, error)

// SetScopeLoader 注入数据范围加载器。
func (g *Guard) SetScopeLoader(loader ScopeLoader) {
	g.scopeLoader = loader
}

// ResolveDataScope 解析用户数据范围（带 Redis 缓存）。
func (g *Guard) ResolveDataScope(ctx context.Context, userId string) (*DataScopeResult, error) {
	if userId == "" {
		return &DataScopeResult{}, nil
	}
	key := userScopeKey + userId
	val, err := g.rdb.GetCtx(ctx, key)
	if err == nil && val != "" {
		var out DataScopeResult
		if json.Unmarshal([]byte(val), &out) == nil {
			return &out, nil
		}
	}
	if g.scopeLoader == nil {
		return &DataScopeResult{}, nil
	}
	out, err := g.scopeLoader(ctx, userId)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = &DataScopeResult{}
	}
	if out.OrgIds == nil {
		out.OrgIds = []string{}
	}
	raw, _ := json.Marshal(out)
	_ = g.rdb.SetexCtx(ctx, key, string(raw), cacheTTL)
	return out, nil
}

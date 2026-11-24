package rbac

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	userPermsKey = "rbac:user:perms:"
	userResKey   = "rbac:user:resources:"
	cacheTTL     = 30 * 60 // seconds
)

// PermLoader 从持久层加载用户权限码。
type PermLoader func(ctx context.Context, userId string) ([]string, error)

// Guard 权限校验与缓存。
type Guard struct {
	rdb         *redis.Redis
	loader      PermLoader
	scopeLoader ScopeLoader
}

func NewGuard(rdb *redis.Redis, loader PermLoader) *Guard {
	return &Guard{rdb: rdb, loader: loader}
}

// Require 要求当前用户具备指定权限码。
func (g *Guard) Require(ctx context.Context, userId, code string) error {
	codes, err := g.ListPerms(ctx, userId)
	if err != nil {
		return err
	}
	for _, c := range codes {
		if c == code {
			return nil
		}
	}
	return errorx.New(403, "无权限")
}

// ListPerms 获取用户权限码（带 Redis 缓存）。
func (g *Guard) ListPerms(ctx context.Context, userId string) ([]string, error) {
	key := userPermsKey + userId
	val, err := g.rdb.GetCtx(ctx, key)
	if err == nil && val != "" {
		var codes []string
		if json.Unmarshal([]byte(val), &codes) == nil {
			return codes, nil
		}
	}
	codes, err := g.loader(ctx, userId)
	if err != nil {
		return nil, err
	}
	if codes == nil {
		codes = []string{}
	}
	raw, _ := json.Marshal(codes)
	_ = g.rdb.SetexCtx(ctx, key, string(raw), cacheTTL)
	return codes, nil
}

// InvalidateUser 清理用户 RBAC 缓存（含数据范围）。
func (g *Guard) InvalidateUser(ctx context.Context, userId string) {
	_, _ = g.rdb.DelCtx(ctx, userPermsKey+userId)
	_, _ = g.rdb.DelCtx(ctx, userResKey+userId)
	_, _ = g.rdb.DelCtx(ctx, userScopeKey+userId)
}

// InvalidateUsers 批量清理。
func (g *Guard) InvalidateUsers(ctx context.Context, userIds []string) {
	for _, id := range userIds {
		if id == "" {
			continue
		}
		g.InvalidateUser(ctx, id)
	}
}

// CacheResources 缓存用户资源 JSON。
func (g *Guard) CacheResources(ctx context.Context, userId string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = g.rdb.SetexCtx(ctx, userResKey+userId, string(raw), cacheTTL)
}

// GetCachedResources 读取资源缓存。
func (g *Guard) GetCachedResources(ctx context.Context, userId string, dest any) (bool, error) {
	val, err := g.rdb.GetCtx(ctx, userResKey+userId)
	if err != nil {
		return false, err
	}
	if val == "" {
		return false, nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("decode resource cache: %w", err)
	}
	return true, nil
}

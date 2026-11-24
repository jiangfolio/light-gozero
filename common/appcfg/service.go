package appcfg

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const cacheKey = "config:app"

// Item 缓存中的单条配置。
type Item struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	ValueType   string `json:"value_type"`
	GroupName   string `json:"group_name"`
	Status      int64  `json:"status"`
}

// Loader 从持久化加载启用配置。
type Loader interface {
	ListEnabled(ctx context.Context) ([]Item, error)
}

// Service 带 Redis 短缓存的应用配置读取。
type Service struct {
	rdb      *redis.Redis
	loader   Loader
	cacheTTL time.Duration
}

// New 创建配置服务；cacheTTL 建议 60s。
func New(rdb *redis.Redis, loader Loader, cacheTTL time.Duration) *Service {
	if cacheTTL <= 0 {
		cacheTTL = 60 * time.Second
	}
	return &Service{rdb: rdb, loader: loader, cacheTTL: cacheTTL}
}

// GetString 读取字符串配置；缺失或停用时返回 def。
func (s *Service) GetString(ctx context.Context, key, def string) string {
	m, err := s.load(ctx)
	if err != nil {
		return def
	}
	item, ok := m[key]
	if !ok || item.Status != 1 || item.ConfigValue == "" {
		return def
	}
	return item.ConfigValue
}

// GetInt 读取整型配置；解析失败或缺失时返回 def。
func (s *Service) GetInt(ctx context.Context, key string, def int) int {
	raw := s.GetString(ctx, key, "")
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

// GetBool 读取布尔配置；支持 true/false/1/0，解析失败或缺失时返回 def。
func (s *Service) GetBool(ctx context.Context, key string, def bool) bool {
	raw := s.GetString(ctx, key, "")
	if raw == "" {
		return def
	}
	switch raw {
	case "true", "1":
		return true
	case "false", "0":
		return false
	default:
		return def
	}
}

// ListByKeys 按 key 列表取启用项（保持 keys 顺序）；未命中的跳过。
func (s *Service) ListByKeys(ctx context.Context, keys []string) ([]Item, error) {
	m, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Item, 0, len(keys))
	for _, k := range keys {
		item, ok := m[k]
		if !ok || item.Status != 1 {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// Invalidate 删除缓存，下次读库重载。
func (s *Service) Invalidate(ctx context.Context) error {
	_, err := s.rdb.DelCtx(ctx, cacheKey)
	return err
}

func (s *Service) load(ctx context.Context) (map[string]Item, error) {
	// 1. 读 Redis 全量缓存
	val, err := s.rdb.GetCtx(ctx, cacheKey)
	if err == nil && val != "" {
		var list []Item
		if json.Unmarshal([]byte(val), &list) == nil {
			return indexItems(list), nil
		}
	}
	// 2. miss 则装载库内启用项并回写缓存
	list, err := s.loader.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if payload, err := json.Marshal(list); err == nil {
		_ = s.rdb.SetexCtx(ctx, cacheKey, string(payload), int(s.cacheTTL.Seconds()))
	}
	return indexItems(list), nil
}

func indexItems(list []Item) map[string]Item {
	m := make(map[string]Item, len(list))
	for _, it := range list {
		m[it.ConfigKey] = it
	}
	return m
}

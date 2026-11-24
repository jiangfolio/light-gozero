package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	tokenKeyPrefix = "auth:token:"
	userTokensKey  = "auth:user:tokens:"
)

// Info 存入 Redis 的会话内容。
type Info struct {
	UserId             string `json:"user_id"`
	Username           string `json:"username"`
	OrgId              string `json:"org_id"`
	LoginAt            int64  `json:"login_at"`
	MustChangePassword bool   `json:"must_change_password"`
}

// TTLProvider 按请求取当前会话 TTL，支持改配热生效。
type TTLProvider func(ctx context.Context) time.Duration

// Store 基于 Redis 的不透明 token 会话存储。
type Store struct {
	rdb    *redis.Redis
	ttlFn  TTLProvider
}

func NewStore(rdb *redis.Redis, ttlFn TTLProvider) *Store {
	if ttlFn == nil {
		ttlFn = func(context.Context) time.Duration { return 24 * time.Hour }
	}
	return &Store{rdb: rdb, ttlFn: ttlFn}
}

// Create 生成 token 并写入 Redis。
func (s *Store) Create(ctx context.Context, info Info) (string, error) {
	// 1. 生成高熵不透明 token
	token, err := newToken()
	if err != nil {
		return "", err
	}
	info.LoginAt = time.Now().Unix()
	payload, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	// 2. 取当前 TTL（动态配置），写入会话键并登记用户 token 集合
	ttl := s.ttlFn(ctx)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	seconds := int(ttl.Seconds())
	if err := s.rdb.SetexCtx(ctx, tokenKeyPrefix+token, string(payload), seconds); err != nil {
		return "", err
	}
	_, err = s.rdb.SaddCtx(ctx, userTokensKey+info.UserId, token)
	if err != nil {
		_, _ = s.rdb.DelCtx(ctx, tokenKeyPrefix+token)
		return "", err
	}
	_ = s.rdb.ExpireCtx(ctx, userTokensKey+info.UserId, seconds)
	return token, nil
}

// Get 按 token 读取会话。
func (s *Store) Get(ctx context.Context, token string) (*Info, error) {
	val, err := s.rdb.GetCtx(ctx, tokenKeyPrefix+token)
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, nil
	}
	var info Info
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Delete 注销单个 token。
func (s *Store) Delete(ctx context.Context, token, userId string) error {
	// 1. 删除会话
	if _, err := s.rdb.DelCtx(ctx, tokenKeyPrefix+token); err != nil {
		return err
	}
	// 2. 从用户集合移除
	if userId != "" {
		_, err := s.rdb.SremCtx(ctx, userTokensKey+userId, token)
		return err
	}
	return nil
}

// DeleteAllByUser 吊销该用户全部会话（改密/停用后调用）。
func (s *Store) DeleteAllByUser(ctx context.Context, userId string) error {
	if userId == "" {
		return nil
	}
	setKey := userTokensKey + userId
	tokens, err := s.rdb.SmembersCtx(ctx, setKey)
	if err != nil {
		return err
	}
	for _, token := range tokens {
		_, _ = s.rdb.DelCtx(ctx, tokenKeyPrefix+token)
	}
	_, err = s.rdb.DelCtx(ctx, setKey)
	return err
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

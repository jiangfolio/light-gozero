package captcha

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"image/color"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const keyPrefix = "captcha:"

// 字符源：去掉易混的 0/O、1/I/L，接近常见后台验证码观感。
const charSource = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// TTLProvider 按请求取验证码 TTL。
type TTLProvider func(ctx context.Context) time.Duration

// LengthProvider 按请求取验证码位数。
type LengthProvider func(ctx context.Context) int

// Store 基于 Redis 的一次性图形验证码（渲染用 base64Captcha）。
type Store struct {
	rdb      *redis.Redis
	ttlFn    TTLProvider
	lengthFn LengthProvider
}

func NewStore(rdb *redis.Redis, ttlFn TTLProvider, lengthFn LengthProvider) *Store {
	if ttlFn == nil {
		ttlFn = func(context.Context) time.Duration { return 120 * time.Second }
	}
	if lengthFn == nil {
		lengthFn = func(context.Context) int { return 4 }
	}
	return &Store{rdb: rdb, ttlFn: ttlFn, lengthFn: lengthFn}
}

// Issue 生成验证码，返回 id 与 PNG base64（不含 data URI 前缀）。
func (s *Store) Issue(ctx context.Context) (id, b64 string, err error) {
	// 1. 取当前长度/TTL，生成 id
	length := s.lengthFn(ctx)
	if length <= 0 {
		length = 4
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	id = hex.EncodeToString(buf)

	// 2. 用浅底 + 空心干扰线的字符驱动绘制（替代老式绿噪点数字样式）
	bg := &color.RGBA{R: 248, G: 250, B: 252, A: 255}
	driver := base64Captcha.NewDriverString(
		40,
		120,
		0,
		base64Captcha.OptionShowHollowLine,
		length,
		charSource,
		bg,
		nil,
		[]string{"RitaSmith.ttf", "actionj.ttf", "chromohv.ttf"},
	).ConvertFonts()
	_, content, answer := driver.GenerateIdQuestionAnswer()
	item, err := driver.DrawCaptcha(content)
	if err != nil {
		return "", "", err
	}
	b64 = stripDataURI(item.EncodeB64string())

	// 3. 写入 Redis（一次性）
	ttl := s.ttlFn(ctx)
	if ttl <= 0 {
		ttl = 120 * time.Second
	}
	if err := s.rdb.SetexCtx(ctx, keyPrefix+id, answer, int(ttl.Seconds())); err != nil {
		return "", "", err
	}
	return id, b64, nil
}

// Verify 校验并删除；大小写不敏感。
func (s *Store) Verify(ctx context.Context, id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	key := keyPrefix + id
	val, err := s.rdb.GetCtx(ctx, key)
	_, _ = s.rdb.DelCtx(ctx, key)
	if err != nil || val == "" {
		return false
	}
	return strings.EqualFold(val, strings.TrimSpace(code))
}

func stripDataURI(s string) string {
	const prefix = "base64,"
	if i := strings.Index(s, prefix); i >= 0 {
		return s[i+len(prefix):]
	}
	return s
}

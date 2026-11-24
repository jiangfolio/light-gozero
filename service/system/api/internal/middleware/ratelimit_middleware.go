package middleware

import (
	"net/http"
	"strings"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/reqmeta"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RateLimitMiddleware 对登录/验证码/公钥按 IP 做周期限流。
type RateLimitMiddleware struct {
	login     *limit.PeriodLimit
	captcha   *limit.PeriodLimit
	publicKey *limit.PeriodLimit
}

func NewRateLimitMiddleware(login, captcha, publicKey *limit.PeriodLimit) *RateLimitMiddleware {
	return &RateLimitMiddleware{login: login, captcha: captcha, publicKey: publicKey}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limiter := m.pick(r.URL.Path)
		if limiter != nil {
			ip := reqmeta.ClientIP(r.Context())
			if ip == "" {
				ip = "unknown"
			}
			code, err := limiter.TakeCtx(r.Context(), ip)
			if err == nil && code == limit.OverQuota {
				httpx.ErrorCtx(r.Context(), w, errorx.New(429, "请求过于频繁，请稍后再试"))
				return
			}
		}
		next(w, r)
	}
}

func (m *RateLimitMiddleware) pick(path string) *limit.PeriodLimit {
	switch {
	case strings.HasSuffix(path, "/auth/login"):
		return m.login
	case strings.HasSuffix(path, "/auth/captcha"):
		return m.captcha
	case strings.HasSuffix(path, "/auth/public_key"):
		return m.publicKey
	default:
		return nil
	}
}

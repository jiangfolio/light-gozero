package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/jiangbyte/light-gozero/common/reqmeta"
)

// ClientMeta 提取客户端 IP / UA 写入 context（全路由）。
func ClientMeta(trustProxy bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r, trustProxy)
			ua := r.Header.Get("User-Agent")
			if len(ua) > 255 {
				ua = ua[:255]
			}
			ctx := reqmeta.WithClient(r.Context(), ip, ua)
			next(w, r.WithContext(ctx))
		}
	}
}

func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			return strings.TrimSpace(parts[0])
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/session"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// UserStatusLoader 按用户 ID 加载启用状态。
type UserStatusLoader func(userId string) (status int64, err error)

type AuthMiddleware struct {
	store  *session.Store
	loader UserStatusLoader
}

func NewAuthMiddleware(store *session.Store, loader UserStatusLoader) *AuthMiddleware {
	return &AuthMiddleware{store: store, loader: loader}
}

// Handle 校验 token、用户启用状态，并拦截强制改密场景下的非白名单请求。
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 读取原始 Authorization，拒绝 Bearer 前缀
		raw := strings.TrimSpace(r.Header.Get("Authorization"))
		if raw == "" || strings.HasPrefix(raw, "Bearer ") || strings.HasPrefix(strings.ToLower(raw), "bearer ") {
			httpx.ErrorCtx(r.Context(), w, errorx.New(401, "未登录或登录已失效"))
			return
		}
		// 2. Redis 查会话
		info, err := m.store.Get(r.Context(), raw)
		if err != nil || info == nil {
			httpx.ErrorCtx(r.Context(), w, errorx.New(401, "未登录或登录已失效"))
			return
		}
		// 3. 校验用户仍启用；停用则吊销当前 token
		if m.loader != nil {
			status, err := m.loader(info.UserId)
			if err != nil || status != 1 {
				_ = m.store.Delete(r.Context(), raw, info.UserId)
				httpx.ErrorCtx(r.Context(), w, errorx.New(401, "未登录或登录已失效"))
				return
			}
		}
		// 4. 强制改密：仅允许改密与登出
		if info.MustChangePassword && !allowWhenMustChange(r.URL.Path) {
			httpx.ErrorCtx(r.Context(), w, errorx.New(403, "请先修改密码"))
			return
		}
		// 5. 写入 authctx 后放行
		ctx := authctx.WithSession(r.Context(), authctx.Session{
			UserId:             info.UserId,
			Username:           info.Username,
			OrgId:              info.OrgId,
			Token:              raw,
			MustChangePassword: info.MustChangePassword,
		})
		next(w, r.WithContext(ctx))
	}
}

func allowWhenMustChange(path string) bool {
	return path == "/api/v1/auth/change_password" || path == "/api/v1/auth/logout"
}

package authctx

import "context"

type ctxKey struct{}

// Session 登录会话快照，由鉴权中间件写入 context。
type Session struct {
	UserId             string
	Username           string
	OrgId              string
	Token              string
	MustChangePassword bool
}

// WithSession 将会话写入 context。
func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

// SessionFrom 读取会话；未登录 ok=false。
func SessionFrom(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(ctxKey{}).(Session)
	return s, ok
}

// UserId 返回当前用户 ID，未登录返回空串。
func UserId(ctx context.Context) string {
	s, ok := SessionFrom(ctx)
	if !ok {
		return ""
	}
	return s.UserId
}

// Token 返回当前请求 token。
func Token(ctx context.Context) string {
	s, ok := SessionFrom(ctx)
	if !ok {
		return ""
	}
	return s.Token
}

// MustChangePassword 是否强制改密。
func MustChangePassword(ctx context.Context) bool {
	s, ok := SessionFrom(ctx)
	return ok && s.MustChangePassword
}

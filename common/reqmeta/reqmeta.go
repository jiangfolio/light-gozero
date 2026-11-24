package reqmeta

import "context"

type ctxKey int

const (
	keyIP ctxKey = iota
	keyUA
)

// WithClient 写入客户端 IP 与 User-Agent。
func WithClient(ctx context.Context, ip, ua string) context.Context {
	ctx = context.WithValue(ctx, keyIP, ip)
	return context.WithValue(ctx, keyUA, ua)
}

func ClientIP(ctx context.Context) string {
	v, _ := ctx.Value(keyIP).(string)
	return v
}

func UserAgent(ctx context.Context) string {
	v, _ := ctx.Value(keyUA).(string)
	return v
}

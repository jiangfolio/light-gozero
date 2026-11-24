package operlog

import "context"

type ctxKey struct{}

// holder 可变摘要槽，由中间件注入，logic 写入后中间件读取。
type holder struct {
	text string
}

// WithHolder 注入可写摘要槽（中间件在调用 handler 前使用）。
func WithHolder(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, &holder{})
}

// SetSummary 写入人类可读操作摘要；无 holder 时静默忽略。
func SetSummary(ctx context.Context, summary string) {
	if h, ok := ctx.Value(ctxKey{}).(*holder); ok && h != nil {
		h.text = summary
	}
}

// GetSummary 读取摘要。
func GetSummary(ctx context.Context) string {
	if h, ok := ctx.Value(ctxKey{}).(*holder); ok && h != nil {
		return h.text
	}
	return ""
}

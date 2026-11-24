package errorx

import "fmt"

// CodeError 带业务码的可对外错误，供统一 ErrorHandler 映射为 code/message。
type CodeError struct {
	Code    int
	Message string
}

func (e *CodeError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// New 构造业务错误。
func New(code int, message string) *CodeError {
	return &CodeError{Code: code, Message: message}
}

// Newf 按格式化消息构造业务错误。
func Newf(code int, format string, args ...any) *CodeError {
	return &CodeError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// CodeOf 提取业务码；非 CodeError 返回 0。
func CodeOf(err error) int {
	if ce, ok := err.(*CodeError); ok && ce != nil {
		return ce.Code
	}
	return 0
}

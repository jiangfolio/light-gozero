package response

import (
	"context"
	"errors"
	"net/http"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Body 统一 HTTP 包络：业务数据放在 data，错误信息放在 message。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

const (
	codeOK           = 0
	codeInternalFail = -1
	msgOK            = "成功"
	msgInternalFail  = "系统内部错误"
)

// Setup 注册 go-zero 官方 Ok/Error Handler，使 OkJsonCtx/ErrorCtx 自动包络。
func Setup() {
	// 1. 成功响应：把业务体包进 code/message/data
	httpx.SetOkHandler(func(_ context.Context, v any) any {
		return Body{
			Code:    codeOK,
			Message: msgOK,
			Data:    v,
		}
	})

	// 2. 失败响应：优先映射业务 CodeError；其余记日志后返回通用失败体
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		var ce *errorx.CodeError
		if errors.As(err, &ce) && ce != nil {
			return http.StatusOK, Body{
				Code:    ce.Code,
				Message: ce.Message,
			}
		}
		logc.Errorf(ctx, "unhandled error: %v", err)
		return http.StatusOK, Body{
			Code:    codeInternalFail,
			Message: msgInternalFail,
		}
	})
}

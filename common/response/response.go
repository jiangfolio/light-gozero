package response

import (
	"context"
	"errors"
	"net/http"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/rest"
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
	codeNotFound     = 404
	codeNotAllowed   = 405
	msgOK            = "成功"
	msgInternalFail  = "系统内部错误"
	msgNotFound      = "接口不存在"
	msgNotAllowed    = "请求方法不允许"
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

// RestOptions 路由级 404/405 也走统一 JSON 包络，供 MustNewServer 使用。
func RestOptions() []rest.RunOption {
	return []rest.RunOption{
		rest.WithNotFoundHandler(NotFoundHandler()),
		rest.WithNotAllowedHandler(NotAllowedHandler()),
	}
}

// NotFoundHandler 未匹配路由时返回统一包络，避免裸文本 "404 page not found"。
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 写出与业务错误一致的 code/message JSON
		httpx.WriteJsonCtx(r.Context(), w, http.StatusNotFound, Body{
			Code:    codeNotFound,
			Message: msgNotFound,
		})
	})
}

// NotAllowedHandler 方法不允许时返回统一包络。
func NotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusMethodNotAllowed, Body{
			Code:    codeNotAllowed,
			Message: msgNotAllowed,
		})
	})
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"context"

	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HelloLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHelloLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HelloLogic {
	return &HelloLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Hello 占位演示接口：返回问候语，并对可选 user_id 做显式数字字符串往返转换。
func (l *HelloLogic) Hello(req *types.HelloReq) (resp *types.HelloResp, err error) {
	// 1. 入站：有 user_id 时显式转 int64，非法值映射为业务错误
	userId := req.UserId
	if userId != "" {
		id, convErr := convert.Int64FromString(userId)
		if convErr != nil {
			return nil, errorx.New(400, "用户 ID 无效")
		}
		// 2. 出站：再显式转回字符串，保证响应字段仍是下划线风格的字符串 ID
		userId = convert.Int64ToString(id)
	}

	// 3. 组装业务 data（外层 code/message 由统一 OkHandler 包络）
	return &types.HelloResp{
		Message: "hello",
		UserId:  userId,
	}, nil
}

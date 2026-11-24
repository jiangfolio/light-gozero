package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type CaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Captcha 签发一次性图形验证码。
func (l *CaptchaLogic) Captcha() (resp *types.CaptchaResp, err error) {
	id, img, err := l.svcCtx.CaptchaStore.Issue(l.ctx)
	if err != nil {
		return nil, err
	}
	return &types.CaptchaResp{CaptchaId: id, CaptchaImage: img}, nil
}

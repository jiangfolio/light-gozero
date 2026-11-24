package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type PublicKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicKeyLogic {
	return &PublicKeyLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// PublicKey 返回密码传输用公钥（不暴露算法细节）。
func (l *PublicKeyLogic) PublicKey() (resp *types.PublicKeyResp, err error) {
	return &types.PublicKeyResp{PublicKey: l.svcCtx.PwdCipher.PublicKeyHex()}, nil
}

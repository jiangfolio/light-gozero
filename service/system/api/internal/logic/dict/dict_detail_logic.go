package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictDetailLogic {
	return &DictDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictDetailLogic) DictDetail(req *types.DictDetailReq) (resp *types.DictItem, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictDetail); err != nil {
		return nil, err
	}
	row, err := l.svcCtx.Repo.Dict.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "字典不存在")
	}
	return toDictItem(row), nil
}

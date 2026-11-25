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

type DictTypeDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeDetailLogic {
	return &DictTypeDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictTypeDetailLogic) DictTypeDetail(req *types.DictTypeDetailReq) (resp *types.DictInfo, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictDetail); err != nil {
		return nil, err
	}
	row, err := l.svcCtx.Repo.Dict.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "字典类型不存在")
	}
	if row.ParentId != "" {
		return nil, errorx.New(400, "目标不是字典类型")
	}
	return toDictInfo(row), nil
}

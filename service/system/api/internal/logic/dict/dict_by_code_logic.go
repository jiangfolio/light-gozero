package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictByCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictByCodeLogic {
	return &DictByCodeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// DictByCode 弱鉴权只读：仅需登录，不校验字典管理权限。
func (l *DictByCodeLogic) DictByCode(req *types.DictByCodeReq) (resp *types.DictByCodeResp, err error) {
	if req.DictCode == "" {
		return nil, errorx.New(400, "缺少 dict_code")
	}
	list, err := l.svcCtx.Repo.Dict.ListEnabledItemsByCode(l.ctx, req.DictCode)
	if err != nil {
		return nil, err
	}
	items := make([]types.DictByCodeItem, 0, len(list))
	for _, row := range list {
		items = append(items, types.DictByCodeItem{
			Label: row.Label, Value: row.Value, Sort: convert.Int64ToString(row.Sort),
		})
	}
	return &types.DictByCodeResp{Items: items}, nil
}

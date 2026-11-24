package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictPageLogic {
	return &DictPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictPageLogic) DictPage(req *types.DictPageReq) (resp *types.DictPageResp, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictPage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	var status int64
	hasStatus := false
	if req.Status != "" {
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	list, total, err := l.svcCtx.Repo.Dict.Page(l.ctx, req.ParentId, req.DictCode, req.Label, status, hasStatus, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.DictItem, 0, len(list))
	for i := range list {
		items = append(items, *toDictItem(&list[i]))
	}
	return &types.DictPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)),
		PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}

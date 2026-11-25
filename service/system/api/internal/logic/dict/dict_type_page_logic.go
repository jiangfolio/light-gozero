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

type DictTypePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypePageLogic {
	return &DictTypePageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictTypePageLogic) DictTypePage(req *types.DictTypePageReq) (resp *types.DictTypePageResp, err error) {
	// 1. 校验权限与分页参数
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
	// 2. 仅查类型行（repo 固定 parent_id=''）
	list, total, err := l.svcCtx.Repo.Dict.PageTypes(l.ctx, req.DictCode, req.Label, status, hasStatus, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.DictInfo, 0, len(list))
	for i := range list {
		items = append(items, *toDictInfo(&list[i]))
	}
	return &types.DictTypePageResp{
		Items: items, Page: convert.Int64ToString(int64(page)),
		PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}

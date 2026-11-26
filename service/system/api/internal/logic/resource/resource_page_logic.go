package resource

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

type ResourcePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourcePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourcePageLogic {
	return &ResourcePageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResourcePageLogic) ResourcePage(req *types.ResourcePageReq) (resp *types.ResourcePageResp, err error) {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourcePage); err != nil {
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
	list, total, err := l.svcCtx.Repo.Resource.Page(
		l.ctx, req.Name, req.Type, req.ParentId, status, hasStatus,
		pageutil.Offset(page, pageSize), pageSize,
	)
	if err != nil {
		return nil, err
	}
	items := make([]types.ResourceItem, 0, len(list))
	for i := range list {
		items = append(items, *toResourceItem(&list[i]))
	}
	return &types.ResourcePageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}

package notice

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

type NoticePageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticePageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticePageLogic {
	return &NoticePageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticePageLogic) NoticePage(req *types.NoticePageReq) (resp *types.NoticePageResp, err error) {
	// 1. 权限与分页
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemNoticePage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	// 2. 可选过滤
	var typ, status int64
	hasType, hasStatus := false, false
	if req.Type != "" {
		typ, err = parseNoticeType(req.Type, 0)
		if err != nil {
			return nil, err
		}
		hasType = true
	}
	if req.Status != "" {
		status, err = parseNoticeStatus(req.Status, 0)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	list, total, err := l.svcCtx.Repo.Notice.Page(
		l.ctx, req.Title, typ, status, hasType, hasStatus, pageutil.Offset(page, pageSize), pageSize,
	)
	if err != nil {
		return nil, err
	}
	items := make([]types.NoticeInfo, 0, len(list))
	for i := range list {
		items = append(items, toNoticeInfo(&list[i]))
	}
	return &types.NoticePageResp{
		Items:    items,
		Page:     convert.Int64ToString(int64(page)),
		PageSize: convert.Int64ToString(int64(pageSize)),
		Total:    convert.Int64ToString(total),
	}, nil
}

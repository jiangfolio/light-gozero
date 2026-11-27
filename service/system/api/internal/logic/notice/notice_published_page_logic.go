package notice

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type NoticePublishedPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticePublishedPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticePublishedPageLogic {
	return &NoticePublishedPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticePublishedPageLogic) NoticePublishedPage(req *types.NoticePublishedPageReq) (resp *types.NoticePageResp, err error) {
	// 1. 仅需登录；强制已发布
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	var typ int64
	hasType := false
	if req.Type != "" {
		typ, err = parseNoticeType(req.Type, 0)
		if err != nil {
			return nil, err
		}
		hasType = true
	}
	list, total, err := l.svcCtx.Repo.Notice.PagePublished(
		l.ctx, req.Title, typ, hasType, pageutil.Offset(page, pageSize), pageSize,
	)
	if err != nil {
		return nil, err
	}
	items := make([]types.NoticeInfo, 0, len(list))
	for i := range list {
		items = append(items, toPublishedNoticeInfo(&list[i]))
	}
	return &types.NoticePageResp{
		Items:    items,
		Page:     convert.Int64ToString(int64(page)),
		PageSize: convert.Int64ToString(int64(pageSize)),
		Total:    convert.Int64ToString(total),
	}, nil
}

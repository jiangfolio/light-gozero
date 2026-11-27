package notice

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/errorx"
	noticemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/notice"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type NoticePublishedDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticePublishedDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticePublishedDetailLogic {
	return &NoticePublishedDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticePublishedDetailLogic) NoticePublishedDetail(req *types.NoticePublishedDetailReq) (resp *types.NoticeInfo, err error) {
	// 1. 仅需登录；草稿对普通用户不可见
	if req.Id == "" {
		return nil, errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Notice.FindById(l.ctx, req.Id)
	if err != nil || row.Status != noticemodel.StatusPublished {
		return nil, errorx.New(404, "通知公告不存在")
	}
	info := toPublishedNoticeInfo(row)
	return &info, nil
}

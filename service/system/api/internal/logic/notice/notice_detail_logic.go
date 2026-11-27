package notice

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type NoticeDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeDetailLogic {
	return &NoticeDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticeDetailLogic) NoticeDetail(req *types.NoticeDetailReq) (resp *types.NoticeInfo, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemNoticeDetail); err != nil {
		return nil, err
	}
	if req.Id == "" {
		return nil, errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Notice.FindById(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.New(404, "通知公告不存在")
	}
	info := toNoticeInfo(row)
	return &info, nil
}

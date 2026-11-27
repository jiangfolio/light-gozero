package notice

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type NoticeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeUpdateLogic {
	return &NoticeUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticeUpdateLogic) NoticeUpdate(req *types.NoticeUpdateReq) error {
	// 1. 权限与目标存在
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemNoticeUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Notice.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "通知公告不存在")
	}
	// 2. 按入参覆盖可改字段（传了才改；标题/正文若传则不可变空白）
	if req.Title != "" {
		title, _, _, err := normalizeNoticeText(req.Title, "x", "", false)
		if err != nil {
			return err
		}
		if title == "" {
			return errorx.New(400, "标题不能为空")
		}
		row.Title = title
	}
	if req.Content != "" {
		_, content, _, err := normalizeNoticeText("x", req.Content, "", false)
		if err != nil {
			return err
		}
		if content == "" {
			return errorx.New(400, "正文不能为空")
		}
		row.Content = content
	}
	if req.Type != "" {
		v, err := parseNoticeType(req.Type, row.Type)
		if err != nil {
			return err
		}
		row.Type = v
	}
	if req.Status != "" {
		v, err := parseNoticeStatus(req.Status, row.Status)
		if err != nil {
			return err
		}
		row.Status = v
	}
	if req.Sort != "" {
		v, err := convert.ParseSort(req.Sort, row.Sort)
		if err != nil {
			return err
		}
		row.Sort = v
	}
	if req.Remark != "" {
		_, _, remark, err := normalizeNoticeText("x", "x", req.Remark, false)
		if err != nil {
			return err
		}
		row.Remark = remark
	}
	applyPublishAt(row, time.Now())
	row.UpdatedBy = authctx.UserId(l.ctx)
	if err := l.svcCtx.Repo.Notice.Save(l.ctx, row); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "更新通知公告「"+summaryTitle(row.Title)+"」")
	return nil
}

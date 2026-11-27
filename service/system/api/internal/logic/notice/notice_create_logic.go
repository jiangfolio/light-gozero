package notice

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	noticemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/notice"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type NoticeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeCreateLogic {
	return &NoticeCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticeCreateLogic) NoticeCreate(req *types.NoticeCreateReq) (resp *types.IdResp, err error) {
	// 1. 权限与文本校验
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemNoticeCreate); err != nil {
		return nil, err
	}
	title, content, remark, err := normalizeNoticeText(req.Title, req.Content, req.Remark, true)
	if err != nil {
		return nil, err
	}
	// 2. 解析类型 / 状态 / 排序
	if strings.TrimSpace(req.Type) == "" {
		return nil, errorx.New(400, "类型不能为空")
	}
	typ, err := parseNoticeType(req.Type, 0)
	if err != nil {
		return nil, err
	}
	status, err := parseNoticeStatus(req.Status, noticemodel.StatusDraft)
	if err != nil {
		return nil, err
	}
	sort, err := convert.ParseSort(req.Sort, 0)
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	op := authctx.UserId(l.ctx)
	row := &noticemodel.SysNotice{
		Id: id, Title: title, Content: content, Type: typ, Status: status,
		Sort: sort, Remark: remark, CreatedBy: op, UpdatedBy: op,
	}
	// 3. 已发布则补发布时间
	applyPublishAt(row, time.Now())
	if err := l.svcCtx.Repo.Notice.Create(l.ctx, row); err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建通知公告「"+summaryTitle(title)+"」")
	return &types.IdResp{Id: id}, nil
}

package notice

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type NoticeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNoticeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeDeleteLogic {
	return &NoticeDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *NoticeDeleteLogic) NoticeDelete(req *types.IdsReq) error {
	// 1. 权限与批量上限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemNoticeDelete); err != nil {
		return err
	}
	ids, err := batchutil.CapRequestIDs(req.Ids)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errorx.New(400, "缺少 ids")
	}
	// 2. 确认全部存在后硬删
	rows, err := l.svcCtx.Repo.Notice.ListByIds(l.ctx, ids)
	if err != nil {
		return err
	}
	if len(rows) != len(ids) {
		return errorx.New(404, "通知公告不存在")
	}
	if err := l.svcCtx.Repo.Notice.Delete(l.ctx, ids); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "删除通知公告 "+convert.Int64ToString(int64(len(ids)))+" 条")
	return nil
}

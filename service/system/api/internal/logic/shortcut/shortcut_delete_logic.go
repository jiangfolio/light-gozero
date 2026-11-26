// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package shortcut

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ShortcutDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShortcutDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShortcutDeleteLogic {
	return &ShortcutDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ShortcutDelete 删除本人快捷入口（按 ids，且强制 user_id 条件）。
func (l *ShortcutDeleteLogic) ShortcutDelete(req *types.IdsReq) error {
	// 1. 登录与入参
	uid := authctx.UserId(l.ctx)
	if uid == "" {
		return errorx.New(401, "未登录或登录已失效")
	}
	ids := batchutil.UniqueStrings(req.Ids)
	if len(ids) == 0 {
		return errorx.New(400, "请选择要删除的快捷入口")
	}
	if len(ids) > shortcutDeleteBatchMax {
		return errorx.New(400, "单次删除数量超出上限")
	}
	// 2. 仅删本人数据
	if err := l.svcCtx.Repo.Shortcut.DeleteByUserAndIds(l.ctx, uid, ids); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "删除快捷入口")
	return nil
}

const shortcutDeleteBatchMax = 30

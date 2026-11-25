package dict

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

type DictTypeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeDeleteLogic {
	return &DictTypeDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictTypeDeleteLogic) DictTypeDelete(req *types.IdsReq) error {
	// 1. 校验权限与入参
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictDelete); err != nil {
		return err
	}
	ids, err := batchutil.CapRequestIDs(req.Ids)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return errorx.New(400, "缺少 ids")
	}
	// 2. 全部必须是字典类型，避免误删项
	rows, err := l.svcCtx.Repo.Dict.ListByIds(l.ctx, ids)
	if err != nil {
		return err
	}
	if len(rows) != len(ids) {
		return errorx.New(404, "字典类型不存在")
	}
	for i := range rows {
		if rows[i].ParentId != "" {
			return errorx.New(400, "目标不是字典类型")
		}
	}
	// 3. 仍有子项则拒绝，避免悬空
	n, err := l.svcCtx.Repo.Dict.CountChildren(l.ctx, ids)
	if err != nil {
		return err
	}
	if n > 0 {
		return errorx.New(400, "字典存在子项，无法删除")
	}
	if err := l.svcCtx.Repo.Dict.Delete(l.ctx, ids); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "删除字典类型 "+convert.Int64ToString(int64(len(ids)))+" 个")
	return nil
}

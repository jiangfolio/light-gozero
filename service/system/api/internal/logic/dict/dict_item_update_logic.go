package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictItemUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemUpdateLogic {
	return &DictItemUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictItemUpdateLogic) DictItemUpdate(req *types.DictItemUpdateReq) error {
	// 1. 校验权限与目标存在，且必须是字典项
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Dict.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "字典项不存在")
	}
	if row.ParentId == "" {
		return errorx.New(400, "目标不是字典项")
	}
	// 2. 按需更新字段；value 若传入不得为空
	if req.Label != "" {
		row.Label = req.Label
	}
	if req.Value != "" {
		row.Value = req.Value
	}
	if req.Sort != "" {
		v, err := convert.ParseSort(req.Sort, row.Sort)
		if err != nil {
			return err
		}
		row.Sort = v
	}
	if req.Status != "" {
		v, err := convert.ParseStatus(req.Status, row.Status)
		if err != nil {
			return err
		}
		row.Status = v
	}
	if req.Remark != "" {
		row.Remark = req.Remark
	}
	row.UpdatedBy = authctx.UserId(l.ctx)
	if err := l.svcCtx.Repo.Dict.Save(l.ctx, row); err != nil {
		return err
	}
	operlog.SetSummary(l.ctx, "更新字典项「"+row.Label+"」")
	return nil
}

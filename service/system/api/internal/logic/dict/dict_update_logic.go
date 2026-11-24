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

type DictUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictUpdateLogic {
	return &DictUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictUpdateLogic) DictUpdate(req *types.DictUpdateReq) error {
	// 1. 校验权限与目标存在
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Dict.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "字典不存在")
	}
	// 2. 类型禁止改 value；项改 value 时不得为空
	if req.Label != "" {
		row.Label = req.Label
	}
	if req.Value != "" {
		if row.ParentId == "" {
			return errorx.New(400, "字典类型取值须为空")
		}
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
	operlog.SetSummary(l.ctx, "更新字典「"+row.Label+"」")
	return nil
}

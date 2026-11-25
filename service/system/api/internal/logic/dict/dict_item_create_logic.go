package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	dictmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/dict"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictItemCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemCreateLogic {
	return &DictItemCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictItemCreateLogic) DictItemCreate(req *types.DictItemCreateReq) (resp *types.IdResp, err error) {
	// 1. 校验权限与必填
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictCreate); err != nil {
		return nil, err
	}
	if req.ParentId == "" {
		return nil, errorx.New(400, "缺少 parent_id")
	}
	if req.Label == "" {
		return nil, errorx.New(400, "名称不能为空")
	}
	if req.Value == "" {
		return nil, errorx.New(400, "字典项取值不能为空")
	}
	// 2. 父级必须是字典类型；编码从父类继承，忽略客户端传入
	parent, err := l.svcCtx.Repo.Dict.FindById(l.ctx, req.ParentId)
	if err != nil {
		return nil, errorx.New(400, "上级字典不存在")
	}
	if parent.ParentId != "" {
		return nil, errorx.New(400, "不能挂到字典项下")
	}
	sort, err := convert.ParseSort(req.Sort, 0)
	if err != nil {
		return nil, err
	}
	status, err := convert.ParseStatus(req.Status, 1)
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	op := authctx.UserId(l.ctx)
	// 3. 写入项行
	if err := l.svcCtx.Repo.Dict.Create(l.ctx, &dictmodel.SysDict{
		Id: id, ParentId: req.ParentId, DictCode: parent.DictCode, Label: req.Label, Value: req.Value,
		Sort: sort, Status: status, Remark: req.Remark, CreatedBy: op, UpdatedBy: op,
	}); err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建字典项「"+req.Label+"」")
	return &types.IdResp{Id: id}, nil
}

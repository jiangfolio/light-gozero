package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	dictmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/dict"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictTypeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictTypeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictTypeCreateLogic {
	return &DictTypeCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictTypeCreateLogic) DictTypeCreate(req *types.DictTypeCreateReq) (resp *types.IdResp, err error) {
	// 1. 校验权限与必填：类型必须有编码与名称
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictCreate); err != nil {
		return nil, err
	}
	if req.DictCode == "" || req.Label == "" {
		return nil, errorx.New(400, "编码和名称不能为空")
	}
	// 2. 编码全局唯一（类型行 value 为空，靠 uk_sys_dict_code_value 约束）
	if _, err := l.svcCtx.Repo.Dict.FindTypeByCode(l.ctx, req.DictCode); err == nil {
		return nil, errorx.New(400, "字典编码已存在")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
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
	// 3. 写入类型行：parent_id/value 强制为空
	if err := l.svcCtx.Repo.Dict.Create(l.ctx, &dictmodel.SysDict{
		Id: id, ParentId: "", DictCode: req.DictCode, Label: req.Label, Value: "",
		Sort: sort, Status: status, Remark: req.Remark, CreatedBy: op, UpdatedBy: op,
	}); err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建字典类型「"+req.Label+"」")
	return &types.IdResp{Id: id}, nil
}

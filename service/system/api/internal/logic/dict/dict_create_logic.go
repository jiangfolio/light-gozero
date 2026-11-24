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

type DictCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictCreateLogic {
	return &DictCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictCreateLogic) DictCreate(req *types.DictCreateReq) (resp *types.IdResp, err error) {
	// 1. 校验权限与必填
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictCreate); err != nil {
		return nil, err
	}
	if req.Label == "" {
		return nil, errorx.New(400, "名称不能为空")
	}

	parentId := req.ParentId
	value := req.Value
	dictCode := req.DictCode
	// 2. 类型行：parent 空、value 强制空；项行：挂到类型、value 必填
	if parentId == "" {
		if dictCode == "" {
			return nil, errorx.New(400, "编码和名称不能为空")
		}
		value = ""
		if _, err := l.svcCtx.Repo.Dict.FindTypeByCode(l.ctx, dictCode); err == nil {
			return nil, errorx.New(400, "字典编码已存在")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	} else {
		parent, err := l.svcCtx.Repo.Dict.FindById(l.ctx, parentId)
		if err != nil {
			return nil, errorx.New(400, "上级字典不存在")
		}
		if parent.ParentId != "" {
			return nil, errorx.New(400, "不能挂到字典项下")
		}
		if value == "" {
			return nil, errorx.New(400, "字典项取值不能为空")
		}
		dictCode = parent.DictCode
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
	if err := l.svcCtx.Repo.Dict.Create(l.ctx, &dictmodel.SysDict{
		Id: id, ParentId: parentId, DictCode: dictCode, Label: req.Label, Value: value,
		Sort: sort, Status: status, Remark: req.Remark, CreatedBy: op, UpdatedBy: op,
	}); err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "创建字典「"+req.Label+"」")
	return &types.IdResp{Id: id}, nil
}

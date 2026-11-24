package resource

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	resmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ResourceCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceCreateLogic {
	return &ResourceCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResourceCreateLogic) ResourceCreate(req *types.ResourceCreateReq) (resp *types.IdResp, err error) {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourceCreate); err != nil {
		return nil, err
	}
	if req.Name == "" || req.Type == "" {
		return nil, errorx.New(400, "名称和类型不能为空")
	}
	switch req.Type {
	case "directory", "menu", "button":
	default:
		return nil, errorx.New(400, "类型无效")
	}
	if req.Type == "button" {
		if req.PermCode == "" || !perm.Valid(req.PermCode) {
			return nil, errorx.New(400, "权限码无效")
		}
	} else if req.PermCode != "" && !perm.Valid(req.PermCode) {
		return nil, errorx.New(400, "权限码无效")
	}
	if req.ParentId != "" {
		if _, err := l.svcCtx.Repo.Resource.FindById(l.ctx, req.ParentId); err != nil {
			return nil, errorx.New(400, "上级资源不存在")
		}
	}
	sort, err := convert.ParseSort(req.Sort, 0)
	if err != nil {
		return nil, err
	}
	status, err := convert.ParseStatus(req.Status, 1)
	if err != nil {
		return nil, err
	}
	hidden, err := convert.ParseFlag(req.Hidden, 0)
	if err != nil {
		return nil, err
	}
	keepAlive, err := convert.ParseFlag(req.KeepAlive, 0)
	if err != nil {
		return nil, err
	}
	affix, err := convert.ParseFlag(req.Affix, 0)
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	op := authctx.UserId(l.ctx)
	err = l.svcCtx.Repo.Resource.Create(l.ctx, &resmodel.SysResource{
		Id: id, ParentId: req.ParentId, Name: req.Name, Type: req.Type,
		PermCode: convert.StringPtr(req.PermCode), Path: req.Path, ComponentPath: req.ComponentPath,
		Redirect: req.Redirect, RouteName: req.RouteName, Icon: req.Icon,
		Hidden: hidden, KeepAlive: keepAlive, Affix: affix, Sort: sort, Status: status,
		CreatedBy: op, UpdatedBy: op,
	})
	if err != nil {
		return nil, err
	}
	return &types.IdResp{Id: id}, nil
}

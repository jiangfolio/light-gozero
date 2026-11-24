package resource

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ResourceUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceUpdateLogic {
	return &ResourceUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResourceUpdateLogic) ResourceUpdate(req *types.ResourceUpdateReq) error {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemResourceUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	res, err := l.svcCtx.Repo.Resource.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "资源不存在")
	}
	if req.Type != "" {
		switch req.Type {
		case "directory", "menu", "button":
			res.Type = req.Type
		default:
			return errorx.New(400, "类型无效")
		}
	}
	if req.ParentId != "" {
		if req.ParentId == req.Id {
			return errorx.New(400, "上级不能是自身")
		}
		if _, err := l.svcCtx.Repo.Resource.FindById(l.ctx, req.ParentId); err != nil {
			return errorx.New(400, "上级资源不存在")
		}
		res.ParentId = req.ParentId
	}
	if req.Name != "" {
		res.Name = req.Name
	}
	if req.PermCode != "" {
		if !perm.Valid(req.PermCode) {
			return errorx.New(400, "权限码无效")
		}
		res.PermCode = convert.StringPtr(req.PermCode)
	}
	if req.Path != "" {
		res.Path = req.Path
	}
	if req.ComponentPath != "" {
		res.ComponentPath = req.ComponentPath
	}
	if req.Redirect != "" {
		res.Redirect = req.Redirect
	}
	if req.RouteName != "" {
		res.RouteName = req.RouteName
	}
	if req.Icon != "" {
		res.Icon = req.Icon
	}
	if req.Sort != "" {
		v, err := convert.ParseSort(req.Sort, res.Sort)
		if err != nil {
			return err
		}
		res.Sort = v
	}
	if req.Status != "" {
		v, err := convert.ParseStatus(req.Status, res.Status)
		if err != nil {
			return err
		}
		res.Status = v
	}
	if req.Hidden != "" {
		v, err := convert.ParseFlag(req.Hidden, res.Hidden)
		if err != nil {
			return err
		}
		res.Hidden = v
	}
	if req.KeepAlive != "" {
		v, err := convert.ParseFlag(req.KeepAlive, res.KeepAlive)
		if err != nil {
			return err
		}
		res.KeepAlive = v
	}
	if req.Affix != "" {
		v, err := convert.ParseFlag(req.Affix, res.Affix)
		if err != nil {
			return err
		}
		res.Affix = v
	}
	if res.Type == "button" && convert.DerefString(res.PermCode) == "" {
		return errorx.New(400, "按钮类型必须填写权限码")
	}
	res.UpdatedBy = authctx.UserId(l.ctx)
	return l.svcCtx.Repo.Resource.Save(l.ctx, res)
}

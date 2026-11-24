package config

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ConfigDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigDetailLogic {
	return &ConfigDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ConfigDetailLogic) ConfigDetail(req *types.ConfigDetailReq) (resp *types.ConfigItem, err error) {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemConfigDetail); err != nil {
		return nil, err
	}
	// 2. 按 id 或 config_key 查询
	if req.Id == "" && req.ConfigKey == "" {
		return nil, errorx.New(400, "请提供 id 或 config_key")
	}
	var rowErr error
	if req.Id != "" {
		row, e := l.svcCtx.Repo.Config.FindById(l.ctx, req.Id)
		if e != nil {
			rowErr = e
		} else {
			return toConfigItem(row), nil
		}
	} else {
		row, e := l.svcCtx.Repo.Config.FindByKey(l.ctx, req.ConfigKey)
		if e != nil {
			rowErr = e
		} else {
			return toConfigItem(row), nil
		}
	}
	if rowErr == gorm.ErrRecordNotFound {
		return nil, errorx.New(404, "配置不存在")
	}
	return nil, rowErr
}

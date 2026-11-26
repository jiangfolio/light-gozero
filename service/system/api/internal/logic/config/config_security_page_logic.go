package config

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type SecurityConfigPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecurityConfigPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecurityConfigPageLogic {
	return &SecurityConfigPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// SecurityConfigPage 分页查询网络安全相关配置（auth/captcha）。
func (l *SecurityConfigPageLogic) SecurityConfigPage(req *types.SecurityConfigPageReq) (*types.ConfigPageResp, error) {
	// 1. 校验安全配置查看权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemSecurityConfigPage); err != nil {
		return nil, err
	}
	// 2. 解析分页与可选状态
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	var status int64
	hasStatus := false
	if req.Status != "" {
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	// 3. 仅查安全分组
	list, total, err := l.svcCtx.Repo.Config.PageByGroups(
		l.ctx, SecurityConfigGroups, req.ConfigKey, status, hasStatus,
		pageutil.Offset(page, pageSize), pageSize,
	)
	if err != nil {
		return nil, err
	}
	items := make([]types.ConfigItem, 0, len(list))
	for i := range list {
		items = append(items, *toConfigItem(&list[i]))
	}
	return &types.ConfigPageResp{
		Items:    items,
		Page:     convert.Int64ToString(int64(page)),
		PageSize: convert.Int64ToString(int64(pageSize)),
		Total:    convert.Int64ToString(total),
	}, nil
}

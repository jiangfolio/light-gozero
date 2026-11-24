package config

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ConfigPublicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigPublicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigPublicLogic {
	return &ConfigPublicLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ConfigPublic 无鉴权返回硬编码白名单配置，供前端读取。
func (l *ConfigPublicLogic) ConfigPublic() (resp *types.ConfigPublicResp, err error) {
	// 1. 仅按代码白名单取启用项，禁止猜 key 扩大暴露面
	list, err := l.svcCtx.AppConfig.ListByKeys(l.ctx, appcfg.PublicKeys)
	if err != nil {
		return nil, err
	}
	// 2. 组装公开字段（不含 id/remark/审计）
	items := make([]types.ConfigPublicItem, 0, len(list))
	for _, it := range list {
		items = append(items, types.ConfigPublicItem{
			ConfigKey:   it.ConfigKey,
			ConfigValue: it.ConfigValue,
			ValueType:   it.ValueType,
			GroupName:   it.GroupName,
		})
	}
	return &types.ConfigPublicResp{Items: items}, nil
}

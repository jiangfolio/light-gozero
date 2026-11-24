package config

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	cfgmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/config"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ConfigPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigPageLogic {
	return &ConfigPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ConfigPageLogic) ConfigPage(req *types.ConfigPageReq) (resp *types.ConfigPageResp, err error) {
	// 1. 校验操作权限
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemConfigPage); err != nil {
		return nil, err
	}
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
	list, total, err := l.svcCtx.Repo.Config.Page(l.ctx, req.GroupName, req.ConfigKey, status, hasStatus, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.ConfigItem, 0, len(list))
	for i := range list {
		items = append(items, *toConfigItem(&list[i]))
	}
	return &types.ConfigPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}

func toConfigItem(row *cfgmodel.SysConfig) *types.ConfigItem {
	return &types.ConfigItem{
		Id: row.Id, ConfigKey: row.ConfigKey, ConfigValue: row.ConfigValue, ValueType: row.ValueType,
		GroupName: row.GroupName, Remark: row.Remark, Status: convert.Int64ToString(row.Status),
		CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy,
		CreatedAt: convert.TimeToString(row.CreatedAt), UpdatedAt: convert.TimeToString(row.UpdatedAt),
	}
}
